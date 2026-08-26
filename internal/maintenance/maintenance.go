package maintenance

import (
	"context"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
	"sync"
	"time"
)

type Severity string

const (
	Info     Severity = "info"
	Warning  Severity = "warning"
	Critical Severity = "critical"
)

type Check struct {
	ID       string
	Name     string
	Category string
	Severity Severity
	Interval time.Duration
	LastRun  *time.Time
	NextRun  time.Time
	Healthy  bool
	Message  string
}
type Record struct {
	ID           string
	SatelliteID  string
	CheckID      string
	Started      time.Time
	Finished     time.Time
	Healthy      bool
	Message      string
	Measurements map[string]float64
}
type Registry struct {
	mu      sync.RWMutex
	checks  map[string]Check
	records []Record
}

func NewRegistry() *Registry { return &Registry{checks: map[string]Check{}} }
func (r *Registry) Register(c Check) error {
	if c.ID == "" || c.Name == "" || c.Interval <= 0 {
		return model.ErrInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.checks[c.ID]; ok {
		return model.ErrConflict
	}
	r.checks[c.ID] = c
	return nil
}
func (r *Registry) Due(now time.Time) []Check {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Check{}
	for _, c := range r.checks {
		if c.NextRun.IsZero() || !c.NextRun.After(now) {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].NextRun.Before(out[j].NextRun) })
	return out
}
func (r *Registry) Run(ctx context.Context, satellite, checkID string, fn func(context.Context) (bool, string, map[string]float64)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.RLock()
	c, ok := r.checks[checkID]
	r.mu.RUnlock()
	if !ok {
		return model.ErrNotFound
	}
	start := time.Now().UTC()
	healthy, msg, measure := fn(ctx)
	finish := time.Now().UTC()
	rec := Record{ID: fmt.Sprintf("%s-%s-%d", satellite, checkID, start.UnixNano()), SatelliteID: satellite, CheckID: checkID, Started: start, Finished: finish, Healthy: healthy, Message: msg, Measurements: measure}
	r.mu.Lock()
	r.records = append(r.records, rec)
	c.LastRun = &finish
	c.NextRun = finish.Add(c.Interval)
	c.Healthy = healthy
	c.Message = msg
	r.checks[checkID] = c
	r.mu.Unlock()
	return nil
}
func (r *Registry) Records(satellite string) []Record {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Record{}
	for _, x := range r.records {
		if satellite == "" || x.SatelliteID == satellite {
			out = append(out, x)
		}
	}
	return out
}
func (r *Registry) Health(satellite string) bool {
	records := r.Records(satellite)
	if len(records) == 0 {
		return false
	}
	for _, x := range records {
		if !x.Healthy {
			return false
		}
	}
	return true
}
func (r *Registry) Summary(satellite string) map[string]int {
	out := map[string]int{"healthy": 0, "failed": 0}
	for _, x := range r.Records(satellite) {
		if x.Healthy {
			out["healthy"]++
		} else {
			out["failed"]++
		}
	}
	return out
}
func AgeOf(c Check, now time.Time) time.Duration {
	if c.LastRun == nil {
		return time.Duration(1<<63 - 1)
	}
	return now.Sub(*c.LastRun)
}
func IsStale(c Check, now time.Time) bool {
	return c.LastRun == nil || now.Sub(*c.LastRun) > c.Interval*2
}
