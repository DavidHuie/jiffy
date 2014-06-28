package jiffy

import (
	"time"
)

var (
	// The maximum number of messages to buffer in a subscription.
	ResponseChannelBufferSize = 100
)

type Subscription struct {
	Name            string
	Topic           *Topic
	ResponseChannel chan *Message
	expireChan      chan int
	uuid            string
}

func NewSubscription(name string, topic *Topic, ttl time.Duration) *Subscription {
	subscription := &Subscription{
		name,
		topic,
		make(chan *Message, ResponseChannelBufferSize),
		make(chan int),
		UUID(),
	}
	go subscription.QueueExpiration(ttl)
	return subscription
}

// Publishes a message to the subscription.
func (subscription *Subscription) Publish(message *Message) {
	subscription.ResponseChannel <- message
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

// Returns all of the topic's data.
func (subscription *Subscription) FetchData() []*Message {
	messages := make([]*Message, 0, len(subscription.Topic.Data))
	for _, message := range subscription.Topic.Data {
		messages = append(messages, message)
	}
	return messages
}

// Returns true if the subscription is active on a topic.
func (subscription *Subscription) Active() bool {
	if topicSubscription, ok := subscription.Topic.Data[subscription.Name]; ok {
		return topicSubscription.uuid == subscription.uuid
	}
	return false
}
