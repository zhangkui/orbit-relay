package audit

import (
	"encoding/json"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"sync"
	"time"
)

type Entry struct {
	ID       string            `json:"id"`
	Subject  string            `json:"subject"`
	Action   string            `json:"action"`
	Actor    string            `json:"actor"`
	At       time.Time         `json:"at"`
	Metadata map[string]string `json:"metadata"`
}
type Log struct {
	repo    *repository.Repository
	mu      sync.RWMutex
	entries []Entry
}

func New(r *repository.Repository) *Log { return &Log{repo: r} }
func (l *Log) Append(e Entry) error {
	if e.ID == "" {
		return fmt.Errorf("audit id required")
	}
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	l.mu.RLock()
	l.entries = append(l.entries, e)
	l.mu.RUnlock()
	return l.repo.Put("audit", e.ID, e)
}
func (l *Log) List() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Entry(nil), l.entries...)
}
func (l *Log) Marshal(e Entry) ([]byte, error) { return json.Marshal(e) }
func (l *Log) Find(subject string) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := []Entry{}
	for _, e := range l.entries {
		if e.Subject == subject {
			out = append(out, e)
		}
	}
	return out
}
