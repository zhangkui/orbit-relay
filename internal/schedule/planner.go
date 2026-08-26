package schedule

import (
	"context"
	"gitlab.com/zhangkui/orbit-relay/internal/antenna"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/orbit"
	"gitlab.com/zhangkui/orbit-relay/internal/window"
	"sort"
	"time"
)

type Planner struct {
	calc *orbit.Calculator
	pool *antenna.Pool
}

func NewPlanner(pool *antenna.Pool) *Planner {
	return &Planner{calc: orbit.NewCalculator(), pool: pool}
}
func (p *Planner) BuildPasses(sat model.Satellite, station model.GroundStation, start time.Time, count int) []orbit.Pass {
	out := []orbit.Pass{}
	for i := 0; i < count; i++ {
		t := start.Add(time.Duration(i) * 95 * time.Minute)
		e := 45 + float64((sat.NORAD+i)%30)
		out = append(out, p.calc.BuildPass(sat.ID+"-"+station.ID+"-"+t.Format("1504"), sat.ID, station.ID, t, 10*time.Minute, e))
	}
	return orbit.SortPasses(out)
}
func (p *Planner) Plan(ctx context.Context, passes []orbit.Pass, priority int) ([]model.ContactWindow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	windows := []model.ContactWindow{}
	for _, pass := range orbit.MergePasses(passes) {
		if err := orbit.ValidatePass(pass); err != nil {
			return nil, err
		}
		w := orbit.WindowFromPass(pass, priority)
		a, err := p.pool.Allocate(w)
		if err != nil {
			return nil, err
		}
		w.Antenna = a.Antenna
		windows = append(windows, w)
	}
	sort.Slice(windows, func(i, j int) bool { return windows[i].Start.Before(windows[j].Start) })
	return windows, nil
}
func FilterByDay(items []model.ContactWindow, day time.Time) []model.ContactWindow {
	out := []model.ContactWindow{}
	for _, w := range items {
		y, m, d := w.Start.Date()
		if y == day.Year() && m == day.Month() && d == day.Day() {
			out = append(out, w)
		}
	}
	return out
}
func BusyMinutes(items []model.ContactWindow) int {
	total := 0
	for _, w := range items {
		total += int(w.End.Sub(w.Start) / time.Minute)
	}
	return total
}
func ValidateNoOverlap(items []model.ContactWindow) bool {
	sorted := window.SortWindows(items)
	for i := 1; i < len(sorted); i++ {
		if window.Overlap(sorted[i-1], sorted[i]) {
			return false
		}
	}
	return true
}
