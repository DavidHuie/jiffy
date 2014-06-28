package jiffy

import (
	"time"
)

const (
	cancelTTL = iota
	extendTTL
)

var (
	// The maximum number of messages to buffer in a subscription.
	ResponseBufferSize = 100
)

type Subscription struct {
	Name       string
	Topic      *Topic
	Response   chan *Message
	expireChan chan int
	uuid       string
}

func NewSubscription(name string, topic *Topic, ttl time.Duration) *Subscription {
	subscription := &Subscription{
		name,
		topic,
		make(chan *Message, ResponseBufferSize),
		make(chan int),
		UUID(),
	}
	go subscription.QueueExpiration(ttl)
	return subscription
}

// Publishes a message to the subscription.
func (subscription *Subscription) Publish(message *Message) {
	subscription.Response <- message
}

// Deletes the subscription from its topic.
func (subscription *Subscription) Expire() {
	subscription.Topic.subscriptionMutex.Lock()
	defer subscription.Topic.subscriptionMutex.Unlock()
	if subscription.Active() {
		delete(subscription.Topic.Data, subscription.Name)
	}
}

// Queues a subscription for deletion after the configured TTL.
func (subscription *Subscription) QueueExpiration(ttl time.Duration) {
	ticker := time.NewTicker(ttl)
	select {
	case <-ticker.C:
		subscription.Expire()
	case <-subscription.expireChan:
		return
	}
}

// Returns true if the subscription is active on a topic.
func (subscription *Subscription) Active() bool {
	if topicSubscription, ok := subscription.Topic.Data[subscription.Name]; ok {
		return topicSubscription.uuid == subscription.uuid
	}
	return false
}
