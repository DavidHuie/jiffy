package jiffy

import (
	"errors"
	"time"
)

var (
	// The maximum number of messages to buffer in a subscription.
	ResponseBufferSize = 100

	// The wait before timing a publish.
	PublishTimeout = 10 * time.Minute
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
	ticker := time.NewTicker(PublishTimeout)
	select {
	case subscription.Response <- message:
	case <-ticker.C:
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

var (
	ExpiredSubscription = errors.New("Subscription is expired")
	TopicExpired        = errors.New("Topic for subscription has expired")
)

// Extends the subscription's expiration by the input TTL.
func (subscription *Subscription) ExtendExpiration(ttl time.Duration) error {
	if !subscription.Active() {
		if !subscription.Topic.Active() {
			return TopicExpired
		}
		return ExpiredSubscription
	}
	subscription.expireAt = time.Now().Add(ttl)
	return nil
}

// Returns true if the subscription is active on a topic.
func (subscription *Subscription) Active() bool {
	if subscription.Expired() || !subscription.Topic.Active() {
		return false
	}
	if topicSubscription, ok := subscription.Topic.Subscriptions[subscription.Name]; ok {
		return topicSubscription.uuid == subscription.uuid
	}
	return false
}
