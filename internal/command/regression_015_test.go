package command_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/command"
	"gitlab.com/zhangkui/orbit-relay/internal/metrics"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
)

func TestBug015_SendFailureRemainsAnError(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	q := command.NewQueue(repository.New(db), metrics.New()); defer q.Close()
	p := model.CommandPacket{ID: "cmd-15", SatelliteID: "sat-15", WindowID: "win-15", Opcode: "PING"}
	if err := q.Enqueue(context.Background(), p); err != nil { t.Fatal(err) }
	if err := q.Send(context.Background(), p.ID, func([]byte) error { return errors.New("uplink unavailable") }); err == nil {
		t.Fatal("transport failure was returned as success")
	}
}
