package httpapi

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type loginRateLimitConfig struct {
	MaxAttempts int
	Window      time.Duration
	MaxEntries  int
	Now         func() time.Time
}

func LoginRateLimitConfig(maxAttempts int, window time.Duration, maxEntries int) loginRateLimitConfig {
	return loginRateLimitConfig{MaxAttempts: maxAttempts, Window: window, MaxEntries: maxEntries}
}

type loginAttempt struct {
	count       int
	windowStart time.Time
	lastSeen    time.Time
}

type loginRateLimiter struct {
	mutex       sync.Mutex
	entries     map[string]loginAttempt
	maxAttempts int
	window      time.Duration
	maxEntries  int
	now         func() time.Time
}

func newLoginRateLimiter(cfg loginRateLimitConfig) *loginRateLimiter {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &loginRateLimiter{
		entries:     make(map[string]loginAttempt),
		maxAttempts: cfg.MaxAttempts,
		window:      cfg.Window,
		maxEntries:  cfg.MaxEntries,
		now:         now,
	}
}

func (limiter *loginRateLimiter) allow(key string) (bool, time.Duration) {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	now := limiter.now()
	entry, exists := limiter.entries[key]
	if !exists || now.Sub(entry.windowStart) >= limiter.window {
		if exists {
			delete(limiter.entries, key)
		}
		return true, 0
	}
	if entry.count < limiter.maxAttempts {
		return true, 0
	}
	retryAfter := limiter.window - now.Sub(entry.windowStart)
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	return false, retryAfter
}

func (limiter *loginRateLimiter) failure(key string) {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	now := limiter.now()
	entry, exists := limiter.entries[key]
	if !exists || now.Sub(entry.windowStart) >= limiter.window {
		if !exists && len(limiter.entries) >= limiter.maxEntries {
			limiter.evictOldest()
		}
		limiter.entries[key] = loginAttempt{count: 1, windowStart: now, lastSeen: now}
		return
	}
	entry.count++
	entry.lastSeen = now
	limiter.entries[key] = entry
}

func (limiter *loginRateLimiter) reset(key string) {
	limiter.mutex.Lock()
	delete(limiter.entries, key)
	limiter.mutex.Unlock()
}

func (limiter *loginRateLimiter) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	for key, entry := range limiter.entries {
		if oldestKey == "" || entry.lastSeen.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.lastSeen
		}
	}
	if oldestKey != "" {
		delete(limiter.entries, oldestKey)
	}
}

func loginLimitKey(request *http.Request, username string) string {
	host := request.RemoteAddr
	if parsedHost, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
		host = parsedHost
	}
	return strings.ToLower(strings.TrimSpace(username)) + "\x00" + host
}
