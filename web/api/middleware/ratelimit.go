package middleware

import (
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	tokens    float64
	lastFill  time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64
	capacity float64
}

func NewRateLimiter(ratePerSec float64, capacity float64) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     ratePerSec,
		capacity: capacity,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[ip]
	if !ok {
		b = &bucket{tokens: rl.capacity, lastFill: time.Now()}
		rl.buckets[ip] = b
	}

	now := time.Now()
	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.lastFill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		now := time.Now()
		for ip, b := range rl.buckets {
			if now.Sub(b.lastFill) > 10*time.Minute {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}
		if !rl.Allow(ip) {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
