package ingest_test

import (
	"context"
	"testing"
	"time"

	"gitlab.com/zhangkui/orbit-relay/internal/ingest"
)

func TestBug007_QueuedJobReceivesSubmissionContext(t *testing.T) {
	q := ingest.New(1); defer q.Close()
	ctx, cancel := context.WithCancel(context.Background()); cancel()
	done := make(chan error, 1)
	if err := q.Submit(context.Background(), ingest.Job{ID: "job-7", Context: ctx, Done: done, Run: func(got context.Context) error { return got.Err() }}); err != nil { t.Fatal(err) }
	select { case err := <-done: if err == nil { t.Fatal("queued job lost cancellation context") }; case <-time.After(time.Second): t.Fatal("queued job did not complete") }
}
