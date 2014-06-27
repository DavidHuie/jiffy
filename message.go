package jiffy

import (
	"time"
)

var (
	MessageTTL = 60 * time.Second
)

type Message struct {
	id      string
	Payload string `json:"payload"`
	ttlChan chan string
}

func NewMessage(id, payload string) *Message {
	return &Message{id, payload, make(chan string)}
}

func (message *Message) QueueExpiration(topic *Topic) {
	expirationFunction := func() {
		delete(topic.State, message.id)
	}
	go CallAfterTTL(expirationFunction, MessageTTL, message.ttlChan)
}

func (message *Message) ExtendExpiration() {
	message.ttlChan <- "extend"
}

func (message *Message) CancelExpiration() {
	message.ttlChan <- "cancel"
}
