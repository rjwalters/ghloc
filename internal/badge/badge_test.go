package badge

import (
	"strings"
	"testing"
)

func TestRenderSVG(t *testing.T) {
	svg := RenderSVG("12.3k")
	svgStr := string(svg)

	if !strings.Contains(svgStr, "<svg") {
		t.Error("output is not SVG")
	}
	if !strings.Contains(svgStr, "lines of code") {
		t.Error("SVG missing label")
	}
	if !strings.Contains(svgStr, "12.3k") {
		t.Error("SVG missing message")
	}
}

func TestFormatLOC(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{42, "42"},
		{999, "999"},
		{1000, "1.0k"},
		{1500, "1.5k"},
		{12345, "12.3k"},
		{999999, "1000.0k"},
		{1000000, "1.0M"},
		{1500000, "1.5M"},
		{12345678, "12.3M"},
	}

	for _, tt := range tests {
		got := FormatLOC(tt.input)
		if got != tt.want {
			t.Errorf("FormatLOC(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
