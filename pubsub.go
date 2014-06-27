package main

import (
	"math/rand"
	"sync"
	"time"
)

const (
	MessageTTL = 60 * time.Second
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

func (topic *Topic) Publish(message *Message, ttl time.Duration) {
	for _, subscription := range topic.Subscriptions {
		go func() {
			subscription.ResponseChannel <- message
		}()
	}
	// Expire message after TTL.
	go func() {
		ticker := time.NewTicker(ttl * time.Second)
		<-ticker.C
		delete(topic.State, message.id)
	}()
}

func (topic *Topic) RecordAndPublish(message *Message) {
	topic.State[message.id] = message
	topic.Publish(message, MessageTTL)
}

type Subscription struct {
	Name            string
	Topic           *Topic
	ResponseChannel chan *Message
}

const (
	ResponseChannelBufferSize = 100
	SubscriptionTTL           = 60 * time.Second
)

// Creates a subscription on the topic if it doesn't exist already,
// and returns it.
func (topic *Topic) GetSubscription(name string) *Subscription {
	topic.SubscriberMutex.Lock()
	defer topic.SubscriberMutex.Unlock()

	// Expire subscription after TTL.
	defer func() {
		ticker := time.NewTicker(SubscriptionTTL)
		<-ticker.C
		delete(topic.Subscriptions, name)
	}()

	if subscription, ok := topic.Subscriptions[name]; ok {
		return subscription
	}
	subscription := &Subscription{
		name,
		topic,
		make(chan *Message, ResponseChannelBufferSize),
	}
	topic.Subscriptions[name] = subscription

	// Fetch the current state for the topic if this is
	// first session.
	go subscription.FetchState()

	return subscription
}

func (subscription *Subscription) FetchState() {
	for _, message := range subscription.Topic.State {
		subscription.ResponseChannel <- message
	}
}

var (
	Rand = rand.Rand{}
)

type Message struct {
	id      string
	Payload string `json:"payload"`
}

func NewMessage(id, payload string) *Message {
	return &Message{id, payload}
}

type Registry struct {
	Topics        map[string]*Topic
	NewTopicMutex sync.Mutex
}

func CreateRegistry() *Registry {
	return &Registry{
		make(map[string]*Topic),
		sync.Mutex{},
	}
}

// Creates a topic on the registry if it doesn't exist already,
// and returns it.
func (registry *Registry) GetTopic(name string) *Topic {
	registry.NewTopicMutex.Lock()
	defer registry.NewTopicMutex.Unlock()
	if topic, ok := registry.Topics[name]; ok {
		return topic
	}
	registry.Topics[name] = CreateTopic(name)
	return registry.Topics[name]
}
