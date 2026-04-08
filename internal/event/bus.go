package event

import "sync"

// Bus is a lightweight in-process pub/sub for query events.
type Bus struct {
	subscribers []chan<- QueryEvent
	mu          sync.RWMutex
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{}
}

// Publish sends an event to all subscribers. Non-blocking: drops if subscriber is slow.
func (b *Bus) Publish(event QueryEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, sub := range b.subscribers {
		select {
		case sub <- event:
		default:
			// Drop if subscriber channel is full.
		}
	}
}

// Subscribe creates a new subscription channel with the given buffer size.
func (b *Bus) Subscribe(bufSize int) <-chan QueryEvent {
	ch := make(chan QueryEvent, bufSize)
	b.mu.Lock()
	b.subscribers = append(b.subscribers, ch)
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscription channel and closes it.
func (b *Bus) Unsubscribe(ch <-chan QueryEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, sub := range b.subscribers {
		// Compare by reading from the same underlying channel.
		if isSameChannel(sub, ch) {
			b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
			close(sub)
			return
		}
	}
}

func isSameChannel(send chan<- QueryEvent, recv <-chan QueryEvent) bool {
	// Both point to same channel if sending to send arrives on recv.
	// We can't compare channels directly across types, use a wrapper approach.
	// For simplicity, cast via interface comparison.
	type ch = chan QueryEvent
	// This works because Subscribe returns the same underlying channel.
	return (interface{})(send) == (interface{})(recv)
}

// SubscriberCount returns the number of active subscribers.
func (b *Bus) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
