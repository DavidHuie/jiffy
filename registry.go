package jiffy

import (
	"sync"
)

type Registry struct {
	Topics     map[string]*Topic
	topicMutex sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		make(map[string]*Topic),
		sync.Mutex{},
	}
}

// Creates a topic from the registry, creating one if it didn't
// exist.
func (registry *Registry) GetTopic(name string) *Topic {
	registry.topicMutex.Lock()
	defer registry.topicMutex.Unlock()

	if topic, ok := registry.Topics[name]; ok {
		return topic
	}
	registry.Topics[name] = NewTopic(name, registry)
	return registry.Topics[name]
}

// Cleans all expired data from the registry.
func (registry *Registry) Clean() {
	for _, topic := range registry.Topics {
		topic.CleanExpiredSubscriptions()
		topic.CleanExpiredCachedMessages()

	}
}
