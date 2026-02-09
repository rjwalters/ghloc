package badge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rjwalters/ghloc/internal/store"
)

// mockStore implements store.Store for testing.
type mockStore struct {
	latest  *store.Snapshot
	history []store.Snapshot
	err     error
}

func (m *mockStore) SaveSnapshot(ctx context.Context, snap *store.Snapshot) error { return m.err }
func (m *mockStore) GetLatest(ctx context.Context, owner, repo string) (*store.Snapshot, error) {
	return m.latest, m.err
}
func (m *mockStore) GetHistory(ctx context.Context, owner, repo string) ([]store.Snapshot, error) {
	return m.history, m.err
}
func (m *mockStore) Close() error { return nil }

func TestEndpointHandler_WithData(t *testing.T) {
	s := &mockStore{
		latest: &store.Snapshot{
			TotalLOC:   12345,
			TotalFiles: 50,
			CreatedAt:  time.Now(),
		},
	}
	handler := NewEndpointHandler(s)

	mux := http.NewServeMux()
	mux.Handle("GET /badge/{owner}/{repo}", handler)

	req := httptest.NewRequest("GET", "/badge/testowner/testrepo", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	var resp struct {
		SchemaVersion int    `json:"schemaVersion"`
		Label         string `json:"label"`
		Message       string `json:"message"`
		Color         string `json:"color"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.SchemaVersion != 1 {
		t.Errorf("schemaVersion: got %d, want 1", resp.SchemaVersion)
	}
	if resp.Message != "12.3k" {
		t.Errorf("message: got %q, want %q", resp.Message, "12.3k")
	}
	if resp.Color != "blue" {
		t.Errorf("color: got %q, want blue", resp.Color)
	}
}

func TestEndpointHandler_NoData(t *testing.T) {
	s := &mockStore{latest: nil}
	handler := NewEndpointHandler(s)

	mux := http.NewServeMux()
	mux.Handle("GET /badge/{owner}/{repo}", handler)

	req := httptest.NewRequest("GET", "/badge/testowner/testrepo", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var resp struct {
		Message string `json:"message"`
		Color   string `json:"color"`
	}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Message != "no data" {
		t.Errorf("message: got %q, want %q", resp.Message, "no data")
	}
	if resp.Color != "lightgrey" {
		t.Errorf("color: got %q, want lightgrey", resp.Color)
	}
}

func TestSVGHandler_WithData(t *testing.T) {
	s := &mockStore{
		latest: &store.Snapshot{
			TotalLOC:   500,
			TotalFiles: 3,
			CreatedAt:  time.Now(),
		},
	}
	handler := NewSVGHandler(s)

	mux := http.NewServeMux()
	mux.Handle("GET /badge/{owner}/{repo}/svg", handler)

	req := httptest.NewRequest("GET", "/badge/testowner/testrepo/svg", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "image/svg+xml" {
		t.Errorf("Content-Type: got %q, want image/svg+xml", ct)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<svg") {
		t.Error("response does not contain <svg")
	}
	if !strings.Contains(body, "500") {
		t.Error("response does not contain the LOC value")
	}
}

func TestFormatLOC(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{42, "42"},
		{999, "999"},
		{1000, "1.0k"},
		{1500, "1.5k"},
		{12345, "12.3k"},
		{999999, "1000.0k"},
		{1000000, "1.0M"},
		{1500000, "1.5M"},
		{12345678, "12.3M"},
	}

	for _, tt := range tests {
		got := FormatLOC(tt.input)
		if got != tt.want {
			t.Errorf("FormatLOC(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
