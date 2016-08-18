package main

import "log"

func init() {
	if err := registerMessageHandler("io.luzifer.outputs.log", MessageHandlerLog{}); err != nil {
		log.Fatalf("Unable to register io.luzifer.outputs.log handler")
	}
}

type MessageHandlerLog struct{}

func (h MessageHandlerLog) Handle(m message) (stopHandling bool, err error) {
	log.Printf("%#v", m)

	return true, nil
}
