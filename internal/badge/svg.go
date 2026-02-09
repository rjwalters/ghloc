package badge

import (
	"bytes"
	"net/http"

	"github.com/narqo/go-badge"
	"github.com/rjwalters/ghloc/internal/store"
)

// SVGHandler serves SVG badges directly.
type SVGHandler struct {
	store store.Store
}

// NewSVGHandler creates a new SVG badge handler.
func NewSVGHandler(s store.Store) *SVGHandler {
	return &SVGHandler{store: s}
}

// ServeHTTP handles GET /badge/{owner}/{repo}/svg requests.
func (h *SVGHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	message := "no data"
	color := badge.ColorLightgrey
	if snap != nil {
		message = FormatLOC(snap.TotalLOC)
		color = badge.ColorBlue
	}

	svg := RenderSVG(message, color)

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "max-age=300")
	w.Write(svg)
}

// RenderSVG generates an SVG badge with the given message and color.
func RenderSVG(message string, colors ...badge.Color) []byte {
	color := badge.ColorBlue
	if len(colors) > 0 {
		color = colors[0]
	}

	var buf bytes.Buffer
	badge.Render("lines of code", message, color, &buf)
	return buf.Bytes()
}
