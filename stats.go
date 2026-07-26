package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type dailyStats struct {
	Date  string           `json:"date"`
	Total int64            `json:"total"`
	Tools map[string]int64 `json:"tools"`
}

var (
	stats     dailyStats
	statsMu   sync.Mutex
	statsPath string
)

func initStats(path string) {
	statsPath = path
	stats.Tools = make(map[string]int64)
	stats.Date = time.Now().Format("2006-01-02")

	// Load existing stats
	if data, err := os.ReadFile(path); err == nil {
		var s dailyStats
		if json.Unmarshal(data, &s) == nil && s.Date == stats.Date {
			stats = s
			serverLogf("[stats] loaded %d requests from today", stats.Total)
		}
	}

	// Hourly flush
	go func() {
		for range time.Tick(1 * time.Hour) {
			flushStats()
		}
	}()
}

func recordRequest(tool string) {
	statsMu.Lock()
	defer statsMu.Unlock()

	today := time.Now().Format("2006-01-02")
	if stats.Date != today {
		stats.Date = today
		stats.Total = 0
		stats.Tools = make(map[string]int64)
	}
	stats.Total++
	stats.Tools[tool]++
}

func flushStats() {
	statsMu.Lock()
	defer statsMu.Unlock()

	if stats.Total == 0 {
		return
	}
	data, err := json.Marshal(stats)
	if err != nil {
		return
	}
	if err := os.WriteFile(statsPath, data, 0644); err != nil {
		serverLogf("[stats] failed to write: %v", err)
	} else {
		serverLogf("[stats] flushed: %d requests today (%v)", stats.Total, stats.Tools)
	}
}
