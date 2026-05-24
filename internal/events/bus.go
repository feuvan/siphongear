package events

import (
	"sync"
)

// Event types
type Event struct {
	Type    string         // run.completed | indicator.changed
	Payload map[string]any
}

type Bus struct {
	mu   sync.RWMutex
	subs map[string][]chan Event
}

func NewBus() *Bus {
	return &Bus{subs: map[string][]chan Event{}}
}

func (b *Bus) Subscribe(topic string, buf int) <-chan Event {
	ch := make(chan Event, buf)
	b.mu.Lock()
	b.subs[topic] = append(b.subs[topic], ch)
	b.mu.Unlock()
	return ch
}

func (b *Bus) Publish(topic string, payload map[string]any) {
	b.mu.RLock()
	chans := append([]chan Event{}, b.subs[topic]...)
	b.mu.RUnlock()
	ev := Event{Type: topic, Payload: payload}
	for _, ch := range chans {
		select {
		case ch <- ev:
		default:
		}
	}
}
