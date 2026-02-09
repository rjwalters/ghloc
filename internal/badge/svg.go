package badge

import (
	"bytes"
	"fmt"

	"github.com/narqo/go-badge"
)

// RenderSVG generates an SVG badge with the given message and color.
func RenderSVG(message string, colors ...badge.Color) []byte {
	color := badge.ColorBlue
	if len(colors) > 0 {
		color = colors[0]
	}

	var buf bytes.Buffer
	badge.Render("lines of code", message, color, &buf)
	return buf.Bytes()
}

// FormatLOC formats a LOC count for display (e.g., "12.3k", "1.5M").
func FormatLOC(loc int64) string {
	switch {
	case loc >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(loc)/1_000_000)
	case loc >= 1_000:
		return fmt.Sprintf("%.1fk", float64(loc)/1_000)
	default:
		return fmt.Sprintf("%d", loc)
	}
}
