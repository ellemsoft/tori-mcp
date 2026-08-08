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
	initStatsAt(path, time.Now())

	// Hourly flush
	go func() {
		for range time.Tick(1 * time.Hour) {
			flushStats()
		}
	}()
}

func initStatsAt(path string, now time.Time) {
	statsPath = path
	stats.Tools = make(map[string]int64)
	stats.Date = now.Format("2006-01-02")

	// Load existing stats
	if data, err := os.ReadFile(path); err == nil {
		var s dailyStats
		if json.Unmarshal(data, &s) == nil && s.Date == stats.Date {
			stats = s
			// Older versions counted every HTTP request under the generic
			// "mcp" key. Keep Total for that history, but don't mix it into
			// the per-tool counters.
			if _, legacy := stats.Tools["mcp"]; legacy {
				delete(stats.Tools, "mcp")
				flushStatsAt(now)
			}
			serverLogf("[stats] loaded %d requests from today", stats.Total)
		} else if s.Date != "" && s.Date != stats.Date {
			stats = s
			flushStatsAt(now)
		}
	}
}

func recordRequest() {
	statsMu.Lock()
	defer statsMu.Unlock()

	rollStatsDateLocked(time.Now())
	stats.Total++
	if stats.Tools == nil {
		stats.Tools = make(map[string]int64)
	}
}

func recordTool(tool string) {
	statsMu.Lock()
	defer statsMu.Unlock()

	rollStatsDateLocked(time.Now())
	if stats.Tools == nil {
		stats.Tools = make(map[string]int64)
	}
	stats.Tools[tool]++
}

func flushStats() {
	flushStatsAt(time.Now())
}

func flushStatsAt(now time.Time) {
	statsMu.Lock()
	defer statsMu.Unlock()

	rolled := rollStatsDateLocked(now)
	if stats.Total == 0 && !rolled {
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

func rollStatsDateLocked(now time.Time) bool {
	today := now.Format("2006-01-02")
	if stats.Date == today {
		if stats.Tools == nil {
			stats.Tools = make(map[string]int64)
		}
		return false
	}
	stats.Date = today
	stats.Total = 0
	stats.Tools = make(map[string]int64)
	return true
}
