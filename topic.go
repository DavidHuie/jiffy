package jiffy

import (
	"sync"
)

var (
	ResponseChannelBufferSize = 100
)

type Topic struct {
	Name            string
	Subscriptions   map[string]*Subscription
	SubscriberMutex sync.Mutex
	State           map[string]*Message
}

func CreateTopic(name string) *Topic {
	return &Topic{
		name,
		make(map[string]*Subscription),
		sync.Mutex{},
		make(map[string]*Message),
	}
}

func (topic *Topic) Publish(message *Message) {
	for _, subscription := range topic.Subscriptions {
		go func() {
			subscription.ResponseChannel <- message
		}()
	}
}

func (topic *Topic) RecordAndPublish(message *Message) {
	if previousMessage, ok := topic.State[message.id]; ok {
		previousMessage.CancelExpiration()
	}
	topic.State[message.id] = message
	topic.Publish(message)
	message.QueueExpiration(topic)
}

// Creates a subscription on the topic if it doesn't exist
// and returns it.
func (topic *Topic) GetSubscription(name string) *Subscription {
	topic.SubscriberMutex.Lock()
	defer topic.SubscriberMutex.Unlock()

	if subscription, ok := topic.Subscriptions[name]; ok {
		subscription.ExtendExpiration()
		return subscription
	}
	subscription := &Subscription{
		name,
		topic,
		make(chan *Message, ResponseChannelBufferSize),
		make(chan string),
	}
	topic.Subscriptions[name] = subscription

	// Expire subscription after a TTL.
	subscription.QueueExpiration()

	// Fetch the current state for the topic if this is
	// first session.
	go subscription.FetchState()

	return subscription
}
