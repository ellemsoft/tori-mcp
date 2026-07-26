package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestSessionLimiterAllow(t *testing.T) {
	limiter := &sessionLimiter{sessions: make(map[string]*sessionBucket)}
	if !limiter.allow("session-a", 2) {
		t.Fatal("first request should be allowed")
	}
	if !limiter.allow("session-a", 2) {
		t.Fatal("second request should be allowed")
	}
	if limiter.allow("session-a", 2) {
		t.Fatal("third request should be rate limited")
	}
	if !limiter.allow("session-b", 2) {
		t.Fatal("different session should have a separate bucket")
	}
	if !limiter.allow("", 2) {
		t.Fatal("empty key should be allowed")
	}
}

func TestRequestLimitKey(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want string
	}{
		{
			name: "session id wins",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/mcp", nil)
				r.Header.Set("Mcp-Session-Id", "abc")
				r.Header.Set("X-Forwarded-For", "203.0.113.10")
				return r
			}(),
			want: "session:abc",
		},
		{
			name: "forwarded for fallback",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/mcp", nil)
				r.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.1")
				return r
			}(),
			want: "ip:203.0.113.10",
		},
		{
			name: "remote addr fallback",
			req:  httptest.NewRequest(http.MethodPost, "/mcp", nil),
			want: "ip:192.0.2.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestLimitKey(tt.req); got != tt.want {
				t.Fatalf("requestLimitKey()=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestSessionMiddlewareRateLimitsBySession(t *testing.T) {
	handler := sessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := range sessionLimitPerMinute {
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Mcp-Session-Id", "session-a")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d status=%d, want %d", i+1, rec.Code, http.StatusNoContent)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Mcp-Session-Id", "session-a")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("over-limit status=%d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestSessionMiddlewareRateLimitsWithoutSessionHeader(t *testing.T) {
	handler := sessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := range sessionLimitPerMinute {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/mcp", nil))
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d status=%d, want %d", i+1, rec.Code, http.StatusNoContent)
		}
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("over-limit status=%d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestOutboundLimiterSerializesConcurrentWaiters(t *testing.T) {
	limiter := newOutboundLimiter(1200) // 20 requests/sec, 50ms per token.
	limiter.mu.Lock()
	limiter.tokens = 0
	limiter.burst = 1
	limiter.last = time.Now()
	limiter.mu.Unlock()

	start := time.Now()
	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.wait()
		}()
	}
	wg.Wait()

	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Fatalf("concurrent waits completed too quickly: %v", elapsed)
	}
}
