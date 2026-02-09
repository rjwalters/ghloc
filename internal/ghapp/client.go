package ghapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v68/github"
)

// Client wraps a go-github client for a specific installation.
type Client struct {
	gh             *github.Client
	installationID int64
}

// NewClient creates a GitHub client authenticated as the given installation.
func NewClient(appID, installationID int64, privateKeyPath string) (*Client, error) {
	tr, err := NewInstallationTransport(appID, installationID, privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("installation transport: %w", err)
	}

	httpClient := &http.Client{Transport: tr}
	gh := github.NewClient(httpClient)

	return &Client{
		gh:             gh,
		installationID: installationID,
	}, nil
}

// GetFileContent retrieves the content of a file from a repository.
func (c *Client) GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, string, error) {
	opts := &github.RepositoryContentGetOptions{Ref: ref}
	file, _, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, "", err
	}
	if file == nil {
		return nil, "", fmt.Errorf("path %s is not a file", path)
	}
	content, err := file.GetContent()
	if err != nil {
		return nil, "", err
	}
	return []byte(content), file.GetSHA(), nil
}

// CreateOrUpdateFile creates or updates a file in a repository.
func (c *Client) CreateOrUpdateFile(ctx context.Context, owner, repo, path, message string, content []byte, sha string) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(message),
		Content: content,
	}
	if sha != "" {
		opts.SHA = github.Ptr(sha)
	}

	_, _, err := c.gh.Repositories.CreateFile(ctx, owner, repo, path, opts)
	return err
}

// GitHub returns the underlying go-github client for advanced operations.
func (c *Client) GitHub() *github.Client {
	return c.gh
}
