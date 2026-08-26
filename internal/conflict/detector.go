package conflict

import (
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"sort"
	"time"
)

type Conflict struct {
	Left   model.ContactWindow
	Right  model.ContactWindow
	Reason string
}

func Pair(a, b model.ContactWindow) bool {
	if a.StationID != b.StationID {
		return false
	}
	if a.Antenna != "auto" && b.Antenna != "auto" && a.Antenna != b.Antenna {
		return false
	}
	return a.Start.Before(b.End) && b.Start.Before(a.End)
}
func Find(items []model.ContactWindow) []Conflict {
	out := []Conflict{}
	sorted := items
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Start.Before(sorted[j].Start) })
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if !sorted[j].Start.Before(sorted[i].End) {
				break
			}
			if Pair(sorted[i], sorted[j]) {
				out = append(out, Conflict{sorted[i], sorted[j], "station antenna overlap"})
			}
		}
	}
	return out
}
func Resolve(items []model.ContactWindow) []model.ContactWindow {
	out := append([]model.ContactWindow(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].Start.Before(out[j].Start)
		}
		return out[i].Priority > out[j].Priority
	})
	kept := []model.ContactWindow{}
	for _, w := range out {
		ok := true
		for _, x := range kept {
			if Pair(w, x) {
				ok = false
				break
			}
		}
		if ok {
			kept = append(kept, w)
		}
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].Start.Before(kept[j].Start) })
	return kept
}
func Within(w model.ContactWindow, t time.Time) bool { return !t.Before(w.Start) && t.Before(w.End) }
func Margin(a, b model.ContactWindow) time.Duration {
	if a.End.Before(b.Start) {
		return b.Start.Sub(a.End)
	}
	if b.End.Before(a.Start) {
		return a.Start.Sub(b.End)
	}
	return 0
}
