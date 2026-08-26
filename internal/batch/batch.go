package batch

import (
	"context"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/command"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
	"sync"
	"time"
)

type Batch struct {
	ID          string
	WindowID    string
	Packets     []model.CommandPacket
	State       string
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Error       string
}

const (
	Planned   = "planned"
	Running   = "running"
	Completed = "completed"
	Failed    = "failed"
	Cancelled = "cancelled"
)

type Runner struct {
	queue   *command.Queue
	mu      sync.Mutex
	batches map[string]Batch
}

func NewRunner(q *command.Queue) *Runner { return &Runner{queue: q, batches: map[string]Batch{}} }
func (r *Runner) Create(id, window string, packets []model.CommandPacket) (Batch, error) {
	if id == "" || window == "" || len(packets) == 0 {
		return Batch{}, model.ErrInvalid
	}
	copyPackets := append([]model.CommandPacket(nil), packets...)
	sort.SliceStable(copyPackets, func(i, j int) bool { return copyPackets[i].Priority > copyPackets[j].Priority })
	b := Batch{ID: id, WindowID: window, Packets: copyPackets, State: Planned, CreatedAt: time.Now().UTC()}
	r.mu.Lock()
	r.batches[id] = b
	r.mu.Unlock()
	return b, nil
}
func (r *Runner) Run(ctx context.Context, id string, send func([]byte) error) error {
	r.mu.Lock()
	b, ok := r.batches[id]
	if !ok {
		r.mu.Unlock()
		return model.ErrNotFound
	}
	if b.State != Planned {
		r.mu.Unlock()
		return model.ErrConflict
	}
	b.State = Running
	now := time.Now().UTC()
	b.StartedAt = &now
	r.batches[id] = b
	r.mu.Unlock()
	for _, p := range b.Packets {
		if err := ctx.Err(); err != nil {
			return r.fail(id, err)
		}
		if err := r.queue.Send(ctx, p.ID, send); err != nil {
			return r.fail(id, err)
		}
	}
	r.mu.Lock()
	b = r.batches[id]
	b.State = Completed
	done := time.Now().UTC()
	b.CompletedAt = &done
	r.batches[id] = b
	r.mu.Unlock()
	return nil
}
func (r *Runner) fail(id string, err error) error {
	r.mu.Lock()
	b := r.batches[id]
	b.State = Failed
	b.Error = err.Error()
	r.batches[id] = b
	r.mu.Unlock()
	return fmt.Errorf("batch %s: %w", id, err)
}
func (r *Runner) Cancel(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.batches[id]
	if !ok {
		return model.ErrNotFound
	}
	if b.State != Planned {
		return model.ErrConflict
	}
	b.State = Cancelled
	r.batches[id] = b
	return nil
}
func (r *Runner) Get(id string) (Batch, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.batches[id]
	return b, ok
}
