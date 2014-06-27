package jiffy

import (
	"time"
)

var (
	// The maximum number of messages to buffer in a subscription.
	ResponseChannelBufferSize = 100
	// How long a subscription lives for after a client stops using it.
	SubscriptionTTL = 60 * time.Second
)

type Subscription struct {
	Name            string
	Topic           *Topic
	ResponseChannel chan *Message
	ttlChan         chan int
}

func NewSubscription(name string, topic *Topic) *Subscription {
	return &Subscription{
		name,
		topic,
		make(chan *Message, ResponseChannelBufferSize),
		make(chan int),
	}
}

// Deletes the subscription its topic.
func (subscription *Subscription) Expire() {
	delete(subscription.Topic.Subscriptions, subscription.Name)
}

// Queues a subscription for deletion after the configured TTL.
func (subscription *Subscription) QueueExpiration() {
	go CallAfterTTL(subscription.Expire, SubscriptionTTL, subscription.ttlChan)
}

// Extends the subscription lifetime.
func (subscription *Subscription) ExtendExpiration() {
	subscription.ttlChan <- extendTTL
}

// Queues up all of the topic's data.
func (subscription *Subscription) FetchData() {
	for _, message := range subscription.Topic.Data {
		subscription.ResponseChannel <- message
	}
}
