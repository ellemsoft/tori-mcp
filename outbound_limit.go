package main

import (
	"log"
	"sync"
	"time"
)

type outboundLimiter struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
	rate   float64 // tokens per second
	burst  float64
}

func newOutboundLimiter(reqPerMin int) *outboundLimiter {
	rate := float64(reqPerMin) / 60.0
	return &outboundLimiter{
		tokens: rate * 5, // burst of ~5 seconds worth
		last:   time.Now(),
		rate:   rate,
		burst:  rate * 5,
	}
}

func (l *outboundLimiter) wait() {
	l.mu.Lock()
	now := time.Now()
	elapsed := now.Sub(l.last).Seconds()
	l.tokens += elapsed * l.rate
	if l.tokens > l.burst {
		l.tokens = l.burst
	}
	l.last = now
	if l.tokens < 1 {
		sleepFor := time.Duration((1 - l.tokens) / l.rate * float64(time.Second))
		log.Printf("[rate-limit] outbound throttled for %v", sleepFor.Round(time.Millisecond))
		l.mu.Unlock()
		time.Sleep(sleepFor)
		l.mu.Lock()
		l.tokens = 0
	} else {
		l.tokens--
	}
	l.mu.Unlock()
}
