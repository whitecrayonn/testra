package eventbus

import (
	"context"
	"sync"
	"time"
)

// Event represents an application event published on the internal bus.
type Event struct {
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Listener receives events through a channel.
type Listener struct {
	ch     chan Event
	closed bool
}

func newListener(buffer int) *Listener {
	return &Listener{ch: make(chan Event, buffer)}
}

// Chan returns the receive channel for the listener.
func (l *Listener) Chan() <-chan Event { return l.ch }

// Close removes the listener and closes its channel. Safe for repeated calls.
func (l *Listener) Close() {
	if l.closed {
		return
	}
	l.closed = true
	close(l.ch)
}

// Bus is a lightweight in-memory event dispatcher backed by Go channels.
type Bus struct {
	mu        sync.RWMutex
	listeners []*Listener
	buffer    int
	closed    bool
}

// New creates a Bus with the given per-listener buffer size.
func New(buffer int) *Bus {
	if buffer <= 0 {
		buffer = 64
	}
	return &Bus{buffer: buffer}
}

// Subscribe registers a new listener and returns it.
func (b *Bus) Subscribe() *Listener {
	l := newListener(b.buffer)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		l.Close()
		return l
	}
	b.listeners = append(b.listeners, l)
	return l
}

// SubscribeFunc registers a callback that runs for each event in its own goroutine.
func (b *Bus) SubscribeFunc(fn func(Event)) {
	l := b.Subscribe()
	go func() {
		for e := range l.Chan() {
			fn(e)
		}
	}()
}

// Unsubscribe removes a listener and closes its channel.
func (b *Bus) Unsubscribe(l *Listener) {
	l.Close()
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, cur := range b.listeners {
		if cur == l {
			b.listeners = append(b.listeners[:i], b.listeners[i+1:]...)
			return
		}
	}
}

var defaultBus *Bus

func init() {
	defaultBus = New(256)
}

// Default returns the package-level default bus.
func Default() *Bus { return defaultBus }

// SetDefault replaces the package-level default bus. Call once during app startup.
func SetDefault(b *Bus) { defaultBus = b }

// Publish sends an event to all active listeners without blocking the caller.
func (b *Bus) Publish(ctx context.Context, e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	for _, l := range b.listeners {
		if l.closed {
			continue
		}
		select {
		case l.ch <- e:
		default:
			// Channel is full; drop to preserve caller latency. Listeners that
			// cannot keep up will miss events rather than block the publisher.
		}
	}
}

// Close shuts down the bus and all listeners.
func (b *Bus) Close() {
	b.mu.Lock()
	listeners := b.listeners
	b.listeners = nil
	b.closed = true
	b.mu.Unlock()
	for _, l := range listeners {
		l.Close()
	}
}
