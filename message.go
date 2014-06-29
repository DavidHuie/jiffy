package jiffy

import (
	"time"
)

type Message struct {
	Name    string
	Payload interface{}
	uuid    string
}

func NewMessage(name string, payload interface{}) *Message {
	return &Message{
		name,
		payload,
		UUID(),
	}
}

// Deletes a message from it's topic after a certain amount of time.
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
