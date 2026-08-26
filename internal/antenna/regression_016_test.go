package antenna

import (
	"testing"
	"time"
)

func TestBug016_HealthyAntennaIsAllocatable(t *testing.T) {
	p := NewPool([]Resource{{StationID: "gs-16", Name: "dish-a", Health: "healthy"}})
	now := time.Now().UTC()
	if !p.Available("gs-16", "dish-a", now, now.Add(time.Minute)) {
		t.Fatal("healthy antenna was rejected by allocation preflight")
	}
}
