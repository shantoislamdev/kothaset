package generator

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	size := 5
	pool := NewWorkerPool(size)

	if pool.Size() != size {
		t.Errorf("Expected pool size %d, got %d", size, pool.Size())
	}

	var active int32
	var maxActive int32
	var wg sync.WaitGroup

	// Try to launch more workers than pool size
	totalTasks := 20

	for i := 0; i < totalTasks; i++ {
		if err := pool.Acquire(context.Background()); err != nil {
			t.Fatalf("unexpected acquire error: %v", err)
		}
		wg.Add(1)

		go func() {
			defer pool.Release()
			defer wg.Done()

			current := atomic.AddInt32(&active, 1)

			// Update max active safely
			for {
				max := atomic.LoadInt32(&maxActive)
				if current <= max {
					break
				}
				if atomic.CompareAndSwapInt32(&maxActive, max, current) {
					break
				}
			}

			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&active, -1)
		}()
	}

	wg.Wait()

	if maxActive > int32(size) {
		t.Errorf("Max active workers %d exceeded pool size %d", maxActive, size)
	}
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	pool := NewWorkerPool(1)
	if err := pool.Acquire(context.Background()); err != nil {
		t.Fatalf("unexpected acquire error: %v", err)
	}
	defer pool.Release()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() {
		done <- pool.Acquire(ctx)
	}()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Acquire did not return after context cancellation")
	}
}
