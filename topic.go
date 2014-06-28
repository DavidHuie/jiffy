package jiffy

import (
	"sync"
	"time"
)

// A topic coordinates the publishing and distribution
// of messages.
type Topic struct {
	Name          string
	Subscriptions map[string]*Subscription
	Data          map[string]*Message
	// Mutex for creating and destroying subscriptions.
	subscriptionMutex sync.Mutex
	// Mutex for creating and destroying messages.
	messageMutex sync.Mutex
}

func CreateTopic(name string) *Topic {
	return &Topic{
		name,
		make(map[string]*Subscription),
		make(map[string]*Message),
		sync.Mutex{},
		sync.Mutex{},
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
func (topic *Topic) RecordAndPublish(message *Message, ttl time.Duration) {
	topic.messageMutex.Lock()
	defer topic.messageMutex.Unlock()
	topic.Data[message.Name] = message
	go topic.Publish(message)
	go message.QueueExpiration(topic, ttl)
}

// Creates a subscription on the topic if it doesn't exist
// and returns it.
func (topic *Topic) GetSubscription(name string, ttl time.Duration) *Subscription {
	topic.subscriptionMutex.Lock()
	defer topic.subscriptionMutex.Unlock()

	if subscription, ok := topic.Subscriptions[name]; ok {
		select {
		case subscription.expireChan <- cancelTTL:
			// If we were able to cancel successfully,
			// just restart the expiration.
			subscription.ttl = ttl
			subscription.QueueExpiration(ttl)
			return subscription
		default:
			// We're too late
			break
		}
	}

	subscription := NewSubscription(name, topic, ttl)
	topic.Subscriptions[name] = subscription

	// Fetch the current state for the topic if this is the
	// first session.
	go subscription.FetchData()

	return subscription
}
