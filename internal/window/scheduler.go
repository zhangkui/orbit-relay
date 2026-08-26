package window

import (
	"context"
	"gitlab.com/zhangkui/orbit-relay/internal/clock"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
	"time"
)

type Scheduler struct{ now clock.Clock }

func NewScheduler(c clock.Clock) *Scheduler {
	if c == nil {
		c = clock.System{}
	}
	return &Scheduler{now: c}
}
func (s *Scheduler) Activate(ctx context.Context, w *model.ContactWindow) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return Transition(w, model.WindowActive, s.now.Now())
}
func (s *Scheduler) Close(ctx context.Context, w *model.ContactWindow, failed bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	next := model.WindowCompleted
	if failed {
		next = model.WindowFailed
	}
	return Transition(w, next, s.now.Now())
}
func SortWindows(items []model.ContactWindow) []model.ContactWindow {
	out := append([]model.ContactWindow(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].Start.Before(out[j].Start)
		}
		return out[i].Priority > out[j].Priority
	})
	return out
}
func Overlap(a, b model.ContactWindow) bool {
	return a.StationID == b.StationID && a.Antenna == b.Antenna && a.Start.Before(b.End) && b.Start.Before(a.End)
}
func WaitUntil(ctx context.Context, t time.Time, c clock.Clock) error {
	d := time.Until(t)
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
