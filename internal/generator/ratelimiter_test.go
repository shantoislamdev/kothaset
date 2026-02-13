package generator

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRateLimiter_Basic(t *testing.T) {
	rl := NewRateLimiter(1200) // 1 token every 50ms
	defer rl.Close()

	start := time.Now()
	for i := 0; i < 3; i++ {
		if err := rl.Wait(context.Background()); err != nil {
			t.Fatalf("unexpected wait error: %v", err)
		}
	}

	elapsed := time.Since(start)
	if elapsed < 90*time.Millisecond {
		t.Fatalf("rate limiter did not throttle as expected, elapsed=%v", elapsed)
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	rl := NewRateLimiter(1) // 1 token per minute
	defer rl.Close()

	// Consume the immediate token so next wait blocks.
	if err := rl.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected wait error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rl.Wait(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRateLimiter_NoOp(t *testing.T) {
	rl := NewRateLimiter(0)
	defer rl.Close()

	start := time.Now()
	for i := 0; i < 100; i++ {
		if err := rl.Wait(context.Background()); err != nil {
			t.Fatalf("unexpected wait error: %v", err)
		}
	}

	if time.Since(start) > 100*time.Millisecond {
		t.Fatalf("no-op limiter should not throttle")
	}
}

func TestRateLimiter_Close(t *testing.T) {
	rl := NewRateLimiter(1)
	rl.Close()
	rl.Close() // idempotent

	err := rl.Wait(context.Background())
	if !errors.Is(err, errRateLimiterClosed) {
		t.Fatalf("expected errRateLimiterClosed, got %v", err)
	}
}
