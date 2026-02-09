package counter

import (
	"fmt"
	"os"
	"os/exec"
)

// CloneRepo performs a shallow clone of a GitHub repository into a temporary directory.
// It uses the provided token for authentication.
// Returns the path to the cloned directory; caller is responsible for cleanup.
func CloneRepo(owner, repo, token string) (string, error) {
	dir, err := os.MkdirTemp("", "ghloc-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	cloneURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)
	cmd := exec.Command("git", "clone", "--depth=1", "--single-branch", cloneURL, dir)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("git clone: %s: %w", string(output), err)
	}

	return dir, nil
}
