package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	Burst             int           `config:"burst"`
	Limit             int           `config:"limit"`
	VisitorExpiration time.Duration `config:"ip_expiration"`
}

type RateLimiter struct {
	cfg      RateLimitConfig
	visitors map[string]*visitor
	mu       sync.Mutex
}

func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	limiter := &RateLimiter{
		cfg:      cfg,
		visitors: make(map[string]*visitor),
		mu:       sync.Mutex{},
	}
	go limiter.cleanupVisitors()

	return limiter
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func (l *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			slog.Error("Error parsing remote address", "addr", r.RemoteAddr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		limiter := l.getVisitor(ip)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) getVisitor(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, exists := l.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(l.cfg.Limit), l.cfg.Burst)
		l.visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (l *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		l.mu.Lock()
		for ip, v := range l.visitors {
			if time.Since(v.lastSeen) > l.cfg.VisitorExpiration*time.Minute {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}
