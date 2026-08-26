package telemetry_test

import (
	"context"
	"hash/crc32"
	"path/filepath"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
	"gitlab.com/zhangkui/orbit-relay/internal/telemetry"
)

func TestBug009_CancelledIngestDoesNotAdvanceSequence(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	c := telemetry.NewCollector(repository.New(db), metrics.New()); ctx, cancel := context.WithCancel(context.Background()); cancel()
	raw := []byte("frame-9")
	if err := c.Ingest(ctx, model.TelemetryFrame{ID: "f-9", SatelliteID: "sat-9", StationID: "gs-9", Sequence: 9, Payload: raw, Checksum: crc32.ChecksumIEEE(raw)}); err == nil { t.Fatal("cancelled ingest was accepted") }
	if _, ok := c.LastSequence("sat-9"); ok { t.Fatal("cancelled ingest advanced sequence") }
}
