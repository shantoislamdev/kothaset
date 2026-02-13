package generator

import "context"

// WorkerPool manages concurrent workers using a semaphore pattern
type WorkerPool struct {
	sem chan struct{}
}

// NewWorkerPool creates a new worker pool with the given concurrency limit
func NewWorkerPool(size int) *WorkerPool {
	if size <= 0 {
		size = 1
	}
	return &WorkerPool{
		sem: make(chan struct{}, size),
	}
}

// Acquire acquires a worker slot (blocks if pool is full).
// Returns when the context is canceled while waiting.
func (p *WorkerPool) Acquire(ctx context.Context) error {
	select {
	case p.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release releases a worker slot
func (p *WorkerPool) Release() {
	<-p.sem
}

// Size returns the pool size
func (p *WorkerPool) Size() int {
	return cap(p.sem)
}

// Active returns the number of active workers
func (p *WorkerPool) Active() int {
	return len(p.sem)
}
