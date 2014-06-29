package jiffy

import (
	"sync"
	"time"
)

var (
	// The maximum number of messages to buffer in a subscription.
	ResponseBufferSize = 100
)

type Subscription struct {
	Name        string
	Topic       *Topic
	Response    chan *Message
	expireChan  chan time.Duration
	expireMutex sync.Mutex
	uuid        string
}

func NewSubscription(name string, topic *Topic, ttl time.Duration) *Subscription {
	subscription := &Subscription{
		name,
		topic,
		make(chan *Message, ResponseBufferSize),
		make(chan time.Duration),
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
	subscription.expireChan <- ttl
}

// Queues a subscription for deletion after the configured TTL. The TTL can
// be extended by sending a new one on the expireChan.
func (subscription *Subscription) QueueExpiration(ttl time.Duration) {
	subscription.expireMutex.Lock()
	defer subscription.expireMutex.Unlock()

	ticker := time.NewTicker(ttl)

	for {
		select {
		case <-ticker.C:
			if subscription.Active() {
				subscription.expire()
			}
		case newTtl := <-subscription.expireChan:
			subscription.Activate()
			ticker = time.NewTicker(newTtl)
		}
		// TODO: kill this goroutine when this subscription dies.
	}

}

// Returns true if the subscription is active on a topic.
func (subscription *Subscription) Active() bool {
	if topicSubscription, ok := subscription.Topic.Subscriptions[subscription.Name]; ok {
		return topicSubscription.uuid == subscription.uuid
	}
	return false
}

// Resubscribes a subscription to its configured topic.
func (subscription *Subscription) Activate() {
	subscription.Topic.subscriptionMutex.Lock()
	subscription.Topic.Subscriptions[subscription.Name] = subscription
	subscription.Topic.subscriptionMutex.Unlock()
}
