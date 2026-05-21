package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitOptions настраивает token bucket (запросов в минуту + burst).
type RateLimitOptions struct {
	Enabled           bool
	RequestsPerMinute int
	Burst             int
}

// RateLimitSet — общий лимит API и ужесточённый для /auth (brute-force).
type RateLimitSet struct {
	API  RateLimitOptions
	Auth RateLimitOptions
}

// RateLimit возвращает middleware: лимит по IP (ClientIP), при превышении — 429.
func RateLimit(opts RateLimitOptions) gin.HandlerFunc {
	if !opts.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	store := newLimiterStore(opts)

	return func(c *gin.Context) {
		if !store.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}
		c.Next()
	}
}

type limiterStore struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	limit    rate.Limit
	burst    int
}

func newLimiterStore(opts RateLimitOptions) *limiterStore {
	rpm := opts.RequestsPerMinute
	if rpm < 1 {
		rpm = 1
	}
	burst := opts.Burst
	if burst < 1 {
		burst = 1
	}
	return &limiterStore{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(float64(rpm) / 60.0),
		burst:    burst,
	}
}

func (s *limiterStore) allow(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}

	s.mu.Lock()
	lim, ok := s.limiters[ip]
	if !ok {
		lim = rate.NewLimiter(s.limit, s.burst)
		s.limiters[ip] = lim
	}
	s.mu.Unlock()

	return lim.Allow()
}
