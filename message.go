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
	uuid    string
}

func NewMessage(id, payload string) *Message {
	return &Message{
		id,
		payload,
		make(chan int),
		UUID(),
	}
}

// Queues a message for deletion from it's topic after
// the configured TTL.
func (message *Message) QueueExpiration(topic *Topic, ttl time.Duration) {
	ticker := time.NewTicker(ttl)
	<-ticker.C
	topic.MessageMutex.Lock()
	defer topic.MessageMutex.Unlock()
	if topicMessage, ok := topic.Data[message.id]; ok {
		// Only delete the correct message.
		if topicMessage.uuid == message.uuid {
			delete(topic.Data, message.id)
		}
	}
}

// Extends a message's TTL.
func (message *Message) ExtendExpiration() {
	message.ttlChan <- extendTTL
}

// Cancels a message's scheduled deletion.
func (message *Message) CancelExpiration() {
	message.ttlChan <- cancelTTL
}
