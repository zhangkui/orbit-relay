package window

import (
	"fmt"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/validation"
	"time"
)

var transitions = map[string]map[string]bool{model.WindowPlanned: {model.WindowActive: true, model.WindowCancelled: true}, model.WindowActive: {model.WindowCompleted: true, model.WindowFailed: true, model.WindowCancelled: true}}

func Transition(w *model.ContactWindow, next string, now time.Time) error {
	if !model.ValidWindowState(next) {
		return model.ErrInvalid
	}
	if w.State == next {
		return nil
	}
	if !transitions[next][w.State] {
		return fmt.Errorf("%s -> %s: %w", w.State, next, model.ErrConflict)
	}
	if next == model.WindowActive && !validation.InWindow(now, w.Start, w.End) {
		return model.ErrWindowClosed
	}
	w.State = next
	return nil
}
func CanQueue(w model.ContactWindow, now time.Time) bool {
	return w.State == model.WindowActive && validation.InWindow(now, w.Start, w.End)
}
func Duration(w model.ContactWindow) time.Duration { return w.End.Sub(w.Start) }
