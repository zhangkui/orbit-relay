package orbit

import (
	"errors"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"math"
	"sort"
	"time"
)

type Position struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
}
type Pass struct {
	ID           string
	SatelliteID  string
	StationID    string
	AOS          time.Time
	LOS          time.Time
	MaxElevation float64
	AzimuthAOS   float64
	AzimuthLOS   float64
	Visible      bool
}
type Calculator struct {
	EarthRadius  float64
	MinElevation float64
}

func NewCalculator() *Calculator { return &Calculator{EarthRadius: 6371, MinElevation: 10} }
func (c *Calculator) GroundDistance(a, b Position) float64 {
	lat1 := a.Latitude * math.Pi / 180
	lat2 := b.Latitude * math.Pi / 180
	dlat := (b.Latitude - a.Latitude) * math.Pi / 180
	dlon := (b.Longitude - a.Longitude) * math.Pi / 180
	x := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	return 2 * c.EarthRadius * math.Atan2(math.Sqrt(x), math.Sqrt(1-x))
}
func (c *Calculator) Elevation(station, subpoint Position) float64 {
	distance := c.GroundDistance(station, subpoint)
	if distance == 0 {
		return 90
	}
	h := subpoint.Altitude
	ratio := distance / (c.EarthRadius + h)
	angle := math.Atan2(h, math.Max(1, distance)) * 180 / math.Pi
	if ratio > math.Pi/2 {
		angle = -angle
	}
	return angle
}
func (c *Calculator) Visible(station, subpoint Position) bool {
	return c.Elevation(station, subpoint) >= c.MinElevation
}
func (c *Calculator) BuildPass(id, sat, station string, start time.Time, duration time.Duration, max float64) Pass {
	return Pass{ID: id, SatelliteID: sat, StationID: station, AOS: start, LOS: start.Add(duration), MaxElevation: max, Visible: max >= c.MinElevation}
}
func (p Pass) Duration() time.Duration   { return p.LOS.Sub(p.AOS) }
func (p Pass) Contains(t time.Time) bool { return !t.Before(p.AOS) && t.Before(p.LOS) }
func filter(items []Pass, fn func(Pass) bool) []Pass {
	out := []Pass{}
	for _, p := range items {
		if fn(p) {
			out = append(out, p)
		}
	}
	return out
}
func FilterVisible(items []Pass) []Pass { return filter(items, func(p Pass) bool { return p.Visible }) }
func SortPasses(items []Pass) []Pass {
	out := append([]Pass(nil), items...)
	sort.Slice(out, func(i, j int) bool { return out[i].AOS.Before(out[j].AOS) })
	return out
}
func MergePasses(items []Pass) []Pass {
	if len(items) < 2 {
		return append([]Pass(nil), items...)
	}
	sorted := SortPasses(items)
	out := []Pass{sorted[0]}
	for _, p := range sorted[1:] {
		last := &out[len(out)-1]
		if p.StationID == last.StationID && !p.AOS.After(last.LOS.Add(time.Second)) {
			if p.LOS.After(last.LOS) {
				last.LOS = p.LOS
			}
			if p.MaxElevation > last.MaxElevation {
				last.MaxElevation = p.MaxElevation
			}
		} else {
			out = append(out, p)
		}
	}
	return out
}
func ValidatePass(p Pass) error {
	if p.ID == "" || p.SatelliteID == "" || p.StationID == "" {
		return errors.New("pass identity required")
	}
	if !p.LOS.After(p.AOS) {
		return errors.New("pass interval invalid")
	}
	if p.MaxElevation < 0 || p.MaxElevation > 90 {
		return errors.New("elevation invalid")
	}
	return nil
}
func WindowFromPass(p Pass, priority int) model.ContactWindow {
	return model.ContactWindow{ID: p.ID, SatelliteID: p.SatelliteID, StationID: p.StationID, Antenna: "auto", Start: p.AOS, End: p.LOS, State: model.WindowPlanned, Priority: priority}
}
