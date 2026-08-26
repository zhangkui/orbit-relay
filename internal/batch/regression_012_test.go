package batch_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/batch"
	"gitlab.com/zhangkui/orbit-relay/internal/command"
	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
)

func TestBug012_SendFailureFailsWholeBatch(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	q := command.NewQueue(repository.New(db), metrics.New()); defer q.Close()
	p := model.CommandPacket{ID: "cmd-12", SatelliteID: "sat-12", WindowID: "win-12", Opcode: "PING"}
	if err := q.Enqueue(context.Background(), p); err != nil { t.Fatal(err) }
	r := batch.NewRunner(q)
	if _, err := r.Create("batch-12", "win-12", []model.CommandPacket{p}); err != nil { t.Fatal(err) }
	if err := r.Run(context.Background(), "batch-12", func([]byte) error { return errors.New("radio down") }); err == nil { t.Fatal("send failure must fail the batch") }
	b, _ := r.Get("batch-12")
	if b.State != batch.Failed || b.Error == "" { t.Fatalf("batch hid send failure: %#v", b) }
}
