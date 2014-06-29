package jiffy

import (
	"time"
)

type Message struct {
	Name     string
	Payload  interface{}
	uuid     string
	expireAt time.Time
}

func NewMessage(name string, payload interface{}, ttl time.Duration) *Message {
	return &Message{
		name,
		payload,
		UUID(),
		time.Now().Add(ttl),
	}
}

func (message *Message) Expired() bool {
	return time.Now().After(message.expireAt)
}
