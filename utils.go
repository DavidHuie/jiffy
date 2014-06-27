package main

import (
	"time"
)

// Calls the input function after the input TTL has passed. However,
// the TTL is restarted if a message is received on input channel.
func CallAfterTTL(f func(), ttl time.Duration, c chan string) {
	ticker := time.NewTicker(ttl)
	for {
		select {
		case <-ticker.C:
			f()
			return
		case message := <-c:
			if message == "cancel" {
				return
			}
			ticker = time.NewTicker(ttl)
		}
	}
}
