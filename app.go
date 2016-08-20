package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Luzifer/processing-hub/scripthelpers"
	"github.com/Luzifer/rconfig"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

var (
	cfg = struct {
		ConfigBackend  string `flag:"config" default:"env" description:"Config backend to use"`
		Listen         string `flag:"listen" default:":3000" default:"IP/Port to listen on"`
		SQSQueueURL    string `flag:"sqs-queue-url" env:"SQS_QUEUE_URL" description:"URL of the SQS queue to use for messages"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	inputHandlers   = map[string]inputHandler{}
	messageHandlers = map[string]messageHandler{}
	config          configStore

	sqsClient *sqs.SQS

	version = "dev"
)

type configStore interface {
	Get(path string) (interface{}, error)
	GetString(path string) (string, error)
	GetInt(path string) (int64, error)
}

type inputHandler interface {
	RegisterOnRouter(*mux.Router)
	GetHelp() inputHandlerHelp
}

type inputHandlerHelp struct {
	Path        string
	Description string
}

type messageHandler interface {
	Handle(message) (stopHandling bool, err error)
}

func registerInputHandler(pathPrefix string, handler inputHandler) error {
	if _, ok := inputHandlers[pathPrefix]; ok {
		return errors.New("Duplicate handler for prefix " + pathPrefix)
	}
	inputHandlers[pathPrefix] = handler
	return nil
}

func registerMessageHandler(messageType string, handler messageHandler) error {
	if _, ok := messageHandlers[messageType]; ok {
		return errors.New("Duplicate handler for message type " + messageType)
	}
	messageHandlers[messageType] = handler
	return nil
}

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("processing-hub %s\n", version)
		os.Exit(0)
	}
}

func main() {
	config = getConfigStore()

	r := mux.NewRouter()

	for path, handler := range inputHandlers {
		handler.RegisterOnRouter(r.PathPrefix("/" + path + "/").Subrouter())
	}

	go http.ListenAndServe(cfg.Listen, r)

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("Unable to create SQS session: %s", err)
	}
	sqsClient = sqs.New(sess)

	rcvConfig := sqs.ReceiveMessageInput{
		MaxNumberOfMessages: aws.Int64(1),
		QueueUrl:            aws.String(cfg.SQSQueueURL),
		VisibilityTimeout:   aws.Int64(30),
		WaitTimeSeconds:     aws.Int64(20),
	}
	for {
		msgs, err := sqsClient.ReceiveMessage(&rcvConfig)
		if err != nil {
			log.Fatalf("Unable to receive messages: %s", err)
		}

		for _, msg := range msgs.Messages {
			go handleSQSMessage(msg)
		}
	}
}

func handleSQSMessage(msg *sqs.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		run := true
		for run {
			select {
			case <-time.Tick(25 * time.Second):
				sqsClient.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
					VisibilityTimeout: aws.Int64(30),
					ReceiptHandle:     msg.ReceiptHandle,
				})
			case <-ctx.Done():
				run = false
			}
		}
	}()

	m, err := DecodeMessage(msg.Body)
	if err != nil {
		log.Printf("[ERR] Wasn't able to decode message %s: %s", *msg.MessageId, err)
		return
	}

	if handler, ok := messageHandlers[m.Type()]; ok {
		stopHandling, err := handler.Handle(m)
		if err != nil {
			log.Printf("[ERR] Wasn't able to process message %s: %s", *msg.MessageId, err)
			return
		}

		defer func() {
			if err := deleteMessage(msg.ReceiptHandle); err != nil {
				log.Printf("[ERR] Unable to delete message %s: %s", *msg.MessageId, err)
			}
		}()
		if stopHandling {
			return
		}
	}

	//TODO: Run through JS handlers
}

func deleteMessage(receiptHandle *string) error {
	_, err := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(cfg.SQSQueueURL),
		ReceiptHandle: receiptHandle,
	})

	return err
}

func enqueueMessage(m *message) error {
	if err := m.Validate(); err != nil {
		return err
	}

	body, err := m.Encode()
	if err != nil {
		return err
	}

	_, err = sqsClient.SendMessage(&sqs.SendMessageInput{
		MessageBody: body,
		QueueUrl:    aws.String(cfg.SQSQueueURL),
	})

	return err
}

func executeExecutorScript() {
	vm := otto.New()
	vm.Set("$", scripthelpers.GetHelperMap())

	code := `
	var bar = function(data) {
		var d = $.request('http://requestb.in/1m5p8ca1', {
			"data": JSON.stringify(data),
			"method": "POST",
			"contentType": "application/json",
		});
		console.log("Body: " + d.Body + "\nError: " + d.Error);
		return d.StatusCode;
	}

	var foo = function() {
		console.log("foobar");
	}

	exported = {
		"post_data": bar,
		"console": foo,
	}
	`

	rt, err := vm.Object(code)
	if err != nil {
		log.Printf("error: %s", err.Error())
	}

	v, err := rt.Call("post_data", map[string]string{"bar": "foo", "blubb": "bla"})
	log.Printf("Status: %#v, Err: %s", v, err)
}
