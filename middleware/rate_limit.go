package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
	IdleTTL            time.Duration
}

type rateLimitEntry struct {
	tokens   float64
	lastSeen time.Time
}

type RateLimiter struct {
	mu          sync.Mutex
	entries     map[string]*rateLimitEntry
	rate        float64
	burst       float64
	idleTTL     time.Duration
	nextCleanup time.Time
}

func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	if cfg.RequestsPerSecond <= 0 {
		cfg.RequestsPerSecond = 5
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 20
	}
	if cfg.IdleTTL <= 0 {
		cfg.IdleTTL = 10 * time.Minute
	}

	return &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		rate:    cfg.RequestsPerSecond,
		burst:   float64(cfg.Burst),
		idleTTL: cfg.IdleTTL,
	}
}

func (l *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, retryAfter := l.allow(c.ClientIP(), time.Now())
		if !allowed {
			c.Header("Retry-After", formatRetryAfter(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

func (l *RateLimiter) allow(ip string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.After(l.nextCleanup) {
		l.removeExpired(now)
		l.nextCleanup = now.Add(l.idleTTL)
	}

	entry, ok := l.entries[ip]
	if !ok {
		l.entries[ip] = &rateLimitEntry{
			tokens:   l.burst - 1,
			lastSeen: now,
		}
		return true, 0
	}

	elapsed := now.Sub(entry.lastSeen).Seconds()
	entry.tokens = math.Min(l.burst, entry.tokens+elapsed*l.rate)
	entry.lastSeen = now
	if entry.tokens >= 1 {
		entry.tokens--
		return true, 0
	}

	return false, time.Duration(math.Ceil((1-entry.tokens)/l.rate*float64(time.Second)))
}

func (l *RateLimiter) removeExpired(now time.Time) {
	for ip, entry := range l.entries {
		if now.Sub(entry.lastSeen) >= l.idleTTL {
			delete(l.entries, ip)
		}
	}
}

func formatRetryAfter(retryAfter time.Duration) string {
	seconds := int(math.Ceil(retryAfter.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	return strconv.Itoa(seconds)
}
