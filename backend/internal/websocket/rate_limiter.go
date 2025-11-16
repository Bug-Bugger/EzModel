package websocket

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket algorithm for rate limiting
type RateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the specified configuration
// maxTokens: maximum number of tokens (burst capacity)
// refillRate: duration between token refills
func NewRateLimiter(maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// refill updates the token count based on elapsed time
// MUST be called with r.mu lock held
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed / r.refillRate)

	if tokensToAdd > 0 {
		r.tokens = min(r.tokens+tokensToAdd, r.maxTokens)
		r.lastRefill = now
	}
}

// Allow checks if an action is allowed based on available tokens
// Returns true if a token is available and consumed, false otherwise
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens based on elapsed time
	r.refill()

	// Check if a token is available
	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

// GetAvailableTokens returns the current number of available tokens after refilling
func (r *RateLimiter) GetAvailableTokens() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens to get accurate current count
	r.refill()

	return r.tokens
}

// Reset resets the rate limiter to its initial state
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens = r.maxTokens
	r.lastRefill = time.Now()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
