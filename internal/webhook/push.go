package webhook

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/rjwalters/ghloc/internal/badge"
	"github.com/rjwalters/ghloc/internal/chart"
	"github.com/rjwalters/ghloc/internal/counter"
	"github.com/rjwalters/ghloc/internal/ghapp"
	"github.com/rjwalters/ghloc/internal/store"
)

// handlePush processes a push event: clone, count, store, and optionally commit artifacts.
func (h *Handler) handlePush(e *github.PushEvent) error {
	repo := e.GetRepo()
	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	commitSHA := e.GetHeadCommit().GetID()
	ref := e.GetRef()
	installationID := e.GetInstallation().GetID()

	// Look up installation ID if not present (e.g., repo webhook from gh webhook forward)
	if installationID == 0 {
		ctx := context.Background()
		id, err := ghapp.FindInstallationID(ctx, h.config.AppID, h.config.PrivateKeyPath, owner, repoName)
		if err != nil {
			return fmt.Errorf("find installation: %w", err)
		}
		installationID = id
		log.Printf("push: resolved installation ID %d for %s/%s", installationID, owner, repoName)
	}

	// Only process pushes to the default branch
	defaultBranch := repo.GetDefaultBranch()
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	expectedRef := "refs/heads/" + defaultBranch
	if ref != expectedRef {
		log.Printf("push: skipping non-default branch push: %s (expected %s)", ref, expectedRef)
		return nil
	}

	log.Printf("push: processing %s/%s @ %s", owner, repoName, commitSHA[:8])

	// Get installation token for cloning
	token, err := ghapp.InstallationToken(h.config.AppID, installationID, h.config.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	// Shallow clone
	dir, err := counter.CloneRepo(owner, repoName, token)
	if err != nil {
		return fmt.Errorf("clone: %w", err)
	}
	defer os.RemoveAll(dir)

	// Count lines of code
	result, err := counter.Count(dir)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	log.Printf("push: %s/%s — %d lines of code across %d files",
		owner, repoName, result.TotalCode, result.TotalFiles)

	// Build and save snapshot
	snap := &store.Snapshot{
		Owner:      owner,
		Repo:       repoName,
		CommitSHA:  commitSHA,
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

	ctx := context.Background()
	if err := h.store.SaveSnapshot(ctx, snap); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	// Optionally commit badge and chart to repo
	if h.config.CommitArtifacts {
		if err := h.commitArtifacts(ctx, owner, repoName, installationID, result); err != nil {
			log.Printf("push: commit artifacts error (non-fatal): %v", err)
		}
	}

	return nil
}

// commitArtifacts commits badge.svg and chart.svg to the .ghloc/ directory in the repo.
func (h *Handler) commitArtifacts(ctx context.Context, owner, repo string, installationID int64, result *counter.LOCResult) error {
	client, err := ghapp.NewClient(h.config.AppID, installationID, h.config.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	// Generate and commit badge
	badgeSVG := badge.RenderSVG(badge.FormatLOC(result.TotalCode))
	if err := commitFile(ctx, client, owner, repo, ".ghloc/badge.svg", "Update LOC badge [skip ci]", badgeSVG); err != nil {
		return fmt.Errorf("commit badge: %w", err)
	}

	// Generate and commit chart
	history, err := h.store.GetHistory(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("get history: %w", err)
	}
	if len(history) > 1 {
		chartSVG := chart.RenderHistoryChart(history)
		if err := commitFile(ctx, client, owner, repo, ".ghloc/chart.svg", "Update LOC chart [skip ci]", chartSVG); err != nil {
			return fmt.Errorf("commit chart: %w", err)
		}
	}

	return nil
}

// commitFile creates or updates a file in a GitHub repository.
func commitFile(ctx context.Context, client *ghapp.Client, owner, repo, path, message string, content []byte) error {
	// Try to get existing file SHA for update
	_, sha, err := client.GetFileContent(ctx, owner, repo, path, "")
	if err != nil {
		// File doesn't exist yet, that's fine
		sha = ""
	}
	return client.CreateOrUpdateFile(ctx, owner, repo, path, message, content, sha)
}
