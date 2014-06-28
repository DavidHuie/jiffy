package jiffy

import (
	"time"
)

const (
	cancelTTL = iota
	extendTTL
)

var (
	registry      *Registry
	cleanInterval = 10 * time.Minute
)

// Returns a topic from the global registry.
func GetTopic(name string) *Topic {
	return registry.GetTopic(name)
}

func init() {
	registry = CreateRegistry()
	go registry.CleanTopics(cleanInterval)
}
