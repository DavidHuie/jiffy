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
	uuid          string
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
		UUID(),
		sync.Mutex{},
		sync.Mutex{},
	}
}

// Publishes a message to all subscribers.
func (topic *Topic) Publish(message *Message) {
	for _, subscription := range topic.Subscriptions {
		go func(s *Subscription) {
			s.Publish(message)
		}(subscription)
	}
}

// Records a message to the topic's cache.
func (topic *Topic) Record(message *Message, ttl time.Duration) {
	topic.messageMutex.Lock()
	defer topic.messageMutex.Unlock()
	topic.Data[message.Name] = message
	go message.QueueExpiration(topic, ttl)
}

// Returns all of the topic's cached data.
func (topic *Topic) FetchData() []*Message {
	messages := make([]*Message, 0, len(topic.Data))
	for _, message := range topic.Data {
		messages = append(messages, message)
	}
	return messages
}

// Publishes a message and stores it in the topic.
func (topic *Topic) RecordAndPublish(message *Message, ttl time.Duration) {
	topic.Record(message, ttl)
	go topic.Publish(message)
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
			go subscription.QueueExpiration(ttl)
			return subscription
		default:
			// We're too late
		}
	}

	subscription := NewSubscription(name, topic, ttl)
	topic.Subscriptions[name] = subscription
	return subscription
}
