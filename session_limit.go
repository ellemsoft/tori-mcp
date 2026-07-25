package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

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
		sessionID := r.Header.Get("Mcp-Session-Id")
		ua := r.Header.Get("User-Agent")
		org := r.Header.Get("OpenAI-Organization")
		log.Printf("[req] session=%s ua=%s org=%s method=%s", truncS(sessionID, 32), truncS(ua, 80), truncS(org, 40), r.Method)
		if !sl.allow(sessionID, 60) {
			log.Printf("[rate-limit] session %s exceeded limit", truncS(sessionID, 16))
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		recordRequest("mcp")
		next.ServeHTTP(w, r)
	})
}

func truncS(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
