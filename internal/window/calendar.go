package window

import (
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
	"time"
)

type Day struct {
	Date    time.Time
	Windows []model.ContactWindow
	Busy    time.Duration
	Count   int
}

func Calendar(items []model.ContactWindow, from, to time.Time) []Day {
	days := map[string]*Day{}
	for _, w := range items {
		if w.End.Before(from) || w.Start.After(to) {
			continue
		}
		key := w.Start.Format("2006-01-02")
		d := days[key]
		if d == nil {
			date, _ := time.Parse("2006-01-02", key)
			d = &Day{Date: date}
			days[key] = d
		}
		d.Windows = append(d.Windows, w)
		d.Busy += w.End.Sub(w.Start)
		d.Count++
	}
	out := make([]Day, 0, len(days))
	for _, d := range days {
		sort.Slice(d.Windows, func(i, j int) bool { return d.Windows[i].Start.Before(d.Windows[j].Start) })
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.Before(out[j].Date) })
	return out
}
func Next(items []model.ContactWindow, now time.Time) (model.ContactWindow, bool) {
	var best model.ContactWindow
	found := false
	for _, w := range items {
		if w.Start.Before(now) || model.IsTerminalWindow(w.State) {
			continue
		}
		if !found || w.Start.Before(best.Start) {
			best = w
			found = true
		}
	}
	return best, found
}
func Previous(items []model.ContactWindow, now time.Time) (model.ContactWindow, bool) {
	var best model.ContactWindow
	found := false
	for _, w := range items {
		if w.End.After(now) {
			continue
		}
		if !found || w.End.After(best.End) {
			best = w
			found = true
		}
	}
	return best, found
}
func Active(items []model.ContactWindow, now time.Time) []model.ContactWindow {
	out := []model.ContactWindow{}
	for _, w := range items {
		if w.State == model.WindowActive && !now.Before(w.Start) && now.Before(w.End) {
			out = append(out, w)
		}
	}
	return out
}
func GroupByStation(items []model.ContactWindow) map[string][]model.ContactWindow {
	out := map[string][]model.ContactWindow{}
	for _, w := range items {
		out[w.StationID] = append(out[w.StationID], w)
	}
	for k := range out {
		sort.Slice(out[k], func(i, j int) bool { return out[k][i].Start.Before(out[k][j].Start) })
	}
	return out
}
func Gaps(items []model.ContactWindow) []time.Duration {
	sorted := append([]model.ContactWindow(nil), items...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Start.Before(sorted[j].Start) })
	out := []time.Duration{}
	for i := 1; i < len(sorted); i++ {
		if sorted[i].Start.After(sorted[i-1].End) {
			out = append(out, sorted[i].Start.Sub(sorted[i-1].End))
		}
	}
	return out
}
func TotalDuration(items []model.ContactWindow) time.Duration {
	var d time.Duration
	for _, w := range items {
		if w.End.After(w.Start) {
			d += w.End.Sub(w.Start)
		}
	}
	return d
}
