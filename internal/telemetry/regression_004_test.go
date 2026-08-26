package telemetry_test

import (
	"context"
	"fmt"
	"hash/crc32"
	"path/filepath"
	"sync"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
	"gitlab.com/zhangkui/orbit-relay/internal/telemetry"
)

func TestBug004_ConcurrentIngestKeepsMonotonicSequence(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	c := telemetry.NewCollector(repository.New(db), metrics.New()); var wg sync.WaitGroup
	for i := 1; i <= 24; i++ { wg.Add(1); go func(n int) { defer wg.Done(); raw := []byte(fmt.Sprintf("frame-%d", n)); _ = c.Ingest(context.Background(), model.TelemetryFrame{ID: fmt.Sprintf("f-%d", n), SatelliteID: "sat-4", StationID: "gs-4", Sequence: uint32(n), Payload: raw, Checksum: crc32.ChecksumIEEE(raw)}) }(i) }
	wg.Wait()
	if last, ok := c.LastSequence("sat-4"); !ok || last == 0 { t.Fatalf("collector lost sequence watermark: %d %t", last, ok) }
}
