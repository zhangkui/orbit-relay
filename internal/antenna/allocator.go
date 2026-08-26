package antenna

import (
	"errors"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
	"strings"
	"sync"
	"time"
)

type Resource struct {
	StationID string
	Name      string
	Band      string
	BusyUntil time.Time
	Health    string
}
type Allocation struct {
	WindowID  string
	StationID string
	Antenna   string
	Start     time.Time
	End       time.Time
	Priority  int
}
type Pool struct {
	mu          sync.Mutex
	resources   map[string]Resource
	allocations map[string]Allocation
}

func NewPool(resources []Resource) *Pool {
	p := &Pool{resources: map[string]Resource{}, allocations: map[string]Allocation{}}
	for _, r := range resources {
		p.resources[r.StationID+"/"+r.Name] = r
	}
	return p
}
func (p *Pool) Available(station, antenna string, start, end time.Time) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	r, ok := p.resources[station+"/"+antenna]
	if !ok || r.Health == "offline" {
		return false
	}
	for _, a := range p.allocations {
		if a.StationID == station && a.Antenna == antenna && a.Start.Before(end) && start.Before(a.End) {
			return false
		}
	}
	return true
}
func (p *Pool) Allocate(w model.ContactWindow) (Allocation, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	names := []string{}
	for key, r := range p.resources {
		if r.StationID == w.StationID && r.Health != "offline" {
			parts := strings.Split(key, "/")
			names = append(names, parts[1])
		}
	}
	sort.Strings(names)
	for _, name := range names {
		free := true
		for _, a := range p.allocations {
			if a.StationID == w.StationID && a.Antenna == name && a.Start.Before(w.End) && w.Start.Before(a.End) {
				free = false
				break
			}
		}
		if free {
			a := Allocation{WindowID: w.ID, StationID: w.StationID, Antenna: name, Start: w.Start, End: w.End, Priority: w.Priority}
			p.allocations[w.ID] = a
			return a, nil
		}
	}
	return Allocation{}, errors.New("no antenna available")
}
func (p *Pool) Release(windowID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.allocations, windowID)
}
func (p *Pool) List() []Allocation {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Allocation, 0, len(p.allocations))
	for _, a := range p.allocations {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out
}
