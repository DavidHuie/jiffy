package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/DavidHuie/jiffy"
)

const (
	SessionTimeout = 20 * time.Second
)

var (
	registry *jiffy.Registry
)

func subscriptionHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	topicName := request.Form.Get("topic")
	sessionID := request.Form.Get("id")

	topic := registry.GetTopic(topicName)
	subscription := topic.GetSubscription(sessionID)

	log.Printf("Subscription request: Id=%v TopicName=%v", sessionID, topicName)

	// Create a ticker that'll notify us when it's time to timeout
	// the session.
	timeoutTicker := time.NewTicker(SessionTimeout)

	select {
	case <-timeoutTicker.C:
		response.WriteHeader(http.StatusOK)
		return
	case event := <-subscription.ResponseChannel:
		// Batching would be good to have here, otherwise we'll
		// have to perform one request per event.
		jsonEvent, err := json.Marshal(event)
		if err != nil {
			log.Println(err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.Write(jsonEvent)
		return
	}
}

// Publishes a message
func publishHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	topicName := request.Form.Get("topic")
	topic := registry.GetTopic(topicName)
	messagePayload := request.Form.Get("message")
	message := jiffy.NewMessage("asdfjk;l", messagePayload)
	topic.RecordAndPublish(message)
	response.WriteHeader(http.StatusOK)
}

func init() {
	registry = jiffy.CreateRegistry()
}

func main() {
	http.HandleFunc("/subscribe", subscriptionHandler)
	http.HandleFunc("/publish", publishHandler)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
