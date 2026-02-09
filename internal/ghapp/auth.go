package ghapp

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
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
