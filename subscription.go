package jiffy

import (
	"sync"
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
	expireMutex     sync.Mutex
	expireChan      chan int
	uuid            string
	ttl             time.Duration
}

func NewSubscription(name string, topic *Topic, ttl time.Duration) *Subscription {
	subscription := &Subscription{
		name,
		topic,
		make(chan *Message, ResponseChannelBufferSize),
		sync.Mutex{},
		make(chan int),
		UUID(),
		ttl,
	}
	go subscription.QueueExpiration()
	return subscription
}

// Deletes the subscription from its topic.
func (subscription *Subscription) Expire() {
	subscription.Topic.SubscriberMutex.Lock()
	defer subscription.Topic.SubscriberMutex.Unlock()
	if topicSubscription, ok := subscription.Topic.Subscriptions[subscription.Name]; ok {
		// Only delete the subscription if it's the right one.
		if topicSubscription.uuid == subscription.uuid {
			delete(subscription.Topic.Subscriptions, subscription.Name)
		}
	}
}

// Queues a subscription for deletion after the configured TTL.
func (subscription *Subscription) QueueExpiration() {
	ticker := time.NewTicker(subscription.ttl)
	select {
	case <-ticker.C:
		subscription.Expire()
	case <-subscription.expireChan:
		return
	}
}

// Cancels a queued expiration.
func (subscription *Subscription) CancelExpiration() {
	subscription.expireChan <- cancelTTL
}

// Queues up all of the topic's data.
func (subscription *Subscription) FetchData() {
	for _, message := range subscription.Topic.Data {
		subscription.ResponseChannel <- message
	}
}

// Returns true if the subscription is active on a topic.
func (subscription *Subscription) Active() bool {
	if _, ok := subscription.Topic.Data[subscription.Name]; ok {
		return true
	}
	return false
}
