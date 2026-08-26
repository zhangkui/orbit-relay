package service

import (
	"path/filepath"
	"testing"
	"time"

	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/repository"
	"gitlab.com/zhangkui/orbit-relay/internal/store"
	"gitlab.com/zhangkui/orbit-relay/internal/window"
)

func TestBug018_RestorePreservesFailedWindowState(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "relay.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	repo := repository.New(db)
	now := time.Now().UTC()
	failed := model.ContactWindow{ID: "win-18", SatelliteID: "sat-18", StationID: "gs-18", Start: now.Add(time.Hour), End: now.Add(2*time.Hour), State: model.WindowFailed}
	if err := repo.Put("window", failed.ID, failed); err != nil { t.Fatal(err) }
	restored, err := Restore(repo); if err != nil { t.Fatal(err) }
	if got, ok := restored[failed.ID]; !ok || got.State != model.WindowFailed { t.Fatalf("failed window disappeared after restore: %#v", restored) }
	if _, ok := window.Next([]model.ContactWindow{failed}, now); ok { t.Fatal("failed window became schedulable after restore") }
}
