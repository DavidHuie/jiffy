package jiffy

import (
	"time"
)

var (
	registry      *Registry
	cleanInterval = time.Minute
)

// Returns a topic from the global registry.
func GetTopic(name string) *Topic {
	return registry.GetTopic(name)
}

func init() {
	registry = CreateRegistry()
	go registry.CleanTopics(cleanInterval)
}
