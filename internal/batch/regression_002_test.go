package batch_test

import (
	"path/filepath"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/batch"
	"gitlab.com/zhangkui/orbit-relay/internal/command"
	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
)

func TestBug002_BatchSnapshotDoesNotAliasCallerState(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	q := command.NewQueue(repository.New(db), metrics.New()); defer q.Close()
	packets := []model.CommandPacket{{ID: "low", Priority: 1}, {ID: "high", Priority: 9}}
	if _, err := batch.NewRunner(q).Create("batch-2", "window-2", packets); err != nil { t.Fatal(err) }
	if packets[0].ID != "low" { t.Fatalf("batch sorting changed caller-owned slice: %#v", packets) }
}
