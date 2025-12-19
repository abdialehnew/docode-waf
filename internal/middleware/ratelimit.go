package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"golang.org/x/time/rate"
)
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}
func NewRateLimiter(requestsPerSecond int, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
		// Clean up old limiters periodically
		go func() {
			time.Sleep(10 * time.Minute)
			rl.mu.Lock()
			delete(rl.limiters, ip)
			rl.mu.Unlock()
		}()
	return limiter
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			// Add to context for logging
			ctx := context.WithValue(r.Context(), "blocked", true)
			ctx = context.WithValue(ctx, "block_reason", "rate_limit")
			r = r.WithContext(ctx)
			return
		}
		next.ServeHTTP(w, r)
	})
func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	return ip
