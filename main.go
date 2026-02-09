package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/narqo/go-badge"
	"github.com/rjwalters/ghloc/internal/chart"
	"github.com/rjwalters/ghloc/internal/counter"
	"github.com/rjwalters/ghloc/internal/store"

	locbadge "github.com/rjwalters/ghloc/internal/badge"
)

func main() {
	dir := flag.String("dir", ".", "directory to count")
	output := flag.String("output", ".ghloc", "output directory for artifacts")
	flag.Parse()

	log.SetFlags(0)

	// 1. Count LOC
	result, err := counter.Count(*dir)
	if err != nil {
		log.Fatalf("count: %v", err)
	}
	fmt.Printf("Counted %d lines of code across %d files (%d languages)\n",
		result.TotalCode, result.TotalFiles, len(result.Languages))

	// 2. Load existing history
	historyPath := filepath.Join(*output, "history.json")
	history, err := store.LoadHistory(historyPath)
	if err != nil {
		log.Fatalf("load history: %v", err)
	}

	// 3. Append new snapshot
	snap := store.Snapshot{
		TotalLOC:   result.TotalCode,
		TotalFiles: result.TotalFiles,
		CreatedAt:  time.Now().UTC(),
	}
	for _, lang := range result.Languages {
		snap.Languages = append(snap.Languages, store.LanguageRecord{
			Language: lang.Language,
			Lines:    lang.Lines,
			Code:     lang.Code,
			Comments: lang.Comments,
			Blanks:   lang.Blanks,
			Files:    lang.Files,
		})
	}
	history = append(history, snap)

	// 4. Ensure output directory exists
	if err := os.MkdirAll(*output, 0755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	// 5. Write badge SVG
	badgeSVG := locbadge.RenderSVG(locbadge.FormatLOC(result.TotalCode), badge.ColorBlue)
	badgePath := filepath.Join(*output, "badge.svg")
	if err := os.WriteFile(badgePath, badgeSVG, 0644); err != nil {
		log.Fatalf("write badge: %v", err)
	}
	fmt.Printf("Wrote %s\n", badgePath)

	// 6. Write chart SVG
	chartSVG := chart.RenderHistoryChart(history)
	chartPath := filepath.Join(*output, "chart.svg")
	if err := os.WriteFile(chartPath, chartSVG, 0644); err != nil {
		log.Fatalf("write chart: %v", err)
	}
	fmt.Printf("Wrote %s\n", chartPath)

	// 7. Save updated history
	if err := store.SaveHistory(historyPath, history); err != nil {
		log.Fatalf("save history: %v", err)
	}
	fmt.Printf("Wrote %s (%d snapshots)\n", historyPath, len(history))
}
