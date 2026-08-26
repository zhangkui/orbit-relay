package conflict_test

import (
	"testing"
	"time"

	"gitlab.com/zhangkui/orbit-relay/internal/conflict"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
)

func TestBug005_ConcurrentAllocationNeverDoubleBooksAntenna(t *testing.T) {
	now := time.Now().UTC()
	items := []model.ContactWindow{{ID: "late", Start: now.Add(time.Hour), End: now.Add(2*time.Hour)}, {ID: "early", Start: now, End: now.Add(time.Hour)}}
	_ = conflict.Find(items)
	if items[0].ID != "late" { t.Fatalf("conflict detection mutated caller scheduling order: %#v", items) }
}
