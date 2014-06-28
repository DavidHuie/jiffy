package jiffy

import (
	"time"
)

type Message struct {
	Name    string `json:"message"`
	Payload string `json:"payload"`
	ttlChan chan int
	uuid    string
}

func NewMessage(name, payload string) *Message {
	return &Message{
		name,
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
	topic.messageMutex.Lock()
	defer topic.messageMutex.Unlock()
	if topicMessage, ok := topic.Data[message.Name]; ok {
		if topicMessage.uuid == message.uuid {
			delete(topic.Data, message.Name)
		}
	}
}
