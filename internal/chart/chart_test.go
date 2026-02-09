package chart

import (
	"strings"
	"testing"
	"time"

	"github.com/rjwalters/ghloc/internal/store"
)

func TestRenderHistoryChart_WithData(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	snapshots := []store.Snapshot{
		{TotalLOC: 100, CreatedAt: base},
		{TotalLOC: 250, CreatedAt: base.Add(7 * 24 * time.Hour)},
		{TotalLOC: 400, CreatedAt: base.Add(14 * 24 * time.Hour)},
		{TotalLOC: 380, CreatedAt: base.Add(21 * 24 * time.Hour)},
		{TotalLOC: 500, CreatedAt: base.Add(28 * 24 * time.Hour)},
	}

	svg := RenderHistoryChart(snapshots)
	svgStr := string(svg)

	if !strings.Contains(svgStr, "<svg") {
		t.Error("output is not SVG")
	}
	if !strings.Contains(svgStr, "Lines of Code") {
		t.Error("chart missing title 'Lines of Code'")
	}
	if !strings.Contains(svgStr, "<path") {
		t.Error("chart missing line path")
	}
	if !strings.Contains(svgStr, "<circle") {
		t.Error("chart missing data point circles")
	}
}

func TestRenderHistoryChart_NoData(t *testing.T) {
	svg := RenderHistoryChart(nil)
	svgStr := string(svg)

	if !strings.Contains(svgStr, "No data yet") {
		t.Error("empty chart should contain 'No data yet'")
	}
}

func TestRenderHistoryChart_SinglePoint(t *testing.T) {
	snapshots := []store.Snapshot{
		{TotalLOC: 500, CreatedAt: time.Now()},
	}

	svg := RenderHistoryChart(snapshots)
	if len(svg) == 0 {
		t.Fatal("expected non-empty SVG")
	}
	if !strings.Contains(string(svg), "<svg") {
		t.Error("output is not SVG")
	}
}

func TestRenderHistoryChart_LargeValues(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	snapshots := []store.Snapshot{
		{TotalLOC: 100000, CreatedAt: base},
		{TotalLOC: 500000, CreatedAt: base.Add(30 * 24 * time.Hour)},
		{TotalLOC: 1200000, CreatedAt: base.Add(60 * 24 * time.Hour)},
	}

	svg := RenderHistoryChart(snapshots)
	svgStr := string(svg)

	if !strings.Contains(svgStr, "<path") {
		t.Error("chart missing line path")
	}
}

func TestNiceAxisTicks(t *testing.T) {
	ticks := niceAxisTicks(0, 1000, 5)
	if len(ticks) == 0 {
		t.Fatal("expected at least 1 tick")
	}
	for i := 1; i < len(ticks); i++ {
		if ticks[i] <= ticks[i-1] {
			t.Errorf("ticks not ascending: %v", ticks)
			break
		}
	}
	t.Logf("ticks for 0-1000: %v", ticks)
}

func TestFormatAxisValue(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{500, "500"},
		{1500, "2k"},
		{50000, "50k"},
		{1500000, "1.5M"},
	}
	for _, tt := range tests {
		got := formatAxisValue(tt.input)
		if got != tt.want {
			t.Errorf("formatAxisValue(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
