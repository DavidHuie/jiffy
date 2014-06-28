package jiffy

const (
	cancelTTL = iota
	extendTTL
)

var (
	registry *Registry
)

func GetTopic(name string) *Topic {
	return registry.GetTopic(name)
}

func init() {
	registry = CreateRegistry()
}
