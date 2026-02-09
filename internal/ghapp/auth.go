package ghapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v68/github"
)

// NewAppTransport creates an http.RoundTripper that authenticates as a GitHub App.
func NewAppTransport(appID int64, privateKeyPath string) (http.RoundTripper, error) {
	tr, err := ghinstallation.NewAppsTransportKeyFromFile(http.DefaultTransport, appID, privateKeyPath)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

// NewInstallationTransport creates an http.RoundTripper that authenticates as
// a specific installation of the GitHub App.
func NewInstallationTransport(appID int64, installationID int64, privateKeyPath string) (http.RoundTripper, error) {
	tr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, privateKeyPath)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

// InstallationToken returns an access token for the given installation.
func InstallationToken(appID int64, installationID int64, privateKeyPath string) (string, error) {
	tr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, privateKeyPath)
	if err != nil {
		return "", err
	}
	token, err := tr.Token(nil)
	if err != nil {
		return "", err
	}
	return token, nil
}

// FindInstallationID looks up the installation ID for a given repo by
// authenticating as the App (JWT) and querying the GitHub API.
func FindInstallationID(ctx context.Context, appID int64, privateKeyPath, owner, repo string) (int64, error) {
	tr, err := NewAppTransport(appID, privateKeyPath)
	if err != nil {
		return 0, fmt.Errorf("app transport: %w", err)
	}
	client := github.NewClient(&http.Client{Transport: tr})

	install, _, err := client.Apps.FindRepositoryInstallation(ctx, owner, repo)
	if err != nil {
		return 0, fmt.Errorf("find installation for %s/%s: %w", owner, repo, err)
	}
	return install.GetID(), nil
}
