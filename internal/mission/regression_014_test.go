package mission

import (
	"errors"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/model"
)

func TestBug014_PauseIsNotIdempotentForTerminalMission(t *testing.T) {
	e := NewEngine()
	if err := e.Create(Mission{ID: "m-14", Name: "recovery", SatelliteID: "sat-14", State: StateRunning}); err != nil { t.Fatal(err) }
	if err := e.Pause("m-14"); err != nil { t.Fatal(err) }
	if err := e.Pause("m-14"); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("second pause must be rejected, got %v", err)
	}
}
