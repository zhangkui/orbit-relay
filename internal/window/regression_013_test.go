package window

import (
	"testing"
	"time"

	"gitlab.com/zhangkui/orbit-relay/internal/model"
)

func TestBug013_PlannedWindowCannotQueueBeforeActivation(t *testing.T) {
	now := time.Now().UTC()
	w := model.ContactWindow{State: model.WindowPlanned, Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)}
	if err := Transition(&w, model.WindowCancelled, now); err != nil {
		t.Fatalf("planned window cancellation was rejected: %v", err)
	}
	if w.State != model.WindowCancelled {
		t.Fatalf("cancellation changed the planned window into %q", w.State)
	}
}
