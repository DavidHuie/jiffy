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
	ttlChan chan int
}

func NewMessage(id, payload string) *Message {
	return &Message{id, payload, make(chan int)}
}

// Queues a message for deletion from it's topic after
// the configured TTL.
func (message *Message) QueueExpiration(topic *Topic) {
	expirationFunction := func() {
		delete(topic.Data, message.id)
	}
	go CallAfterTTL(expirationFunction, MessageTTL, message.ttlChan)
}

// Extends a message's TTL.
func (message *Message) ExtendExpiration() {
	message.ttlChan <- extendTTL
}

// Cancels a message's scheduled deletion.
func (message *Message) CancelExpiration() {
	message.ttlChan <- cancelTTL
}
