package telemetry_test

import (
	"context"
	"path/filepath"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
	"gitlab.com/zhangkui/orbit-relay/internal/telemetry"
)

func TestBug017_RejectedFrameDoesNotAdvanceSequence(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	c := telemetry.NewCollector(repository.New(db), metrics.New())
	f := model.TelemetryFrame{ID: "frame-17", SatelliteID: "sat-17", StationID: "gs-17", Sequence: 9, Payload: []byte("damaged"), Checksum: 1}
	if err := c.Ingest(context.Background(), f); err == nil { t.Fatal("invalid frame was accepted") }
	if _, ok := c.LastSequence("sat-17"); ok { t.Fatal("rejected frame advanced the receive sequence") }
}
