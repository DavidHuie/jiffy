package jiffy

import (
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
	if len(topic.Data) != 0 {
		t.Errorf("Invalid data count")
	}
	// Getting the topic again should return the exact
	// same topic.
	topic2 := GetTopic("test_topic1")
	if topic.uuid != topic2.uuid {
		t.Errorf("Topics should be equal")
	}
}

func TestTopicCachedData(t *testing.T) {
	topic := GetTopic("test_topic2")
	message1 := NewMessage("test-name1", "my message1")
	message2 := NewMessage("test-name2", "my message2")
	topic.Record(message1, 50*time.Millisecond)
	topic.Record(message2, 50*time.Millisecond)

	messages := topic.CachedData()
	if len(messages) != 2 {
		t.Errorf("Invalid number of messages returned")
	}
	if (messages[0] != message1) || (messages[1] != message2) {
		t.Errorf("Invalid messages returned")
	}

	time.Sleep(100 * time.Millisecond)
	messages = topic.CachedData()
	if numMsgs := len(messages); numMsgs != 0 {
		t.Errorf("Invalid number of messages returned: %v", numMsgs)
	}
}

func TestSubscriptionResponse(t *testing.T) {
	topic := GetTopic("test_topic3")
	message := NewMessage("test-message", "my message")
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
	message := NewMessage("test-name1", "my message1")
	topic.RecordAndPublish(message, time.Minute)

	messages := topic.CachedData()
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
	sub := topic.GetSubscription("sub", 100*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	if sub.Active() != false {
		t.Errorf("Subscription should be inactive")
	}
}

func TestSubExtendExpiration(t *testing.T) {
	topic := GetTopic("test_topic6")
	sub := topic.GetSubscription("sub", 100*time.Millisecond)

	// Sleep so that the expiration goroutine is queued.
	// TODO: take this out.
	time.Sleep(10 * time.Millisecond)

	if status := sub.Active(); status != true {
		t.Errorf("Subscription should be active, got %v", status)
	}

	// TODO: ensure that subscription expirations are extendable
	// immediately after creating them.
	sub.ExtendExpiration(300 * time.Millisecond)
	time.Sleep(200 * time.Millisecond)

	if status := sub.Active(); status != true {
		t.Errorf("Subscription should be active, got %v", status)
	}

	time.Sleep(150 * time.Millisecond)

	if status := sub.Active(); status != false {
		t.Errorf("Subscription should be inactive, got %v", status)
	}
}
