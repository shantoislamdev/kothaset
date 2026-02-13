package generator

import (
	"context"
	"errors"
	"sync"
	"time"
)

var errRateLimiterClosed = errors.New("rate limiter closed")

// RateLimiter enforces requests-per-minute limits for provider calls.
type RateLimiter struct {
	tokens    chan struct{}
	ticker    *time.Ticker
	done      chan struct{}
	closeOnce sync.Once
	disabled  bool
}

// NewRateLimiter creates a new rate limiter for the given requests per minute.
// Values <= 0 disable throttling.
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	if requestsPerMinute <= 0 {
		return &RateLimiter{disabled: true}
	}

	interval := time.Minute / time.Duration(requestsPerMinute)
	if interval <= 0 {
		interval = time.Nanosecond
	}

	rl := &RateLimiter{
		tokens: make(chan struct{}, 1),
		ticker: time.NewTicker(interval),
		done:   make(chan struct{}),
	}

	// Allow one request immediately.
	rl.tokens <- struct{}{}

	go func() {
		for {
			select {
			case <-rl.done:
				return
			case <-rl.ticker.C:
				select {
				case rl.tokens <- struct{}{}:
				default:
				}
			}
		}
	}()

	return rl
}

// Wait blocks until a request token is available or context is canceled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	if r == nil || r.disabled {
		return nil
	}

	// Fast-path to make close behavior deterministic even if a token is buffered.
	select {
	case <-r.done:
		return errRateLimiterClosed
	default:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.done:
		return errRateLimiterClosed
	case <-r.tokens:
		return nil
	}
}

// Close stops the limiter ticker and unblocks pending waiters.
func (r *RateLimiter) Close() {
	if r == nil || r.disabled {
		return
	}

	r.closeOnce.Do(func() {
		r.ticker.Stop()
		close(r.done)
	})
}
