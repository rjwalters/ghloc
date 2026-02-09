package chart

import (
	"net/http"

	"github.com/rjwalters/ghloc/internal/store"
)

// HistoryHandler serves LOC history charts.
type HistoryHandler struct {
	store store.Store
}

// NewHistoryHandler creates a new chart history handler.
func NewHistoryHandler(s store.Store) *HistoryHandler {
	return &HistoryHandler{store: s}
}

// ServeHTTP handles GET /chart/{owner}/{repo} requests.
func (h *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	if owner == "" || repo == "" {
		http.Error(w, "owner and repo required", http.StatusBadRequest)
		return
	}

	snapshots, err := h.store.GetHistory(r.Context(), owner, repo)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	svg := RenderHistoryChart(snapshots)

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "max-age=300")
	w.Write(svg)
}
