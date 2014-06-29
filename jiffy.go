package jiffy

import (
	"time"
)

var (
	registry *Registry

	// How often to clean expired messages, expired subscriptions,
	// and subscription-less topics.
	cleanInterval = 10 * time.Second
)

// Returns a topic from the global registry.
func GetTopic(name string) *Topic {
	return registry.GetTopic(name)
}

func init() {
	registry = NewRegistry()

	// Periodically clean the registry.
	go func() {
		ticker := time.NewTicker(cleanInterval)
		for {
			<-ticker.C
			registry.Clean()
		}
	}()
}
