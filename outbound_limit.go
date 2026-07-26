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
	if reqPerMin < 1 {
		reqPerMin = 1
	}
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
	defer l.mu.Unlock()

	for {
		now := time.Now()
		elapsed := now.Sub(l.last).Seconds()
		if elapsed > 0 {
			l.tokens += elapsed * l.rate
			if l.tokens > l.burst {
				l.tokens = l.burst
			}
			l.last = now
		}
		if l.tokens >= 1 {
			l.tokens--
			return
		}
		sleepFor := time.Duration((1-l.tokens)/l.rate*float64(time.Second) + 0.5)
		log.Printf("[rate-limit] outbound throttled for %v", sleepFor.Round(time.Millisecond))
		time.Sleep(sleepFor)
	}
}
