package websocket

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	maxTokens := 100
	refillRate := 10 * time.Millisecond

	rl := NewRateLimiter(maxTokens, refillRate)

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if rl.maxTokens != maxTokens {
		t.Errorf("Expected maxTokens=%d, got %d", maxTokens, rl.maxTokens)
	}

	if rl.tokens != maxTokens {
		t.Errorf("Expected initial tokens=%d, got %d", maxTokens, rl.tokens)
	}

	if rl.refillRate != refillRate {
		t.Errorf("Expected refillRate=%v, got %v", refillRate, rl.refillRate)
	}
}

func TestRateLimiter_Allow_InitialBurst(t *testing.T) {
	maxTokens := 10
	rl := NewRateLimiter(maxTokens, 10*time.Millisecond)

	// Should allow exactly maxTokens requests initially
	for i := 0; i < maxTokens; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should be allowed (initial burst)", i+1)
		}
	}

	// Next request should be denied (tokens exhausted)
	if rl.Allow() {
		t.Error("Request after burst should be denied")
	}
}

func TestRateLimiter_Allow_TokenRefill(t *testing.T) {
	maxTokens := 10
	refillRate := 10 * time.Millisecond
	rl := NewRateLimiter(maxTokens, refillRate)

	// Exhaust all tokens
	for i := 0; i < maxTokens; i++ {
		rl.Allow()
	}

	// Verify tokens are exhausted
	if rl.Allow() {
		t.Error("Should be denied immediately after exhausting tokens")
	}

	// Wait for 1 token to refill
	time.Sleep(refillRate + 1*time.Millisecond)

	// Should now allow 1 request
	if !rl.Allow() {
		t.Error("Should allow 1 request after refill period")
	}

	// Should be denied again
	if rl.Allow() {
		t.Error("Should be denied after consuming refilled token")
	}
}

func TestRateLimiter_Allow_MultipleRefills(t *testing.T) {
	maxTokens := 100
	refillRate := 5 * time.Millisecond
	rl := NewRateLimiter(maxTokens, refillRate)

	// Exhaust all tokens
	for i := 0; i < maxTokens; i++ {
		rl.Allow()
	}

	// Wait for multiple tokens to refill (wait for 10 refill periods = 10 tokens)
	time.Sleep(10*refillRate + 2*time.Millisecond)

	// Should allow approximately 10 requests
	allowedCount := 0
	for i := 0; i < 15; i++ {
		if rl.Allow() {
			allowedCount++
		}
	}

	// Should have allowed around 10 requests (with some tolerance)
	if allowedCount < 8 || allowedCount > 12 {
		t.Errorf("Expected ~10 requests allowed after 10 refill periods, got %d", allowedCount)
	}
}

func TestRateLimiter_Allow_MaxTokensCap(t *testing.T) {
	maxTokens := 10
	refillRate := 1 * time.Millisecond
	rl := NewRateLimiter(maxTokens, refillRate)

	// Consume 5 tokens
	for i := 0; i < 5; i++ {
		rl.Allow()
	}

	// Wait for a very long time (should refill to maxTokens, not beyond)
	time.Sleep(100 * time.Millisecond)

	// Should allow exactly maxTokens, not more
	allowedCount := 0
	for i := 0; i < maxTokens+5; i++ {
		if rl.Allow() {
			allowedCount++
		}
	}

	if allowedCount != maxTokens {
		t.Errorf("Expected exactly %d requests after long wait, got %d", maxTokens, allowedCount)
	}
}

func TestRateLimiter_GetAvailableTokens(t *testing.T) {
	maxTokens := 10
	rl := NewRateLimiter(maxTokens, 10*time.Millisecond)

	// Initial tokens
	if tokens := rl.GetAvailableTokens(); tokens != maxTokens {
		t.Errorf("Expected %d available tokens initially, got %d", maxTokens, tokens)
	}

	// Consume 3 tokens
	for i := 0; i < 3; i++ {
		rl.Allow()
	}

	// Should have 7 tokens remaining
	if tokens := rl.GetAvailableTokens(); tokens != 7 {
		t.Errorf("Expected 7 available tokens, got %d", tokens)
	}
}

func TestRateLimiter_GetAvailableTokens_WithRefill(t *testing.T) {
	maxTokens := 10
	refillRate := 10 * time.Millisecond
	rl := NewRateLimiter(maxTokens, refillRate)

	// Exhaust all tokens
	for i := 0; i < maxTokens; i++ {
		rl.Allow()
	}

	// Verify exhausted
	if tokens := rl.GetAvailableTokens(); tokens != 0 {
		t.Errorf("Expected 0 tokens after exhaustion, got %d", tokens)
	}

	// Wait for refill
	time.Sleep(2*refillRate + 1*time.Millisecond)

	// GetAvailableTokens should trigger refill and return updated count
	tokens := rl.GetAvailableTokens()
	if tokens < 1 || tokens > 3 {
		t.Errorf("Expected 1-3 tokens after refill (with tolerance), got %d", tokens)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	maxTokens := 10
	rl := NewRateLimiter(maxTokens, 10*time.Millisecond)

	// Consume all tokens
	for i := 0; i < maxTokens; i++ {
		rl.Allow()
	}

	// Verify exhausted
	if rl.Allow() {
		t.Error("Should be denied after exhaustion")
	}

	// Reset
	rl.Reset()

	// Should have full tokens again
	if tokens := rl.GetAvailableTokens(); tokens != maxTokens {
		t.Errorf("Expected %d tokens after reset, got %d", maxTokens, tokens)
	}

	// Should allow full burst again
	for i := 0; i < maxTokens; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should be allowed after reset", i+1)
		}
	}
}

