package cmd

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitForRetryHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	started := time.Now()
	err := waitForRetry(ctx, time.Minute)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForRetry error = %v, want context.Canceled", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("waitForRetry took %s after cancellation", elapsed)
	}
}
