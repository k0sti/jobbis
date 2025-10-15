package utils

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_BasicWait(t *testing.T) {
	// Create rate limiter: 2 requests per second, burst of 5
	rl := NewRateLimiter(2.0, 5)
	ctx := context.Background()

	// Should immediately allow first request
	start := time.Now()
	if err := rl.Wait(ctx); err != nil {
		t.Errorf("First wait should succeed: %v", err)
	}
	elapsed := time.Since(start)

	// Should complete almost immediately (with some tolerance)
	if elapsed > 50*time.Millisecond {
		t.Errorf("First wait took too long: %v", elapsed)
	}
}

func TestRateLimiter_Burst(t *testing.T) {
	// Create rate limiter: 10 requests per second, burst of 3
	rl := NewRateLimiter(10.0, 3)
	ctx := context.Background()

	// First 3 requests should succeed immediately (burst capacity)
	start := time.Now()
	for i := 0; i < 3; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Errorf("Request %d should succeed: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	// All 3 should complete quickly (within burst)
	if elapsed > 50*time.Millisecond {
		t.Errorf("Burst requests took too long: %v", elapsed)
	}
}

func TestRateLimiter_RateLimit(t *testing.T) {
	// Create rate limiter: 2 requests per second, burst of 1
	rl := NewRateLimiter(2.0, 1)
	ctx := context.Background()

	// First request should succeed immediately
	start := time.Now()
	if err := rl.Wait(ctx); err != nil {
		t.Errorf("First request should succeed: %v", err)
	}

	// Second request should wait approximately 500ms (1/2 requests per second)
	if err := rl.Wait(ctx); err != nil {
		t.Errorf("Second request should succeed: %v", err)
	}
	elapsed := time.Since(start)

	// Should have waited approximately 500ms (with some tolerance)
	if elapsed < 400*time.Millisecond || elapsed > 600*time.Millisecond {
		t.Errorf("Expected wait time around 500ms, got %v", elapsed)
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	// Create rate limiter: 1 request per second, burst of 1
	rl := NewRateLimiter(1.0, 1)

	// First request succeeds
	if err := rl.Wait(context.Background()); err != nil {
		t.Errorf("First request should succeed: %v", err)
	}

	// Second request with cancelled context should fail
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := rl.Wait(ctx)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestRateLimiter_MultipleRequests(t *testing.T) {
	// Create rate limiter: 5 requests per second, burst of 2
	rl := NewRateLimiter(5.0, 2)
	ctx := context.Background()

	// Make 5 requests and measure total time
	start := time.Now()
	for i := 0; i < 5; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	// First 2 are immediate (burst), next 3 need ~200ms each
	// Expected: ~600ms (3 * 200ms) with some tolerance
	if elapsed < 500*time.Millisecond || elapsed > 800*time.Millisecond {
		t.Errorf("Expected total time around 600ms, got %v", elapsed)
	}
}
