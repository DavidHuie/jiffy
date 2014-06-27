package jiffy

import (
	"time"
)

var (
	SubscriptionTTL = 60 * time.Second
)

type Subscription struct {
	Name            string
	Topic           *Topic
	ResponseChannel chan *Message
	ttlChan         chan string
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
