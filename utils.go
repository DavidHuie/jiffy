package jiffy

import (
	"time"
)

const (
	cancelTTL = iota
	extendTTL
)

// Calls the input function after the input TTL has passed. However,
// the TTL is restarted if a message is received on input channel.
func CallAfterTTL(f func(), ttl time.Duration, c chan int) {
	ticker := time.NewTicker(ttl)
	for {
		select {
		case <-ticker.C:
			f()
			return
		case message := <-c:
			if message == cancelTTL {
				return
			}
			ticker = time.NewTicker(ttl)
		}
	}
}
