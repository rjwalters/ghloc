package server

import (
	"log"
	"net/http"

	"github.com/rjwalters/ghloc/internal/badge"
	"github.com/rjwalters/ghloc/internal/chart"
	"github.com/rjwalters/ghloc/internal/config"
	"github.com/rjwalters/ghloc/internal/store"
	"github.com/rjwalters/ghloc/internal/webhook"
)

// New creates and configures the HTTP server with all routes.
func New(cfg *config.Config, s store.Store) *http.Server {
	mux := http.NewServeMux()

	// Webhook endpoint
	wh := webhook.NewHandler(cfg, s)
	mux.Handle("POST /webhook", wh)

	// Badge endpoints
	badgeEndpoint := badge.NewEndpointHandler(s)
	badgeSVG := badge.NewSVGHandler(s)
	mux.Handle("GET /badge/{owner}/{repo}", badgeEndpoint)
	mux.Handle("GET /badge/{owner}/{repo}/svg", badgeSVG)

	// Chart endpoint
	chartHistory := chart.NewHistoryHandler(s)
	mux.Handle("GET /chart/{owner}/{repo}", chartHistory)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Root
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(indexHTML))
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: logMiddleware(mux),
	}

	log.Printf("server: configured on :%s", cfg.Port)
	return srv
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

const indexHTML = `<!DOCTYPE html>
<html>
<head><title>ghloc</title></head>
<body>
<h1>ghloc — GitHub Lines of Code</h1>
<p>A GitHub App that counts lines of code in your repositories.</p>
<h2>Endpoints</h2>
<ul>
<li><code>GET /badge/{owner}/{repo}</code> — shields.io endpoint JSON</li>
<li><code>GET /badge/{owner}/{repo}/svg</code> — SVG badge</li>
<li><code>GET /chart/{owner}/{repo}</code> — LOC history chart</li>
<li><code>POST /webhook</code> — GitHub webhook receiver</li>
</ul>
</body>
</html>`
