package events_test

import (
	"fmt"
	"sync"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/events"
)

func TestBug003_ConcurrentPublishKeepsEventSequence(t *testing.T) {
	b := events.New(); var wg sync.WaitGroup
	for i := 0; i < 32; i++ { wg.Add(1); go func(n int) { defer wg.Done(); b.Publish(events.TypeCommandQueued, fmt.Sprintf("cmd-%d", n), nil) }(i) }
	wg.Wait(); items := b.List("", 1)
	if len(items) != 32 { t.Fatalf("event history lost records: %d", len(items)) }
	for i, e := range items { if e.ID != uint64(i+1) { t.Fatalf("non-monotonic event id at %d: %d", i, e.ID) } }
}
