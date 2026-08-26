package service

import (
	"context"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/clock"
	"gitlab.com/zhangkui/orbit-relay/internal/command"
	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/report"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/telemetry"
	"gitlab.com/zhangkui/orbit-relay/internal/window"
	"sync"
	"time"
)

type Lab struct {
	repo       *repository.Repository
	windows    map[string]model.ContactWindow
	satellites map[string]model.Satellite
	stations   map[string]model.GroundStation
	mu         sync.RWMutex
	scheduler  *window.Scheduler
	collector  *telemetry.Collector
	commands   *command.Queue
	reports    *report.Builder
}

func New(r *repository.Repository) *Lab {
	m := metrics.New()
	c := telemetry.NewCollector(r, m)
	q := command.NewQueue(r, m)
	l := &Lab{repo: r, windows: map[string]model.ContactWindow{}, satellites: map[string]model.Satellite{}, stations: map[string]model.GroundStation{}, scheduler: window.NewScheduler(clock.System{}), collector: c, commands: q}
	l.reports = report.New(c)
	return l
}
func (l *Lab) Close() { l.commands.Close() }
func (l *Lab) AddSatellite(ctx context.Context, s model.Satellite) error {
	if s.ID == "" || s.Name == "" {
		return model.ErrInvalid
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.satellites[s.ID] = s
	return l.repo.Put("satellite", s.ID, s)
}
func (l *Lab) AddStation(ctx context.Context, s model.GroundStation) error {
	if s.ID == "" {
		return model.ErrInvalid
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stations[s.ID] = s
	return l.repo.Put("station", s.ID, s)
}
func (l *Lab) PlanWindow(ctx context.Context, w model.ContactWindow) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if w.ID == "" || w.Start.After(w.End) || w.SatelliteID == "" || w.StationID == "" {
		return model.ErrInvalid
	}
	w.State = model.WindowPlanned
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, other := range l.windows {
		if window.Overlap(w, other) && !model.IsTerminalWindow(other.State) {
			return fmt.Errorf("window overlap: %w", model.ErrConflict)
		}
	}
	l.windows[w.ID] = w
	return l.repo.Put("window", w.ID, w)
}
func (l *Lab) GetWindow(id string) (model.ContactWindow, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	w, ok := l.windows[id]
	if !ok {
		return model.ContactWindow{}, model.ErrNotFound
	}
	return w, nil
}
func (l *Lab) ActivateWindow(ctx context.Context, id string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.windows[id]
	if !ok {
		return model.ErrNotFound
	}
	if err := l.scheduler.Activate(ctx, &w); err != nil {
		return err
	}
	l.windows[id] = w
	return l.repo.Put("window", id, w)
}
func (l *Lab) CloseWindow(ctx context.Context, id string, failed bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.windows[id]
	if !ok {
		return model.ErrNotFound
	}
	if err := l.scheduler.Close(ctx, &w, failed); err != nil {
		return err
	}
	l.windows[id] = w
	return l.repo.Put("window", id, w)
}
func (l *Lab) Ingest(ctx context.Context, f model.TelemetryFrame) error {
	return l.collector.Ingest(ctx, f)
}
func (l *Lab) QueueCommand(ctx context.Context, p model.CommandPacket) error {
	l.mu.RLock()
	w, ok := l.windows[p.WindowID]
	l.mu.RUnlock()
	if !ok {
		return model.ErrNotFound
	}
	if err := command.CanSend(w, p, time.Now().UTC()); err != nil {
		return err
	}
	return l.commands.Enqueue(ctx, p)
}
func (l *Lab) SendCommand(ctx context.Context, id string, send func([]byte) error) error {
	return l.commands.Send(ctx, id, send)
}
func (l *Lab) Report(ctx context.Context, id string) (model.LinkReport, error) {
	w, err := l.GetWindow(id)
	if err != nil {
		return model.LinkReport{}, err
	}
	sent := 0
	for _, p := range l.commands.List() {
		if p.State == model.CommandSent {
			sent++
		}
	}
	return l.reports.Build(ctx, w, sent)
}
func (l *Lab) ListCommands() []model.CommandPacket { return l.commands.List() }
