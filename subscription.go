package jiffy

import (
	"time"
)

var (
	// The maximum number of messages to buffer in a subscription.
	ResponseBufferSize = 100
)

type Subscription struct {
	Name     string
	Topic    *Topic
	Response chan *Message
	uuid     string
	expireAt time.Time
}

func NewSubscription(name string, topic *Topic, ttl time.Duration) *Subscription {
	subscription := &Subscription{
		name,
		topic,
		make(chan *Message, ResponseBufferSize),
		UUID(),
		time.Now().Add(ttl),
	}
	return subscription
}

// Publishes a message to the subscription.
func (subscription *Subscription) Publish(message *Message) {
	subscription.Response <- message
}

// Deletes the subscription from its topic.
func (subscription *Subscription) Expire() {
	subscription.expireAt = time.Now()
}

// Expires the subscription.
func (subscription *Subscription) Expired() bool {
	return time.Now().After(subscription.expireAt)
}

// Extends the subscription's expiration by the input TTL.
func (subscription *Subscription) ExtendExpiration(ttl time.Duration) {
	subscription.Activate()
	subscription.expireAt = time.Now().Add(ttl)
}

// Returns true if the subscription is active on a topic.
func (subscription *Subscription) Active() bool {
	if topicSubscription, ok := subscription.Topic.Subscriptions[subscription.Name]; ok {
		return (topicSubscription.uuid == subscription.uuid) && !subscription.Expired()
	}
	return false
}

// Resubscribes a subscription to its configured topic.
func (subscription *Subscription) Activate() {
	subscription.Topic.Subscriptions[subscription.Name] = subscription
}
