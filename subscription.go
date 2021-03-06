package jiffy

import (
	"errors"
	"time"
)

var (
	// The maximum number of messages to buffer in a subscription.
	ResponseBufferSize = 100
	// The wait before timing out a publish.
	PublishTimeout      = 10 * time.Minute
	ExpiredSubscription = errors.New("Subscription is expired")
	PublishTimedOut     = errors.New("Publish timed out")
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
func (subscription *Subscription) Publish(message *Message) error {
	ticker := time.NewTicker(PublishTimeout)
	select {
	case subscription.Response <- message:
		return nil
	case <-ticker.C:
		return PublishTimedOut
	}
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
func (subscription *Subscription) ExtendExpiration(ttl time.Duration) error {
	if !subscription.Active() {
		return ExpiredSubscription
	}
	subscription.expireAt = time.Now().Add(ttl)
	return nil
}

// Returns true if the subscription is active on a topic.
func (subscription *Subscription) Active() bool {
	if subscription.Expired() {
		return false
	}
	if topicSubscription, ok := subscription.Topic.Subscriptions[subscription.Name]; ok {
		return topicSubscription.uuid == subscription.uuid
	}
	return false
}
