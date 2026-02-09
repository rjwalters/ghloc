package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rjwalters/ghloc/internal/config"
	"github.com/rjwalters/ghloc/internal/store"
)

func TestHandler_RejectsGet(t *testing.T) {
	h := NewHandler(&config.Config{WebhookSecret: "secret"}, nil)

	req := httptest.NewRequest("GET", "/webhook", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET status: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_RejectsInvalidSignature(t *testing.T) {
	h := NewHandler(&config.Config{WebhookSecret: "secret"}, nil)

	body := `{"action": "push"}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("invalid sig status: got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_AcceptsPushEvent(t *testing.T) {
	secret := "test-secret"
	s := &memStore{}
	h := NewHandler(&config.Config{WebhookSecret: secret}, s)

	// Minimal push event payload
	body := `{
		"ref": "refs/heads/main",
		"head_commit": {"id": "abc123def456"},
		"repository": {
			"name": "test-repo",
			"default_branch": "main",
			"owner": {"login": "testowner"}
		},
		"installation": {"id": 12345}
	}`

	sig := signPayload(secret, []byte(body))

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", "sha256="+sig)
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	// Push events are processed async, handler returns 202
	if w.Code != http.StatusAccepted {
		t.Errorf("push status: got %d, want %d", w.Code, http.StatusAccepted)
	}
}

func TestHandler_AcceptsInstallationEvent(t *testing.T) {
	secret := "test-secret"
	h := NewHandler(&config.Config{WebhookSecret: secret}, nil)

	body := `{
		"action": "created",
		"installation": {"id": 99999, "account": {"login": "testowner"}}
	}`

	sig := signPayload(secret, []byte(body))

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", "sha256="+sig)
	req.Header.Set("X-GitHub-Event", "installation")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("installation status: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_UnknownEventReturnsOK(t *testing.T) {
	secret := "test-secret"
	h := NewHandler(&config.Config{WebhookSecret: secret}, nil)

	body := `{"action": "completed"}`
	sig := signPayload(secret, []byte(body))

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", "sha256="+sig)
	req.Header.Set("X-GitHub-Event", "check_run")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("unknown event status: got %d, want %d", w.Code, http.StatusOK)
	}
}

// signPayload computes the HMAC-SHA256 signature for webhook verification.
func signPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// memStore is a minimal in-memory Store for testing.
type memStore struct {
	snapshots []store.Snapshot
}

func (m *memStore) SaveSnapshot(_ context.Context, snap *store.Snapshot) error {
	m.snapshots = append(m.snapshots, *snap)
	return nil
}
func (m *memStore) GetLatest(_ context.Context, owner, repo string) (*store.Snapshot, error) {
	return nil, nil
}
func (m *memStore) GetHistory(_ context.Context, owner, repo string) ([]store.Snapshot, error) {
	return m.history(owner, repo), nil
}
func (m *memStore) history(owner, repo string) []store.Snapshot {
	var result []store.Snapshot
	for _, s := range m.snapshots {
		if s.Owner == owner && s.Repo == repo {
			result = append(result, s)
		}
	}
	return result
}
func (m *memStore) Close() error { return nil }
