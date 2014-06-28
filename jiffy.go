package jiffy

const (
	cancelTTL = iota
	extendTTL
)

var (
	registry *Registry
)

// Returns a topic from the global registry.
func GetTopic(name string) *Topic {
	return registry.GetTopic(name)
}

func init() {
	registry = CreateRegistry()
}
