package command

import (
	"context"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/ingest"
	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/protocol"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"sort"
	"sync"
	"time"
)

type Queue struct {
	repo    *repository.Repository
	jobs    *ingest.Queue
	metrics *metrics.Registry
	mu      sync.Mutex
	packets map[string]model.CommandPacket
}

func NewQueue(r *repository.Repository, m *metrics.Registry) *Queue {
	return &Queue{repo: r, jobs: ingest.New(32), metrics: m, packets: map[string]model.CommandPacket{}}
}
func (q *Queue) Close() { q.jobs.Close() }
func (q *Queue) Enqueue(ctx context.Context, p model.CommandPacket) error {
	if err := protocol.ValidateCommand(p); err != nil {
		return fmt.Errorf("command validation: %w", err)
	}
	p.State = model.CommandQueued
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	q.mu.Lock()
	q.packets[p.ID] = p
	q.mu.Unlock()
	if err := q.repo.Put("command", p.ID, p); err != nil {
		return err
	}
	q.metrics.Add("commands.queued", 1)
	return nil
}
func (q *Queue) List() []model.CommandPacket {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]model.CommandPacket, 0, len(q.packets))
	for _, p := range q.packets {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].Priority > out[j].Priority
	})
	return out
}
func (q *Queue) Send(ctx context.Context, id string, send func([]byte) error) error {
	q.mu.Lock()
	p, ok := q.packets[id]
	q.mu.Unlock()
	if !ok {
		return model.ErrNotFound
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	raw, err := protocol.EncodeCommand(p)
	if err != nil {
		return err
	}
	p.State = model.CommandSending
	if err := send(raw); err != nil {
		p.State = model.CommandFailed
		p.Error = err.Error()
		q.metrics.Add("commands.failed", 1)
	} else {
		now := time.Now().UTC()
		p.State = model.CommandSent
		p.SentAt = &now
		q.metrics.Add("commands.sent", 1)
	}
	q.mu.Lock()
	q.packets[id] = p
	q.mu.Unlock()
	_ = q.repo.Put("command", id, p)
	if p.State == model.CommandSent {
		return fmt.Errorf("send command: %w", err)
	}
	return nil
}
