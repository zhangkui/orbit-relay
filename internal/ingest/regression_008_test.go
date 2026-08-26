package ingest_test

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/zhangkui/orbit-relay/internal/ingest"
)

func TestBug008_RetryStopsBeforeCancelledAttempt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background()); cancel(); called := false
	err := ingest.Retry(ctx, 3, func() error { called = true; return errors.New("uplink down") })
	if err == nil || called { t.Fatalf("cancelled retry started an attempt: err=%v called=%t", err, called) }
}
