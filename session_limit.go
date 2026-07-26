package main

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const sessionLimitPerMinute = 60

type sessionLimiter struct {
	mu       sync.Mutex
	sessions map[string]*sessionBucket
}

type sessionBucket struct {
	count  int
	window time.Time
}

func newSessionLimiter() *sessionLimiter {
	sl := &sessionLimiter{sessions: make(map[string]*sessionBucket)}
	go sl.cleanup(10 * time.Minute)
	return sl
}

func (sl *sessionLimiter) allow(sessionID string, maxPerMin int) bool {
	if sessionID == "" {
		return true
	}
	sl.mu.Lock()
	defer sl.mu.Unlock()
	now := time.Now()
	b, ok := sl.sessions[sessionID]
	if !ok || now.Sub(b.window) > time.Minute {
		b = &sessionBucket{count: 0, window: now}
		sl.sessions[sessionID] = b
	}
	b.count++
	return b.count <= maxPerMin
}

func (sl *sessionLimiter) cleanup(olderThan time.Duration) {
	for {
		time.Sleep(time.Minute)
		sl.mu.Lock()
		cutoff := time.Now().Add(-olderThan)
		for id, b := range sl.sessions {
			if b.window.Before(cutoff) {
				delete(sl.sessions, id)
			}
		}
		sl.mu.Unlock()
	}
}

func sessionMiddleware(next http.Handler) http.Handler {
	sl := newSessionLimiter()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limitKey := requestLimitKey(r)
		log.Printf("[req] method=%s path=%s", r.Method, r.URL.Path)
		if !sl.allow(limitKey, sessionLimitPerMinute) {
			log.Printf("[rate-limit] client exceeded limit")
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		recordRequest("mcp")
		next.ServeHTTP(w, r)
	})
}

func requestLimitKey(r *http.Request) string {
	if sessionID := strings.TrimSpace(r.Header.Get("Mcp-Session-Id")); sessionID != "" {
		return "session:" + sessionID
	}
	if forwardedFor := firstForwardedFor(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		return "ip:" + forwardedFor
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return "ip:" + realIP
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil && host != "" {
		return "ip:" + host
	}
	if remoteAddr := strings.TrimSpace(r.RemoteAddr); remoteAddr != "" {
		return "remote:" + remoteAddr
	}
	return ""
}

func firstForwardedFor(header string) string {
	if i := strings.IndexByte(header, ','); i >= 0 {
		header = header[:i]
	}
	return strings.TrimSpace(header)
}
