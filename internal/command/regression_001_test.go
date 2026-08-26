package command_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/command"
	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
)

func TestBug001_ConcurrentEnqueueKeepsQueueAndMetricsConsistent(t *testing.T) {
	runtime.GOMAXPROCS(1)
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	q := command.NewQueue(repository.New(db), metrics.New()); defer q.Close()
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ { wg.Add(1); go func(n int) { defer wg.Done(); err := q.Enqueue(context.Background(), model.CommandPacket{ID: fmt.Sprintf("cmd-%d", n), SatelliteID: "sat-1", WindowID: "win-1", Opcode: "PING"}); if err != nil { t.Error(err) } }(i) }
	wg.Wait()
	if got := len(q.List()); got != 24 { t.Fatalf("queue lost concurrent commands: got %d", got) }
}
