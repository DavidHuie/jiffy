package jiffy

import (
	"sync"
)

// A topic coordinates the publishing and distribution
// of messages.
type Topic struct {
	Name            string
	Subscriptions   map[string]*Subscription
	SubscriberMutex sync.Mutex
	Data            map[string]*Message
}

func CreateTopic(name string) *Topic {
	return &Topic{
		name,
		make(map[string]*Subscription),
		sync.Mutex{},
		make(map[string]*Message),
	}
}

// Publishes a message to all subscribers.
func (topic *Topic) Publish(message *Message) {
	for _, subscription := range topic.Subscriptions {
		go func() {
			subscription.ResponseChannel <- message
		}()
	}
}

// Publishes a message and stores it in the topic.
func (topic *Topic) RecordAndPublish(message *Message) {
	// If we're replacing a message with the same id,
	// we want to cancel it's expiration so that it doesn't
	// interfere with the new one.
	if previousMessage, ok := topic.Data[message.id]; ok {
		previousMessage.CancelExpiration()
	}
	topic.Data[message.id] = message
	topic.Publish(message)

	// Always queue messages up for deletion.
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

	subscription := NewSubscription(name, topic)
	topic.Subscriptions[name] = subscription

	// Expire subscription after a TTL.
	subscription.QueueExpiration()

	// Fetch the current state for the topic if this is the
	// first session.
	go subscription.FetchData()

	return subscription
}
