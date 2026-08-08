package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFlushStatsAtWritesZeroCountAfterDateRollover(t *testing.T) {
	oldStats, oldPath := stats, statsPath
	t.Cleanup(func() {
		stats = oldStats
		statsPath = oldPath
	})

	statsPath = filepath.Join(t.TempDir(), "stats.json")
	stats = dailyStats{
		Date:  "2026-07-26",
		Total: 12,
		Tools: map[string]int64{"mcp": 12},
	}

	flushStatsAt(time.Date(2026, 7, 27, 0, 30, 0, 0, time.UTC))

	data, err := os.ReadFile(statsPath)
	if err != nil {
		t.Fatalf("read stats: %v", err)
	}
	var got dailyStats
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}
	if got.Date != "2026-07-27" {
		t.Fatalf("date = %q, want 2026-07-27", got.Date)
	}
	if got.Total != 0 {
		t.Fatalf("total = %d, want 0", got.Total)
	}
	if len(got.Tools) != 0 {
		t.Fatalf("tools = %v, want empty map", got.Tools)
	}
}

func TestInitStatsAtRewritesStaleStatsFile(t *testing.T) {
	oldStats, oldPath := stats, statsPath
	t.Cleanup(func() {
		stats = oldStats
		statsPath = oldPath
	})

	statsPath = filepath.Join(t.TempDir(), "stats.json")
	old := dailyStats{
		Date:  "2026-07-26",
		Total: 12,
		Tools: map[string]int64{"mcp": 12},
	}
	data, err := json.Marshal(old)
	if err != nil {
		t.Fatalf("marshal stats: %v", err)
	}
	if err := os.WriteFile(statsPath, data, 0644); err != nil {
		t.Fatalf("write stats: %v", err)
	}

	initStatsAt(statsPath, time.Date(2026, 7, 27, 0, 30, 0, 0, time.UTC))

	data, err = os.ReadFile(statsPath)
	if err != nil {
		t.Fatalf("read stats: %v", err)
	}
	var got dailyStats
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}
	if got.Date != "2026-07-27" {
		t.Fatalf("date = %q, want 2026-07-27", got.Date)
	}
	if got.Total != 0 {
		t.Fatalf("total = %d, want 0", got.Total)
	}
	if len(got.Tools) != 0 {
		t.Fatalf("tools = %v, want empty map", got.Tools)
	}
}

func TestRecordRequestAndToolTrackSeparateCounts(t *testing.T) {
	oldStats := stats
	t.Cleanup(func() { stats = oldStats })

	stats = dailyStats{
		Date:  time.Now().Format("2006-01-02"),
		Tools: make(map[string]int64),
	}
	recordRequest()
	recordTool("search")
	recordTool("search")
	recordTool("show")

	if stats.Total != 1 {
		t.Fatalf("total = %d, want 1", stats.Total)
	}
	if got := stats.Tools["search"]; got != 2 {
		t.Fatalf("search count = %d, want 2", got)
	}
	if got := stats.Tools["show"]; got != 1 {
		t.Fatalf("show count = %d, want 1", got)
	}
	if _, ok := stats.Tools["mcp"]; ok {
		t.Fatal("generic mcp counter should not be recorded")
	}
}

func TestInitStatsAtMigratesLegacyMCPCounter(t *testing.T) {
	oldStats, oldPath := stats, statsPath
	t.Cleanup(func() {
		stats = oldStats
		statsPath = oldPath
	})

	statsPath = filepath.Join(t.TempDir(), "stats.json")
	now := time.Now()
	data, err := json.Marshal(dailyStats{
		Date:  now.Format("2006-01-02"),
		Total: 12,
		Tools: map[string]int64{"mcp": 12},
	})
	if err != nil {
		t.Fatalf("marshal stats: %v", err)
	}
	if err := os.WriteFile(statsPath, data, 0644); err != nil {
		t.Fatalf("write stats: %v", err)
	}

	initStatsAt(statsPath, now)

	data, err = os.ReadFile(statsPath)
	if err != nil {
		t.Fatalf("read stats: %v", err)
	}
	var got dailyStats
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}
	if got.Total != 12 {
		t.Fatalf("total = %d, want 12", got.Total)
	}
	if len(got.Tools) != 0 {
		t.Fatalf("tools = %v, want empty map", got.Tools)
	}
}
