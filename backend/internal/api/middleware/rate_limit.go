// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter provides per-IP token bucket rate limiting.
type IPRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

func NewIPRateLimiter(r rate.Limit, burst int, ttl time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*ipLimiter),
		rate:     r,
		burst:    burst,
		ttl:      ttl,
	}
}

func (l *IPRateLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if entry, ok := l.limiters[ip]; ok {
		entry.lastSeen = now
		return entry.limiter
	}

	lim := rate.NewLimiter(l.rate, l.burst)
	l.limiters[ip] = &ipLimiter{limiter: lim, lastSeen: now}
	return lim
}

func (l *IPRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-l.ttl)
	for ip, entry := range l.limiters {
		if entry.lastSeen.Before(cutoff) {
			delete(l.limiters, ip)
		}
	}
}

// RateLimitMiddleware enforces per-IP rate limiting.
func RateLimitMiddleware(l *IPRateLimiter) gin.HandlerFunc {
	// Best-effort cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			l.cleanup()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}

		if !l.get(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
