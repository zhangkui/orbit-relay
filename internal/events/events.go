package events

import (
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"
)

type Type string

const (
	TypeWindowPlanned Type = "window.planned"
	TypeWindowActive  Type = "window.active"
	TypeWindowClosed  Type = "window.closed"
	TypeFrameAccepted Type = "frame.accepted"
	TypeFrameRejected Type = "frame.rejected"
	TypeCommandQueued Type = "command.queued"
	TypeCommandSent   Type = "command.sent"
	TypeCommandFailed Type = "command.failed"
)

type Event struct {
	ID      uint64         `json:"id"`
	Type    Type           `json:"type"`
	Subject string         `json:"subject"`
	At      time.Time      `json:"at"`
	Data    map[string]any `json:"data"`
}
type Bus struct {
	mu     sync.RWMutex
	next   uint64
	events []Event
	subs   map[Type][]chan Event
}

func New() *Bus { return &Bus{subs: map[Type][]chan Event{}} }
func (b *Bus) Publish(t Type, subject string, data map[string]any) Event {
	// Publish mutates b.next and b.events, so it must hold the write lock
	// exclusively. A read lock here allowed concurrent publishers to race on
	// the sequence counter (duplicate/lost IDs) and on append (overwritten or
	// lost history), leaving subscribers with an unreliable view of the
	// communication window state.
	b.mu.Lock()
	b.next++
	e := Event{ID: b.next, Type: t, Subject: subject, At: time.Now().UTC(), Data: data}
	b.events = append(b.events, e)
	subs := append([]chan Event(nil), b.subs[t]...)
	b.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- e:
		default:
		}
	}
	return e
}
func (b *Bus) Subscribe(t Type, buffer int) <-chan Event {
	if buffer < 1 {
		buffer = 1
	}
	ch := make(chan Event, buffer)
	b.mu.Lock()
	b.subs[t] = append(b.subs[t], ch)
	b.mu.Unlock()
	return ch
}
func (b *Bus) List(subject string, from uint64) []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := []Event{}
	for _, e := range b.events {
		if e.ID >= from && (subject == "" || e.Subject == subject) {
			e.Data = clone(e.Data)
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func clone(in map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		if nested, ok := v.(map[string]any); ok {
			v = clone(nested)
		}
		out[k] = v
	}
	return out
}
func Replay(items []Event, fn func(Event) error) error {
	for _, e := range items {
		if e.Type == "" || e.Subject == "" {
			return errors.New("invalid event")
		}
		if err := fn(e); err != nil {
			return err
		}
	}
	return nil
}
func Encode(e Event) ([]byte, error)   { return json.Marshal(e) }
func Decode(raw []byte) (Event, error) { var e Event; err := json.Unmarshal(raw, &e); return e, err }
