package auth

import (
	"sync"
	"time"
)

type limiterEntry struct {
	failures     int
	blockedUntil time.Time
}

// LoginLimiter throttles repeated login failures in a single service instance.
type LoginLimiter struct {
	mu          sync.Mutex
	entries     map[string]limiterEntry
	maxFailures int
	blockFor    time.Duration
}

func NewLoginLimiter(maxFailures int, blockFor time.Duration) *LoginLimiter {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if blockFor <= 0 {
		blockFor = 5 * time.Minute
	}
	return &LoginLimiter{
		entries:     make(map[string]limiterEntry),
		maxFailures: maxFailures,
		blockFor:    blockFor,
	}
}

func (l *LoginLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.entries[key]
	if !ok {
		return true
	}
	if now.UTC().Before(entry.blockedUntil) {
		return false
	}
	if !entry.blockedUntil.IsZero() && !now.UTC().Before(entry.blockedUntil) {
		delete(l.entries, key)
	}
	return true
}

func (l *LoginLimiter) RegisterFailure(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := l.entries[key]
	entry.failures++
	if entry.failures >= l.maxFailures {
		entry.blockedUntil = now.UTC().Add(l.blockFor)
		entry.failures = 0
	}
	l.entries[key] = entry
}

func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}