func TestRateLimiter_Concurrency(t *testing.T) {
	maxTokens := 1000
	refillRate := 1 * time.Millisecond
	rl := NewRateLimiter(maxTokens, refillRate)

	// Run concurrent requests
	var wg sync.WaitGroup
	goroutines := 10
	requestsPerGoroutine := 200

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				rl.Allow()
			}
		}()
	}

	wg.Wait()

	// Verify no race conditions occurred (test should not panic)
	// The exact count doesn't matter as much as ensuring thread safety
	t.Logf("Concurrent test completed without race conditions")
}

func TestRateLimiter_ConcurrentWithCounting(t *testing.T) {
	maxTokens := 100
	refillRate := 10 * time.Millisecond
	rl := NewRateLimiter(maxTokens, refillRate)

	// Atomic counters for thread-safe counting
	var allowedCount atomic.Int32
	var deniedCount atomic.Int32

	var wg sync.WaitGroup
	goroutines := 10
	requestsPerGoroutine := 100
	totalRequests := goroutines * requestsPerGoroutine

	// Launch concurrent goroutines
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				if rl.Allow() {
					allowedCount.Add(1) // Thread-safe increment
				} else {
					deniedCount.Add(1) // Thread-safe increment
				}
			}
		}()
	}

	wg.Wait()

	// Load final counts
	allowed := allowedCount.Load()
	denied := deniedCount.Load()

	t.Logf("Concurrent counter test: %d allowed, %d denied out of %d total requests",
		allowed, denied, totalRequests)

	// Verify total count is correct
	if int(allowed+denied) != totalRequests {
		t.Errorf("Count mismatch: allowed(%d) + denied(%d) = %d, expected %d",
			allowed, denied, allowed+denied, totalRequests)
	}

	// Should have allowed approximately maxTokens (initial burst)
	// Allow some tolerance for timing variations
	if allowed > int32(maxTokens+10) {
		t.Errorf("Rate limiter too permissive: allowed %d requests (expected ~%d)",
			allowed, maxTokens)
	}

	// Should have denied most requests
	expectedDenied := totalRequests - maxTokens
	if denied < int32(expectedDenied-10) {
		t.Errorf("Rate limiter should have denied more: denied %d (expected ~%d)",
			denied, expectedDenied)
	}

	// Verify atomicity: no requests were lost or double-counted
	if allowed < 0 || denied < 0 {
		t.Errorf("Negative counts detected - atomic operations failed: allowed=%d, denied=%d",
			allowed, denied)
	}
}

func TestRateLimiter_SustainedRate(t *testing.T) {
	// Test sustained rate: 100 messages per second
	maxTokens := 100
	refillRate := 10 * time.Millisecond // 1 token every 10ms = 100 tokens/second
	rl := NewRateLimiter(maxTokens, refillRate)

	// Exhaust initial burst
	for i := 0; i < maxTokens; i++ {
		rl.Allow()
	}

	// Measure sustained rate over 100ms (should get ~10 tokens)
	start := time.Now()
	duration := 100 * time.Millisecond
	allowedCount := 0

	for time.Since(start) < duration {
		if rl.Allow() {
			allowedCount++
		} else {
			// Sleep a bit before retrying
			time.Sleep(1 * time.Millisecond)
		}
	}

	// Should have gotten approximately 10 tokens (100ms / 10ms per token)
	// Allow some tolerance for timing variations
	expectedMin := 8
	expectedMax := 12

	if allowedCount < expectedMin || allowedCount > expectedMax {
		t.Errorf("Expected %d-%d requests in 100ms sustained rate, got %d",
			expectedMin, expectedMax, allowedCount)
	}
}

func TestRateLimiter_HighLoadScenario(t *testing.T) {
	// Simulate the attack scenario from the issue:
	// Malicious client trying to send 10,000 messages/second
	maxTokens := 100
	refillRate := 10 * time.Millisecond // 100 msg/s limit
	rl := NewRateLimiter(maxTokens, refillRate)

	// Try to send 1000 messages as fast as possible
	allowedCount := 0
	deniedCount := 0

	start := time.Now()
	for i := 0; i < 1000; i++ {
		if rl.Allow() {
			allowedCount++
		} else {
			deniedCount++
		}
	}
	elapsed := time.Since(start)

	t.Logf("High load test: %d allowed, %d denied in %v", allowedCount, deniedCount, elapsed)

	// Should have allowed only ~100 (initial burst)
	if allowedCount > 110 {
		t.Errorf("Rate limiter too permissive: allowed %d requests (expected ~100)", allowedCount)
	}

	// Should have denied most requests
	if deniedCount < 890 {
		t.Errorf("Rate limiter should have denied more requests: denied %d (expected ~900)", deniedCount)
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(1000000, 1*time.Nanosecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

func BenchmarkRateLimiter_GetAvailableTokens(b *testing.B) {
	rl := NewRateLimiter(1000000, 1*time.Nanosecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.GetAvailableTokens()
	}
}

func BenchmarkRateLimiter_Concurrent(b *testing.B) {
	rl := NewRateLimiter(1000000, 1*time.Nanosecond)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rl.Allow()
		}
	})
}
