package jiffy

import (
	"sync"
	"time"
)

type Registry struct {
	Topics     map[string]*Topic
	topicMutex sync.Mutex
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
	registry.topicMutex.Lock()
	defer registry.topicMutex.Unlock()
	if topic, ok := registry.Topics[name]; ok {
		return topic
	}
	registry.Topics[name] = CreateTopic(name)
	return registry.Topics[name]
}

// In intervals, subscriptionless topics are deleted from the registry.
func (registry *Registry) CleanTopics(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		for topicName, topic := range registry.Topics {
			go func(name string, topic *Topic, registry *Registry) {
				registry.topicMutex.Lock()
				defer registry.topicMutex.Unlock()
				if len(topic.Subscriptions) == 0 {
					delete(registry.Topics, topicName)
				}
			}(topicName, topic, registry)
		}
	}
}
