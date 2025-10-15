package utils

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, maxBurst int) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(maxBurst),
		maxTokens:  float64(maxBurst),
		refillRate: time.Duration(float64(time.Second) / requestsPerSecond),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		rl.mu.Lock()
		rl.refillTokens()

		if rl.tokens >= 1.0 {
			rl.tokens -= 1.0
			rl.mu.Unlock()
			return nil
		}

		waitTime := rl.refillRate
		rl.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			continue
		}
	}
}

func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	timePassed := now.Sub(rl.lastRefill)
	tokensToAdd := float64(timePassed) / float64(rl.refillRate)

	rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
	rl.lastRefill = now
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
