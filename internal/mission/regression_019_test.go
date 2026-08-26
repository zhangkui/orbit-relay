package mission

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/model"
)

func TestBug019_CompletedMissionCannotStartAgain(t *testing.T) {
	e := NewEngine()
	if err := e.Create(Mission{ID: "m-19", Name: "terminal", SatelliteID: "sat-19"}); err != nil { t.Fatal(err) }
	if err := e.AddStep("m-19", Step{ID: "step-19", Name: "capture"}); err != nil { t.Fatal(err) }
	if err := e.Prepare("m-19"); err != nil { t.Fatal(err) }
	if err := e.Start(context.Background(), "m-19"); err != nil { t.Fatal(err) }
	if err := e.CompleteStep("m-19", "step-19", "ok"); err != nil { t.Fatal(err) }
	if err := e.Start(context.Background(), "m-19"); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("completed mission accepted a second start: %v", err)
	}
}
