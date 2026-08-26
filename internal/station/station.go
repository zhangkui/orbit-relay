package station

import (
	"errors"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"math"
	"sort"
	"sync"
	"time"
)

type Mode string

const (
	Offline  Mode = "offline"
	Standby  Mode = "standby"
	Tracking Mode = "tracking"
	Transmit Mode = "transmit"
	Fault    Mode = "fault"
)

type Antenna struct {
	ID           string
	Name         string
	Band         string
	MinElevation float64
	MaxPower     float64
	Mode         Mode
	LastContact  time.Time
	Temperature  float64
	Health       float64
}
type Station struct {
	ID        string
	Name      string
	Latitude  float64
	Longitude float64
	Altitude  float64
	Antennas  map[string]Antenna
	Enabled   bool
}
type Manager struct {
	mu       sync.RWMutex
	stations map[string]Station
}

func NewManager() *Manager { return &Manager{stations: map[string]Station{}} }
func (m *Manager) Register(s Station) error {
	if s.ID == "" || s.Name == "" || len(s.Antennas) == 0 {
		return model.ErrInvalid
	}
	if math.Abs(s.Latitude) > 90 || math.Abs(s.Longitude) > 180 {
		return model.ErrInvalid
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.stations[s.ID]; ok {
		return model.ErrConflict
	}
	m.stations[s.ID] = s
	return nil
}
func (m *Manager) Get(id string) (Station, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.stations[id]
	if !ok {
		return Station{}, model.ErrNotFound
	}
	s.Antennas = copyAntennas(s.Antennas)
	return s, nil
}
func copyAntennas(in map[string]Antenna) map[string]Antenna {
	out := map[string]Antenna{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
func (m *Manager) SetMode(station, antenna string, mode Mode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.stations[station]
	if !ok {
		return model.ErrNotFound
	}
	a, ok := s.Antennas[antenna]
	if !ok {
		return model.ErrNotFound
	}
	if mode == Transmit && a.Health < 0.5 {
		return errors.New("unhealthy antenna cannot transmit")
	}
	a.Mode = mode
	s.Antennas[antenna] = a
	m.stations[station] = s
	return nil
}
func (m *Manager) Touch(station, antenna string, at time.Time, temp, health float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.stations[station]
	if !ok {
		return model.ErrNotFound
	}
	a, ok := s.Antennas[antenna]
	if !ok {
		return model.ErrNotFound
	}
	a.LastContact = at
	a.Temperature = temp
	a.Health = health
	if temp > 80 || health < 0.2 {
		a.Mode = Fault
	}
	s.Antennas[antenna] = a
	m.stations[station] = s
	return nil
}
func (m *Manager) Available(station, band string, at time.Time) []Antenna {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.stations[station]
	out := []Antenna{}
	for _, a := range s.Antennas {
		if a.Band == band && a.Mode != Offline && a.Mode != Fault && a.Health >= 0.5 && (a.LastContact.IsZero() || at.Sub(a.LastContact) <= 10*time.Minute) {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Health > out[j].Health })
	return out
}
func (m *Manager) Maintenance(station string) []Antenna {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.stations[station]
	out := []Antenna{}
	for _, a := range s.Antennas {
		if a.Mode == Fault || a.Health < 0.5 || a.Temperature > 70 {
			out = append(out, a)
		}
	}
	return out
}
func ElevationLimit(a Antenna) float64 {
	if a.MinElevation < 0 {
		return 0
	}
	if a.MinElevation > 90 {
		return 90
	}
	return a.MinElevation
}
