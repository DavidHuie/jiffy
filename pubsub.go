package main

import (
	"math/rand"
	"sync"
	"time"
)

const (
	MessageTTL                = 60 * time.Second
	SubscriptionTTL           = 60 * time.Second
	ResponseChannelBufferSize = 100
)

var (
	Rand = rand.Rand{}
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

type Subscription struct {
	Name            string
	Topic           *Topic
	ResponseChannel chan *Message
	ttlChan         chan string
}

// Creates a subscription on the topic if it doesn't exist already,
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

func (subscription *Subscription) Expire() {
	delete(subscription.Topic.Subscriptions, subscription.Name)
}

func (subscription *Subscription) QueueExpiration() {
	go CallAfterTTL(subscription.Expire, SubscriptionTTL, subscription.ttlChan)
}

func (subscription *Subscription) ExtendExpiration() {
	subscription.ttlChan <- "extend"
}

func (subscription *Subscription) FetchState() {
	for _, message := range subscription.Topic.State {
		subscription.ResponseChannel <- message
	}
}

type Message struct {
	id      string
	Payload string `json:"payload"`
	ttlChan chan string
}

func NewMessage(id, payload string) *Message {
	return &Message{id, payload, make(chan string)}
}

func (message *Message) QueueExpiration(topic *Topic) {
	expirationFunction := func() {
		delete(topic.State, message.id)
	}
	go CallAfterTTL(expirationFunction, MessageTTL, message.ttlChan)
}

func (message *Message) ExtendExpiration() {
	message.ttlChan <- "extend"
}

func (message *Message) CancelExpiration() {
	message.ttlChan <- "cancel"
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
