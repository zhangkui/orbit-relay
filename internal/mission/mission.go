package mission

import (
	"context"
	"errors"
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
	"sync"
	"time"
)

type State string

const (
	StateDraft    State = "draft"
	StateReady    State = "ready"
	StateRunning  State = "running"
	StatePaused   State = "paused"
	StateComplete State = "complete"
	StateAborted  State = "aborted"
)

type Step struct {
	ID         string
	Name       string
	Kind       string
	WindowID   string
	CommandIDs []string
	DependsOn  []string
	Timeout    time.Duration
	State      State
	StartedAt  *time.Time
	FinishedAt *time.Time
	Output     string
	Error      string
}
type Mission struct {
	ID          string
	Name        string
	SatelliteID string
	Steps       []Step
	State       State
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Labels      map[string]string
}
type Plan struct {
	MissionID string
	Steps     []Step
	Current   int
	Started   time.Time
	Finished  *time.Time
}
type Engine struct {
	mu       sync.RWMutex
	missions map[string]Mission
	plans    map[string]Plan
}

func NewEngine() *Engine { return &Engine{missions: map[string]Mission{}, plans: map[string]Plan{}} }
func (e *Engine) Create(m Mission) error {
	if m.ID == "" || m.Name == "" || m.SatelliteID == "" {
		return model.ErrInvalid
	}
	if m.State == "" {
		m.State = StateDraft
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	m.UpdatedAt = m.CreatedAt
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.missions[m.ID]; ok {
		return model.ErrConflict
	}
	e.missions[m.ID] = m
	return nil
}
func (e *Engine) Get(id string) (Mission, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	m, ok := e.missions[id]
	if !ok {
		return Mission{}, model.ErrNotFound
	}
	m.Steps = append([]Step(nil), m.Steps...)
	return m, nil
}
func (e *Engine) AddStep(id string, s Step) error {
	if s.ID == "" || s.Name == "" {
		return model.ErrInvalid
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	m, ok := e.missions[id]
	if !ok {
		return model.ErrNotFound
	}
	if m.State != StateDraft && m.State != StateReady {
		return model.ErrConflict
	}
	for _, old := range m.Steps {
		if old.ID == s.ID {
			return model.ErrConflict
		}
	}
	s.State = StateDraft
	m.Steps = append(m.Steps, s)
	m.UpdatedAt = time.Now().UTC()
	e.missions[id] = m
	return nil
}
func (e *Engine) Validate(id string) error {
	e.mu.RLock()
	m, ok := e.missions[id]
	e.mu.RUnlock()
	if !ok {
		return model.ErrNotFound
	}
	if len(m.Steps) == 0 {
		return errors.New("mission has no steps")
	}
	seen := map[string]bool{}
	for _, s := range m.Steps {
		if seen[s.ID] {
			return fmt.Errorf("duplicate step: %w", model.ErrInvalid)
		}
		seen[s.ID] = true
		for _, dep := range s.DependsOn {
			if !seen[dep] {
				return fmt.Errorf("dependency %s appears later: %w", dep, model.ErrInvalid)
			}
		}
	}
	return nil
}
func (e *Engine) Prepare(id string) error {
	if err := e.Validate(id); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	m := e.missions[id]
	if m.State != StateDraft {
		return model.ErrConflict
	}
	m.State = StateReady
	m.UpdatedAt = time.Now().UTC()
	e.missions[id] = m
	return nil
}
func (e *Engine) Start(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	m, ok := e.missions[id]
	if !ok {
		return model.ErrNotFound
	}
	if m.State != StateReady && m.State != StatePaused {
		return model.ErrConflict
	}
	m.State = StateRunning
	m.UpdatedAt = time.Now().UTC()
	p := Plan{MissionID: id, Steps: append([]Step(nil), m.Steps...), Started: m.UpdatedAt}
	e.plans[id] = p
	e.missions[id] = m
	return nil
}
func (e *Engine) Pause(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	m, ok := e.missions[id]
	if !ok {
		return model.ErrNotFound
	}
	if m.State != StateRunning && m.State != StatePaused {
		return model.ErrConflict
	}
	m.State = StatePaused
	m.UpdatedAt = time.Now().UTC()
	e.missions[id] = m
	return nil
}
func (e *Engine) Abort(id, reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	m, ok := e.missions[id]
	if !ok {
		return model.ErrNotFound
	}
	if m.State == StateComplete || m.State == StateAborted {
		return model.ErrConflict
	}
	m.State = StateAborted
	m.UpdatedAt = time.Now().UTC()
	for i := range m.Steps {
		if m.Steps[i].State != StateComplete {
			m.Steps[i].State = StateAborted
			m.Steps[i].Error = reason
		}
	}
	e.missions[id] = m
	return nil
}
func (e *Engine) Next(id string) (Step, error) {
	e.mu.RLock()
	m, ok := e.missions[id]
	e.mu.RUnlock()
	if !ok {
		return Step{}, model.ErrNotFound
	}
	for _, s := range m.Steps {
		if s.State != StateComplete {
			ready := true
			for _, dep := range s.DependsOn {
				for _, candidate := range m.Steps {
					if candidate.ID == dep && candidate.State != StateComplete {
						ready = false
					}
				}
			}
			if ready {
				return s, nil
			}
		}
	}
	return Step{}, model.ErrNotFound
}
func (e *Engine) CompleteStep(id, stepID, output string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	m, ok := e.missions[id]
	if !ok {
		return model.ErrNotFound
	}
	for i := range m.Steps {
		if m.Steps[i].ID == stepID {
			if m.Steps[i].State == StateComplete {
				return model.ErrConflict
			}
			now := time.Now().UTC()
			m.Steps[i].State = StateComplete
			m.Steps[i].FinishedAt = &now
			m.Steps[i].Output = output
			m.UpdatedAt = now
			e.missions[id] = m
			e.updateMissionStateLocked(id, m)
			return nil
		}
	}
	return model.ErrNotFound
}
func (e *Engine) updateMissionStateLocked(id string, m Mission) {
	if m.State != StateRunning {
		return
	}
	for _, s := range m.Steps {
		if s.State != StateComplete {
			return
		}
	}
	m.State = StateComplete
	m.UpdatedAt = time.Now().UTC()
	e.missions[id] = m
}
func (e *Engine) Steps(id string) []Step {
	e.mu.RLock()
	defer e.mu.RUnlock()
	m := e.missions[id]
	out := append([]Step(nil), m.Steps...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (e *Engine) Progress(id string) (int, int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	m := e.missions[id]
	done := 0
	for _, s := range m.Steps {
		if s.State == StateComplete {
			done++
		}
	}
	return done, len(m.Steps)
}
func (e *Engine) Eligible(id string) []Step {
	e.mu.RLock()
	defer e.mu.RUnlock()
	m := e.missions[id]
	out := []Step{}
	for _, s := range m.Steps {
		if s.State == StateComplete {
			continue
		}
		ok := true
		for _, d := range s.DependsOn {
			found := false
			for _, x := range m.Steps {
				if x.ID == d {
					found = true
					if x.State != StateComplete {
						ok = false
					}
				}
			}
			if !found {
				ok = false
			}
		}
		if ok {
			out = append(out, s)
		}
	}
	return out
}
func (e *Engine) Order(id string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	m := e.missions[id]
	out := []string{}
	for _, s := range m.Steps {
		out = append(out, s.ID)
	}
	return out
}
