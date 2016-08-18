package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func init() {
	if err := registerMessageHandler("io.luzifer.outputs.pushover", MessageHandlerPushover{}); err != nil {
		log.Fatalf("Unable to register io.luzifer.outputs.pushover handler")
	}
}

type MessageHandlerPushover struct{}

func (h MessageHandlerPushover) Handle(m message) (stopHandling bool, err error) {
	u := "https://api.pushover.net/1/messages.json"

	if _, ok := m["message"]; !ok {
		return false, errors.New("Message needs to contain 'message' attribute for io.luzifer.outputs.pushover")
	}

	if _, ok := m["user"]; !ok {
		return false, errors.New("Message needs to contain 'user' attribute for io.luzifer.outputs.pushover")
	}

	vals := url.Values{
		"timestamp": []string{strconv.FormatInt(m.Date().Unix(), 10)},
	}

	for _, i := range []string{"user", "message", "title", "url", "url_title", "priority"} {
		if d, ok := m[i]; ok {
			vals.Set(i, d)
		}
	}

	res, err := http.PostForm(u, vals)
	if err != nil {
		return false, errors.New("Unable to execute HTTP request: " + err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Received unexpected status code: %d", res.StatusCode)
	}

	return true, nil
}
