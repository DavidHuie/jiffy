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
	Messages      map[string]*Message
	registry      *Registry
	uuid          string
	// Mutex for creating and destroying subscriptions.
	subscriptionMutex sync.Mutex
	// Mutex for creating and destroying messages.
	messageMutex sync.Mutex
}

func NewTopic(name string, registry *Registry) *Topic {
	return &Topic{
		name,
		make(map[string]*Subscription),
		make(map[string]*Message),
		registry,
		UUID(),
		sync.Mutex{},
		sync.Mutex{},
	}
}

// Publishes a message to all subscribers.
func (topic *Topic) Publish(message *Message) {
	for _, subscription := range topic.Subscriptions {
		if !subscription.Active() {
			continue
		}
		go func(s *Subscription) {
			s.Publish(message)
		}(subscription)
	}
}

// Records a message to the topic's cache.
func (topic *Topic) Record(message *Message) {
	topic.Messages[message.Name] = message
}

// Returns all of the topic's cached data.
func (topic *Topic) CachedMessages() []*Message {
	messages := make([]*Message, 0, len(topic.Messages))
	for _, message := range topic.Messages {
		topic.messageMutex.Lock()
		if !message.Expired() {
			messages = append(messages, message)
		}
		topic.messageMutex.Unlock()
	}
	return messages
}

// Destroys expired cached messages.
func (topic *Topic) CleanExpiredCachedMessages() {
	for _, message := range topic.Messages {
		topic.messageMutex.Lock()
		if message.Expired() {
			delete(topic.Messages, message.Name)
		}
		topic.messageMutex.Unlock()
	}
}

// Destroys expired subscriptions.
func (topic *Topic) CleanExpiredSubscriptions() {
	for _, subscription := range topic.Subscriptions {
		topic.subscriptionMutex.Lock()
		if subscription.Expired() {
			delete(topic.Subscriptions, subscription.Name)
		}
		topic.subscriptionMutex.Unlock()
	}
}

// Publishes a message and stores it in the topic.
func (topic *Topic) RecordAndPublish(message *Message) {
	topic.Record(message)
	go topic.Publish(message)
}

// Creates a subscription on the topic if it doesn't exist
// and returns it.
func (topic *Topic) GetSubscription(name string, ttl time.Duration) *Subscription {
	topic.subscriptionMutex.Lock()
	defer topic.subscriptionMutex.Unlock()

	if subscription, ok := topic.Subscriptions[name]; ok {
		subscription.ExtendExpiration(ttl)
		return subscription
	}

	subscription := NewSubscription(name, topic, ttl)
	topic.Subscriptions[name] = subscription
	return subscription
}
