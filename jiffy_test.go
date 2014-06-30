package jiffy

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestGetTopic(t *testing.T) {
	topic := GetTopic("test_topic1")
	if topic.Name != "test_topic1" {
		t.Errorf("Invalid name")
	}
	if len(topic.Subscriptions) != 0 {
		t.Errorf("Invalid number of subscriptions")
	}
	if len(topic.Messages) != 0 {
		t.Errorf("Invalid data count")
	}
	// Getting the topic again should return the exact
	// same topic.
	topic2 := GetTopic("test_topic1")
	if topic.uuid != topic2.uuid {
		t.Errorf("Topics should be equal")
	}
}

func TestTopicCachedMessages(t *testing.T) {
	topic := GetTopic("test_topic2")
	message1 := NewMessage("test-name1", "my message1", 20*time.Millisecond)
	message2 := NewMessage("test-name2", "my message2", 20*time.Millisecond)
	topic.Record(message1)
	topic.Record(message2)

	messages := topic.CachedMessages()
	if len(messages) != 2 {
		t.Errorf("Invalid number of messages returned")
	}

	messageNames := []string{messages[0].Name, messages[1].Name}
	sort.Strings(messageNames)

	if !reflect.DeepEqual(messageNames, []string{"test-name1", "test-name2"}) {
		t.Errorf("Invalid messages returned, %v, %v", messages[0], messages[1])
	}

	time.Sleep(20 * time.Millisecond)
	messages = topic.CachedMessages()
	if numMsgs := len(messages); numMsgs != 0 {
		t.Errorf("Invalid number of messages returned: %v", numMsgs)
	}
}

func TestSubscriptionResponse(t *testing.T) {
	topic := GetTopic("test_topic3")
	message := NewMessage("test-message", "my message", time.Minute)
	topic.GetSubscription("sub1", time.Minute)
	sub1 := topic.GetSubscription("sub1", time.Minute)
	sub2 := topic.GetSubscription("sub2", time.Minute)

	topic.Publish(message)

	msg1 := <-sub1.Response
	msg2 := <-sub2.Response

	if (msg1 != message) || (msg2 != message) {
		t.Errorf("Invalid messages returned to subscribers")
	}
}

func TestTopicRecordAndPublish(t *testing.T) {
	topic := GetTopic("test_topic4")
	sub1 := topic.GetSubscription("sub1", time.Minute)
	sub2 := topic.GetSubscription("sub2", time.Minute)
	message := NewMessage("test-name1", "my message1", time.Minute)
	topic.RecordAndPublish(message)

	messages := topic.CachedMessages()
	if len(messages) != 1 {
		t.Errorf("Invalid number of messages returned: %v", len(messages))
	}
	if messages[0] != message {
		t.Errorf("Invalid message returned")
	}

	msg1 := <-sub1.Response
	msg2 := <-sub2.Response

	if (msg1 != message) || (msg2 != message) {
		t.Errorf("Invalid messages returned to subscribers")
	}
}

func TestSubExpire(t *testing.T) {
	topic := GetTopic("test_topic5")
	sub := topic.GetSubscription("sub", 10*time.Millisecond)
	time.Sleep(11 * time.Millisecond)
	if sub.Active() != false {
		t.Errorf("Subscription should be inactive")
	}
}

func TestSubExtendExpiration(t *testing.T) {
	topic := GetTopic("test_topic6")
	sub := topic.GetSubscription("sub", 10*time.Millisecond)

	if status := sub.Active(); status != true {
		t.Errorf("Subscription should be active, got %v", status)
	}

	if err := sub.ExtendExpiration(20 * time.Millisecond); err != nil {
		t.Errorf("Extending expiration should not throw error")
	}
	time.Sleep(10 * time.Millisecond)

	if status := sub.Active(); status != true {
		t.Errorf("Subscription should be active, got %v", status)
	}

	time.Sleep(10 * time.Millisecond)

	if status := sub.Active(); status != false {
		t.Errorf("Subscription should be inactive, got %v", status)
	}

	if err := sub.ExtendExpiration(20 * time.Millisecond); err != ExpiredSubscription {
		t.Errorf("Extending expiration should throw error")
	}
}

func TestSubClean(t *testing.T) {
	topic := GetTopic("test_topic7")
	topic.GetSubscription("sub", 10*time.Millisecond)

	if len(topic.Subscriptions) != 1 {
		t.Errorf("Subscription should still be around")
	}

	time.Sleep(10 * time.Millisecond)
	topic.CleanExpiredSubscriptions()

	if len(topic.Subscriptions) != 0 {
		t.Errorf("Subscription should be deleted")
	}
}

func TestMessageClean(t *testing.T) {
	topic := GetTopic("test_topic8")
	message := NewMessage("test-name1", "my message1", 10*time.Millisecond)
	topic.RecordAndPublish(message)

	messages := topic.CachedMessages()
	if len(messages) != 1 || messages[0] != message {
		t.Errorf("Message should be present")
	}

	time.Sleep(10 * time.Millisecond)
	topic.CleanExpiredCachedMessages()

	messages = topic.CachedMessages()
	if len(messages) != 0 {
		t.Errorf("Message should be expired")
	}
}

func BenchmarkPublishAndReceive(b *testing.B) {
	registry := NewRegistry()
	topic := registry.GetTopic(fmt.Sprintf("benchmark-%v", b.N))
	subscription := topic.GetSubscription("test-sub", time.Minute)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := NewMessage("test-message", "test-payload", time.Minute)
		subscription.Topic.Publish(msg)
		<-subscription.Response
	}
}

func BenchmarkPublishAndReceiveUniqueTopic(b *testing.B) {
	registry := NewRegistry()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		topic := registry.GetTopic(fmt.Sprintf("bench-%v", i))
		subscription := topic.GetSubscription(fmt.Sprintf("sub-%v", i), time.Minute)
		msg := NewMessage(fmt.Sprintf("message-%v", i), "test-payload", time.Minute)
		subscription.Topic.Publish(msg)
		<-subscription.Response
	}
}
