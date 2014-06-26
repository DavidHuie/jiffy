package main

type Item struct {
	Name    string
	Watches map[string]*Watch
	State   map[string]*Event
}

type Watch struct {
	Name          string
	Item          *Item
	ClientChannel chan *Event
}

type Event struct {
	Name    string
	Payload []byte
}

type Registry struct {
	Items map[string]*Item
}

func CreateItem(name string) *Item {
	return &Item{
		name,
		make(map[string]*Watch),
		make(map[string]*Event),
	}
}

func (item *Item) Publish(event *Event) {
	for _, watch := range item.Watches {
		go func() {
			watch.ClientChannel <- event
		}()
	}
}

func (item *Item) RecordAndPublish(event *Event) {
	item.State[event.Name] = event
	item.Publish(event)
}

func (item *Item) DestroyAndPublish(event *Event) {
	delete(item.State, event.Name)
	item.Publish(event)
}

func CreateWatch(name string, item *Item, channel chan *Event) *Watch {
	watch := &Watch{name, item, channel}
	item.Watches[name] = watch
	return watch
}

// Session lifetime
// * Client connects
// * Client creates Item if it doesn't already exist
// * Client creates a Watch on the item
// * Client receives a a payload consisting of real-time state
//   of the item.
// * When the client emits an event, it gets forwarded to
//   the item and is fanned out to all Watches.

func main() {

}
