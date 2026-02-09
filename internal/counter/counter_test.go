package counter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCount_GhlocRepo(t *testing.T) {
	// Count the ghloc repo itself — we know it has Go files
	repoRoot := findRepoRoot(t)

	result, err := Count(repoRoot)
	if err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if result.TotalFiles == 0 {
		t.Fatal("expected at least 1 file, got 0")
	}
	if result.TotalCode == 0 {
		t.Fatal("expected TotalCode > 0, got 0")
	}
	if result.TotalLines == 0 {
		t.Fatal("expected TotalLines > 0, got 0")
	}
	if len(result.Languages) == 0 {
		t.Fatal("expected at least 1 language, got 0")
	}

	// We know this repo has Go code
	foundGo := false
	for _, lang := range result.Languages {
		if lang.Language == "Go" {
			foundGo = true
			if lang.Files == 0 {
				t.Error("Go language has 0 files")
			}
			if lang.Code == 0 {
				t.Error("Go language has 0 code lines")
			}
			t.Logf("Go: %d files, %d code, %d comments, %d blanks",
				lang.Files, lang.Code, lang.Comments, lang.Blanks)
		}
	}
	if !foundGo {
		t.Error("expected Go language in results")
	}

	t.Logf("Total: %d files, %d code lines, %d languages",
		result.TotalFiles, result.TotalCode, len(result.Languages))
}

func TestCount_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	result, err := Count(dir)
	if err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if result.TotalFiles != 0 {
		t.Errorf("expected 0 files, got %d", result.TotalFiles)
	}
	if result.TotalCode != 0 {
		t.Errorf("expected 0 code, got %d", result.TotalCode)
	}
}

func TestCount_SingleFile(t *testing.T) {
	dir := t.TempDir()

	// Write a small Go file
	code := `package main

import "fmt"

// main prints hello
func main() {
	fmt.Println("hello")
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Count(dir)
	if err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if result.TotalFiles != 1 {
		t.Errorf("expected 1 file, got %d", result.TotalFiles)
	}
	if result.TotalCode == 0 {
		t.Error("expected code lines > 0")
	}

	if len(result.Languages) != 1 {
		t.Fatalf("expected 1 language, got %d", len(result.Languages))
	}
	if result.Languages[0].Language != "Go" {
		t.Errorf("expected Go, got %s", result.Languages[0].Language)
	}

	t.Logf("Single file: %d code, %d comments, %d blanks",
		result.TotalCode, result.TotalComments, result.TotalBlanks)
}

func TestCount_SkipsGitDir(t *testing.T) {
	dir := t.TempDir()

	// Create a fake .git directory with a Go file inside
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "config.go"), []byte("package git\nfunc init() {}\n"), 0644)

	// Create a real Go file outside .git
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	result, err := Count(dir)
	if err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if result.TotalFiles != 1 {
		t.Errorf("expected 1 file (should skip .git/), got %d", result.TotalFiles)
	}
}

func TestCount_SkipsBinaryFiles(t *testing.T) {
	dir := t.TempDir()

	// Write a binary-looking file
	binary := make([]byte, 256)
	for i := range binary {
		binary[i] = byte(i)
	}
	os.WriteFile(filepath.Join(dir, "data.bin"), binary, 0644)

	// Write a Go file
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	result, err := Count(dir)
	if err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	// Should only count the Go file, not the binary
	for _, lang := range result.Languages {
		if lang.Language == "Go" {
			if lang.Files != 1 {
				t.Errorf("expected 1 Go file, got %d", lang.Files)
			}
		}
	}
}

// findRepoRoot walks up from the test file to find the repo root (contains go.mod).
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (go.mod)")
		}
		dir = parent
	}
}
