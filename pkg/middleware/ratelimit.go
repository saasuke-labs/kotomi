package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter manages rate limiting for API endpoints
type RateLimiter struct {
	mu          sync.RWMutex
	visitors    map[string]*visitor
	getLimit    int           // requests per minute for GET
	postLimit   int           // requests per minute for POST
	cleanupRate time.Duration // how often to cleanup old visitors
}

// visitor tracks request rate for a single IP/client
type visitor struct {
	limiterGET  *tokenBucket
	limiterPOST *tokenBucket
	lastSeen    time.Time
}

// tokenBucket implements a simple token bucket rate limiter
type tokenBucket struct {
	tokens       float64
	maxTokens    float64
	refillRate   float64 // tokens per second
	lastRefill   time.Time
	mu           sync.Mutex
}

// NewRateLimiter creates a new rate limiter with configuration from environment variables
func NewRateLimiter() *RateLimiter {
	// Get rate limit configuration from environment variables
	getLimit := getEnvInt("RATE_LIMIT_GET", 100)  // default: 100 requests/minute for GET
	postLimit := getEnvInt("RATE_LIMIT_POST", 5)  // default: 5 requests/minute for POST
	
	rl := &RateLimiter{
		visitors:    make(map[string]*visitor),
		getLimit:    getLimit,
		postLimit:   postLimit,
		cleanupRate: 5 * time.Minute,
	}

	// Start cleanup routine to remove old visitors
	go rl.cleanupVisitors()

	return rl
}

// Handler returns middleware that enforces rate limits
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP (check X-Forwarded-For for proxies, fallback to RemoteAddr)
		ip := r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
			// Use only the first (client) IP
			if commaIdx := strings.Index(ip, ","); commaIdx != -1 {
				ip = strings.TrimSpace(ip[:commaIdx])
			}
		}
		if ip == "" {
			ip = r.Header.Get("X-Real-IP")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}

		// Get or create visitor
		v := rl.getVisitor(ip)

		// Check rate limit based on method
		var allowed bool
		var limit int
		
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			allowed = v.limiterPOST.allow()
			limit = rl.postLimit
		} else {
			allowed = v.limiterGET.allow()
			limit = rl.getLimit
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			remaining := int(v.limiterPOST.getTokens())
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		} else {
			remaining := int(v.limiterGET.getTokens())
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		}

		if !allowed {
			// Calculate retry-after in seconds
			var retryAfter int
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
				retryAfter = v.limiterPOST.getRetryAfter()
			} else {
				retryAfter = v.limiterGET.getRetryAfter()
			}
			
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getVisitor returns an existing visitor or creates a new one
func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			limiterGET:  newTokenBucket(float64(rl.getLimit), float64(rl.getLimit)/60.0),
			limiterPOST: newTokenBucket(float64(rl.postLimit), float64(rl.postLimit)/60.0),
			lastSeen:    time.Now(),
		}
		rl.visitors[ip] = v
	} else {
		v.lastSeen = time.Now()
	}

	return v
}

// cleanupVisitors removes visitors that haven't been seen recently
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanupRate)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// newTokenBucket creates a new token bucket
func newTokenBucket(maxTokens, refillRate float64) *tokenBucket {
	return &tokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// allow checks if a request is allowed and consumes a token
func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefill = now

	// Check if we have tokens available
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// getTokens returns the current number of tokens (thread-safe)
func (tb *tokenBucket) getTokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens
}

// getRetryAfter returns seconds until next token is available (thread-safe)
func (tb *tokenBucket) getRetryAfter() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	// Calculate seconds needed to get from current tokens to 1.0
	tokensNeeded := 1.0 - tb.tokens
	if tokensNeeded <= 0 {
		// Should not happen if allow() returned false, but handle it
		return 1
	}
	
	// Calculate time in seconds (round up to be safe)
	retryAfter := int(tokensNeeded/tb.refillRate) + 1
	if retryAfter < 1 {
		retryAfter = 1
	}
	return retryAfter
}

// getEnvInt retrieves an integer from environment variable or returns default
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
