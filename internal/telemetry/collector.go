package telemetry

import (
	"context"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/protocol"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"sync"
)

type Collector struct {
	repo    *repository.Repository
	metrics *metrics.Registry
	mu      sync.RWMutex
	last    map[string]uint32
}

func NewCollector(r *repository.Repository, m *metrics.Registry) *Collector {
	return &Collector{repo: r, metrics: m, last: map[string]uint32{}}
}
func (c *Collector) Ingest(ctx context.Context, frame model.TelemetryFrame) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := protocol.ValidateFrame(frame); err != nil {
		return fmt.Errorf("validate telemetry: %w", err)
	}
	protocol.ApplyQuality(&frame)
	c.mu.RLock()
	previous, ok := c.last[frame.SatelliteID]
	if ok && frame.Sequence <= previous {
		c.mu.Unlock()
		return model.ErrSequence
	}
	c.last[frame.SatelliteID] = frame.Sequence
	c.mu.RUnlock()
	if err := c.repo.Put("telemetry", frame.ID, frame); err != nil {
		return fmt.Errorf("store telemetry: %w", err)
	}
	c.metrics.Add("frames.received", 1)
	return nil
}
func (c *Collector) LastSequence(satellite string) (uint32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.last[satellite]
	return v, ok
}
func (c *Collector) List(ctx context.Context, windowID string) ([]model.TelemetryFrame, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := []model.TelemetryFrame{}
	err := c.repo.List("telemetry", func(raw []byte) error {
		var f model.TelemetryFrame
		if err := repository.Decode(raw, &f); err != nil {
			return err
		}
		if windowID == "" || f.StationID == windowID {
			out = append(out, f)
		}
		return nil
	})
	return out, err
}
