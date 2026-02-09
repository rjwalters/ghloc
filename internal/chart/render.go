package chart

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rjwalters/ghloc/internal/store"
)

// RenderHistoryChart generates a star-history-style SVG line chart showing LOC over time.
func RenderHistoryChart(snapshots []store.Snapshot) []byte {
	if len(snapshots) == 0 {
		return []byte(emptySVG())
	}

	// Chart dimensions
	const (
		width       = 800
		height      = 400
		marginTop   = 40
		marginRight = 40
		marginBot   = 60
		marginLeft  = 80
		plotW       = width - marginLeft - marginRight
		plotH       = height - marginTop - marginBot
	)

	// Extract data
	times := make([]time.Time, len(snapshots))
	values := make([]float64, len(snapshots))
	var minVal, maxVal float64
	for i, s := range snapshots {
		times[i] = s.CreatedAt
		values[i] = float64(s.TotalLOC)
		if i == 0 || values[i] < minVal {
			minVal = values[i]
		}
		if i == 0 || values[i] > maxVal {
			maxVal = values[i]
		}
	}

	// Add 10% padding to Y axis
	yRange := maxVal - minVal
	if yRange == 0 {
		yRange = maxVal * 0.1
		if yRange == 0 {
			yRange = 100
		}
	}
	yMin := math.Max(0, minVal-yRange*0.1)
	yMax := maxVal + yRange*0.1

	// Time range
	tMin := times[0]
	tMax := times[len(times)-1]
	tRange := tMax.Sub(tMin).Seconds()
	if tRange == 0 {
		tRange = 86400 // 1 day minimum
	}

	// Map data to pixel coordinates
	xCoords := make([]float64, len(snapshots))
	yCoords := make([]float64, len(snapshots))
	for i := range snapshots {
		xCoords[i] = marginLeft + (times[i].Sub(tMin).Seconds()/tRange)*plotW
		yCoords[i] = marginTop + plotH - ((values[i]-yMin)/(yMax-yMin))*plotH
	}

	// Build SVG
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`, width, height, width, height))
	sb.WriteString("\n")

	// Background
	sb.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="white"/>`, width, height))
	sb.WriteString("\n")

	// Gradient definition for area fill
	sb.WriteString(`<defs>`)
	sb.WriteString(`<linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">`)
	sb.WriteString(`<stop offset="0%" stop-color="#4A90D9" stop-opacity="0.3"/>`)
	sb.WriteString(`<stop offset="100%" stop-color="#4A90D9" stop-opacity="0.05"/>`)
	sb.WriteString(`</linearGradient>`)
	sb.WriteString(`</defs>`)
	sb.WriteString("\n")

	// Y-axis grid lines and labels
	yTicks := niceAxisTicks(yMin, yMax, 5)
	for _, tick := range yTicks {
		y := marginTop + plotH - ((tick-yMin)/(yMax-yMin))*plotH
		sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%.1f" x2="%d" y2="%.1f" stroke="#E5E5E5" stroke-width="1"/>`, marginLeft, y, width-marginRight, y))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%.1f" text-anchor="end" font-family="system-ui, sans-serif" font-size="12" fill="#666">%s</text>`, marginLeft-8, y+4, formatAxisValue(tick)))
		sb.WriteString("\n")
	}

	// X-axis date labels
	xTicks := dateAxisTicks(tMin, tMax, 6)
	for _, t := range xTicks {
		x := marginLeft + (t.Sub(tMin).Seconds()/tRange)*plotW
		sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%d" text-anchor="middle" font-family="system-ui, sans-serif" font-size="12" fill="#666">%s</text>`, x, height-marginBot+20, t.Format("Jan 2006")))
		sb.WriteString("\n")
	}

	// Area fill path
	if len(xCoords) > 1 {
		sb.WriteString(`<path d="`)
		sb.WriteString(fmt.Sprintf("M%.1f,%.1f", xCoords[0], float64(marginTop+plotH)))
		for i := range xCoords {
			sb.WriteString(fmt.Sprintf(" L%.1f,%.1f", xCoords[i], yCoords[i]))
		}
		sb.WriteString(fmt.Sprintf(" L%.1f,%.1f", xCoords[len(xCoords)-1], float64(marginTop+plotH)))
		sb.WriteString(`Z" fill="url(#areaGrad)"/>`)
		sb.WriteString("\n")
	}

	// Line path
	sb.WriteString(`<path d="`)
	for i := range xCoords {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("M%.1f,%.1f", xCoords[i], yCoords[i]))
		} else {
			sb.WriteString(fmt.Sprintf(" L%.1f,%.1f", xCoords[i], yCoords[i]))
		}
	}
	sb.WriteString(`" fill="none" stroke="#4A90D9" stroke-width="2.5" stroke-linejoin="round" stroke-linecap="round"/>`)
	sb.WriteString("\n")

	// Data points
	for i := range xCoords {
		sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="3.5" fill="white" stroke="#4A90D9" stroke-width="2"/>`, xCoords[i], yCoords[i]))
		sb.WriteString("\n")
	}

	// Title
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="24" font-family="system-ui, sans-serif" font-size="16" font-weight="600" fill="#333">Lines of Code</text>`, marginLeft))
	sb.WriteString("\n")

	// Axes
	sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#CCC" stroke-width="1"/>`, marginLeft, marginTop, marginLeft, marginTop+plotH))
	sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#CCC" stroke-width="1"/>`, marginLeft, marginTop+plotH, width-marginRight, marginTop+plotH))

	sb.WriteString("\n</svg>")
	return []byte(sb.String())
}

// niceAxisTicks generates clean tick values for a numeric axis.
func niceAxisTicks(min, max float64, count int) []float64 {
	rawStep := (max - min) / float64(count)
	mag := math.Pow(10, math.Floor(math.Log10(rawStep)))
	normalized := rawStep / mag

	var niceStep float64
	switch {
	case normalized <= 1.5:
		niceStep = 1 * mag
	case normalized <= 3:
		niceStep = 2 * mag
	case normalized <= 7:
		niceStep = 5 * mag
	default:
		niceStep = 10 * mag
	}

	start := math.Ceil(min/niceStep) * niceStep
	var ticks []float64
	for v := start; v <= max; v += niceStep {
		ticks = append(ticks, v)
	}
	return ticks
}

// dateAxisTicks generates evenly spaced date ticks.
func dateAxisTicks(min, max time.Time, count int) []time.Time {
	duration := max.Sub(min)
	if duration <= 0 {
		return []time.Time{min}
	}

	step := duration / time.Duration(count)
	var ticks []time.Time
	for i := 0; i <= count; i++ {
		ticks = append(ticks, min.Add(step*time.Duration(i)))
	}
	return ticks
}

// formatAxisValue formats a number for axis labels.
func formatAxisValue(v float64) string {
	switch {
	case v >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.0fk", v/1_000)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

func emptySVG() string {
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 400" width="800" height="400">
<rect width="800" height="400" fill="white"/>
<text x="400" y="200" text-anchor="middle" font-family="system-ui, sans-serif" font-size="16" fill="#999">No data yet</text>
</svg>`
}
