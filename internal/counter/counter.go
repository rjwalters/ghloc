package counter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/boyter/scc/v3/processor"
)

var (
	initOnce sync.Once
	countMu  sync.Mutex // serialize counting for MVP
)

// Count walks the directory tree at dir and counts lines of code using scc.
// The caller should ensure dir is a cloned repository. Thread-safe via mutex.
func Count(dir string) (*LOCResult, error) {
	initOnce.Do(func() {
		processor.ProcessConstants()
	})

	countMu.Lock()
	defer countMu.Unlock()

	langTotals := make(map[string]*LanguageStats)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		// Skip symlinks
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}

		// Detect language
		languages, ext := processor.DetectLanguage(d.Name())
		if len(languages) == 0 {
			return nil // unknown file type, skip
		}

		language := languages[0]
		if len(languages) > 1 {
			language = determineLanguage(d.Name(), languages, content)
		}

		// Create and count a FileJob
		job := &processor.FileJob{
			Filename:  d.Name(),
			Extension: ext,
			Location:  path,
			Language:  language,
			Content:   content,
			Bytes:     int64(len(content)),
		}

		processor.CountStats(job)

		if job.Binary {
			return nil
		}

		// Accumulate per-language stats
		stats, ok := langTotals[language]
		if !ok {
			stats = &LanguageStats{Language: language}
			langTotals[language] = stats
		}
		stats.Lines += job.Lines
		stats.Code += job.Code
		stats.Comments += job.Comment
		stats.Blanks += job.Blank
		stats.Files++

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk dir: %w", err)
	}

	result := &LOCResult{}
	for _, stats := range langTotals {
		result.Languages = append(result.Languages, *stats)
		result.TotalLines += stats.Lines
		result.TotalCode += stats.Code
		result.TotalComments += stats.Comments
		result.TotalBlanks += stats.Blanks
		result.TotalFiles += stats.Files
	}

	return result, nil
}

// determineLanguage tries to pick the best language when multiple are possible.
func determineLanguage(filename string, languages []string, content []byte) string {
	// Use file extension heuristics
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".h":
		// Could be C or C++; default to C
		return "C Header"
	case ".m":
		return "Objective-C"
	}

	// For other ambiguous cases, try scc's built-in determination
	// Pass a reasonable chunk of content for keyword analysis
	sample := content
	if len(sample) > 20000 {
		sample = sample[:20000]
	}

	return processor.DetermineLanguage(filename, languages[0], languages, sample)
}
