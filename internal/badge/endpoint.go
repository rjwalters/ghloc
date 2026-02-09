package badge

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rjwalters/ghloc/internal/store"
)

// shieldsResponse is the shields.io endpoint JSON schema.
type shieldsResponse struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

// EndpointHandler serves shields.io-compatible JSON badge data.
type EndpointHandler struct {
	store store.Store
}

// NewEndpointHandler creates a new badge endpoint handler.
func NewEndpointHandler(s store.Store) *EndpointHandler {
	return &EndpointHandler{store: s}
}

// ServeHTTP handles GET /badge/{owner}/{repo} requests.
func (h *EndpointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	if owner == "" || repo == "" {
		http.Error(w, "owner and repo required", http.StatusBadRequest)
		return
	}

	snap, err := h.store.GetLatest(r.Context(), owner, repo)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := shieldsResponse{
		SchemaVersion: 1,
		Label:         "lines of code",
		Color:         "blue",
	}

	if snap == nil {
		resp.Message = "no data"
		resp.Color = "lightgrey"
	} else {
		resp.Message = FormatLOC(snap.TotalLOC)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=300")
	json.NewEncoder(w).Encode(resp)
}

// FormatLOC formats a LOC count for display (e.g., "12.3k", "1.5M").
func FormatLOC(loc int64) string {
	switch {
	case loc >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(loc)/1_000_000)
	case loc >= 1_000:
		return fmt.Sprintf("%.1fk", float64(loc)/1_000)
	default:
		return fmt.Sprintf("%d", loc)
	}
}
