package window

import (
	"testing"
	"time"

	"gitlab.com/zhangkui/orbit-relay/internal/model"
)

func TestBug011_ActiveWindowCanCloseExactlyOnce(t *testing.T) {
	now := time.Now().UTC()
	w := model.ContactWindow{State: model.WindowActive, Start: now.Add(-time.Minute), End: now.Add(time.Minute)}
	if err := Transition(&w, model.WindowCompleted, now); err != nil {
		t.Fatalf("active window should close: %v", err)
	}
	if w.State != model.WindowCompleted {
		t.Fatalf("close retained wrong state %q", w.State)
	}
}
