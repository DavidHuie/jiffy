package main

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/DavidHuie/jiffy"
)

const (
	SessionTimeout = 20 * time.Second
)

func subscriptionHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	topicName := request.Form.Get("topic")
	sessionID := request.Form.Get("id")

	topic := jiffy.GetTopic(topicName)
	subscription := topic.GetSubscription(sessionID, 60*time.Second)

	log.Printf("Subscription request: Id=%v TopicName=%v", sessionID, topicName)

	// Create a ticker that'll notify us when it's time to timeout
	// the session.
	timeoutTicker := time.NewTicker(SessionTimeout)

	select {
	case <-timeoutTicker.C:
		response.WriteHeader(http.StatusOK)
		return
	case message := <-subscription.Response:
		// Batching would be good to have here, otherwise we'll
		// have to perform one request per message.
		jsonMessage, err := json.Marshal(message)
		if err != nil {
			log.Println(err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.Write(jsonMessage)
		return
	}
}

// Publishes a message
func publishHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	topicName := request.Form.Get("topic")
	topic := jiffy.GetTopic(topicName)
	messagePayload := request.Form.Get("message")
	message := jiffy.NewMessage("asdfjk;l", messagePayload)
	topic.RecordAndPublish(message, 60*time.Second)
	response.WriteHeader(http.StatusOK)
}

func main() {
	runtime.GOMAXPROCS(8)
	http.HandleFunc("/subscribe", subscriptionHandler)
	http.HandleFunc("/publish", publishHandler)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
