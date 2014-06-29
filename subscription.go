package jiffy

import (
	"sync"
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
	Name        string
	Topic       *Topic
	Response    chan *Message
	expireChan  chan int
	expireMutex sync.Mutex
	uuid        string
}

func NewSubscription(name string, topic *Topic, ttl time.Duration) *Subscription {
	subscription := &Subscription{
		name,
		topic,
		make(chan *Message, ResponseBufferSize),
		make(chan int),
		sync.Mutex{},
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
func (subscription *Subscription) expire() {
	subscription.Topic.subscriptionMutex.Lock()
	defer subscription.Topic.subscriptionMutex.Unlock()

	if subscription.Active() {
		delete(subscription.Topic.Subscriptions, subscription.Name)
	}
}

func (subscription *Subscription) ExtendExpiration(ttl time.Duration) {
	select {
	case subscription.expireChan <- cancelTTL:
		// Do nothing if we can cancel.
	default:
		// Cycle mutex to ensure that expiration goroutine
		// has finished.
		subscription.expireMutex.Lock()
		subscription.expireMutex.Unlock()
	}

	subscription.Topic.subscriptionMutex.Lock()
	subscription.Topic.Subscriptions[subscription.Name] = subscription
	subscription.Topic.subscriptionMutex.Unlock()

	go subscription.QueueExpiration(ttl)
}

// Queues a subscription for deletion after the configured TTL.
func (subscription *Subscription) QueueExpiration(ttl time.Duration) {
	subscription.expireMutex.Lock()
	defer subscription.expireMutex.Unlock()

	ticker := time.NewTicker(ttl)
	select {
	case <-ticker.C:
		if subscription.Active() {
			subscription.expire()
		}
	case <-subscription.expireChan:
		return
	}
}

// Returns true if the subscription is active on a topic.
func (subscription *Subscription) Active() bool {
	if topicSubscription, ok := subscription.Topic.Subscriptions[subscription.Name]; ok {
		return topicSubscription.uuid == subscription.uuid
	}
	return false
}
