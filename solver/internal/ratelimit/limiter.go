package ratelimit

import (
	"sync"
	"time"
	"golang.org/x/time/rate"
)

type Limiter struct {
	mu      sync.RWMutex
	buckets map[string]*rateLimiterEntry
	rate    rate.Limit
	burst   int
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func New(rps float64, burst int) *Limiter {
	l := &Limiter{buckets: make(map[string]*rateLimiterEntry), rate: rate.Limit(rps), burst: burst}
	go l.cleanup()
	return l
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	entry, ok := l.buckets[key]
	if !ok {
		entry = &rateLimiterEntry{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.buckets[key] = entry
	}
	entry.lastSeen = time.Now()
	l.mu.Unlock()
	return entry.limiter.Allow()
}

func (l *Limiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		l.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for key, entry := range l.buckets {
			if entry.lastSeen.Before(cutoff) { delete(l.buckets, key) }
		}
		l.mu.Unlock()
	}
}
