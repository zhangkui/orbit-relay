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

func TestBug020_ReportExcludesFailedCommands(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	q := command.NewQueue(repository.New(db), metrics.New()); defer q.Close()
	p := model.CommandPacket{ID: "cmd-20", SatelliteID: "sat-20", WindowID: "win-20", Opcode: "PING"}
	if err := q.Enqueue(context.Background(), p); err != nil { t.Fatal(err) }
	_ = q.Send(context.Background(), p.ID, func([]byte) error { return errors.New("downlink failure") })
	items := q.List()
	if len(items) != 1 || items[0].State != model.CommandFailed || items[0].Error == "" { t.Fatalf("queue snapshot concealed failed delivery: %#v", items) }
}
