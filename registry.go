package jiffy

import (
	"sync"
)

type Registry struct {
	Topics        map[string]*Topic
	NewTopicMutex sync.Mutex
}

func CreateRegistry() *Registry {
	return &Registry{
		make(map[string]*Topic),
		sync.Mutex{},
	}
}

// Creates a topic on the registry if it doesn't exist
// and returns it.
func (registry *Registry) GetTopic(name string) *Topic {
	registry.NewTopicMutex.Lock()
	defer registry.NewTopicMutex.Unlock()
	if topic, ok := registry.Topics[name]; ok {
		return topic
	}
	registry.Topics[name] = CreateTopic(name)
	return registry.Topics[name]
}
