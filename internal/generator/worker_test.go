package generator

import (
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
		pool.Acquire()
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
