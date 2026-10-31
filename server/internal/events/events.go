// Package events is a tiny in-process publish/subscribe bus.
//
// The web dashboard needs to show what the server already knows (lobbies,
// automatch matches, players, map pool changes) as it happens. Publishing is
// fire-and-forget: a slow or missing subscriber must never block the game
// request that produced the event, so full subscriber buffers drop events
// instead of blocking the caller. A subscriber that receives a drop can simply
// re-fetch the snapshot endpoints, which are always authoritative.
package events

import (
	"sync"
	"sync/atomic"
	"time"
)

// Event is one change worth streaming to the dashboard.
type Event struct {
	// ID is monotonically increasing for the lifetime of the process. It is
	// also used as the SSE event id so a client can detect gaps.
	ID uint64 `json:"id"`
	// Type is a stable dotted name, e.g. "lobby.created" or "pool.updated".
	Type string `json:"type"`
	// Game is the game id the event belongs to (age2, age4, ...). May be empty
	// for process-wide events.
	Game string `json:"game,omitempty"`
	// Time is when the event was published.
	Time time.Time `json:"time"`
	// Data is the optional event payload. It must be JSON-encodable.
	Data any `json:"data,omitempty"`
}

// subscriberBuffer is how many events a single subscriber may lag before its
// oldest pending events are dropped.
const subscriberBuffer = 128

// Bus is a concurrency-safe fan-out of Event values.
type Bus struct {
	mu     sync.RWMutex
	subs   map[uint64]chan Event
	nextID uint64
	lastID atomic.Uint64
}

// NewBus returns an empty bus.
func NewBus() *Bus {
	return &Bus{subs: make(map[uint64]chan Event)}
}

// Default is the bus used by the game models and the web server.
var Default = NewBus()

// Publish assigns an id and timestamp to ev, then fans it out to every
// subscriber. It never blocks.
func (b *Bus) Publish(ev Event) Event {
	if ev.Time.IsZero() {
		ev.Time = time.Now().UTC()
	}
	ev.ID = b.lastID.Add(1)

	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- ev:
		default:
			// The subscriber is behind. Dropping is preferred over stalling a
			// game request; the client can resync from the snapshot endpoints.
		}
	}
	return ev
}

// Subscribe returns a channel of events plus a cancel function. The channel is
// closed by cancel, so a range loop terminates cleanly. Cancel is safe to call
// more than once.
func (b *Bus) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, subscriberBuffer)

	b.mu.Lock()
	id := b.nextID
	b.nextID++
	b.subs[id] = ch
	b.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subs, id)
			b.mu.Unlock()
			close(ch)
		})
	}
	return ch, cancel
}

// LastID reports the id of the most recently published event.
func (b *Bus) LastID() uint64 { return b.lastID.Load() }

// Publish sends an event on the default bus. It is the convenience form used by
// the game models.
func Publish(eventType string, game string, data any) Event {
	return Default.Publish(Event{Type: eventType, Game: game, Data: data})
}

// Subscribe subscribes to the default bus.
func Subscribe() (<-chan Event, func()) { return Default.Subscribe() }

// LastID reports the last id published on the default bus.
func LastID() uint64 { return Default.LastID() }
