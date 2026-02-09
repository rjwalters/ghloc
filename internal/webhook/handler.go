package webhook

import (
	"log"
	"net/http"

	"github.com/google/go-github/v68/github"
	"github.com/rjwalters/ghloc/internal/config"
	"github.com/rjwalters/ghloc/internal/store"
)

// Handler handles incoming GitHub webhook events.
type Handler struct {
	config *config.Config
	store  store.Store
}

// NewHandler creates a new webhook handler.
func NewHandler(cfg *config.Config, s store.Store) *Handler {
	return &Handler{
		config: cfg,
		store:  s,
	}
}

// ServeHTTP handles POST /webhook requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := github.ValidatePayload(r, []byte(h.config.WebhookSecret))
	if err != nil {
		log.Printf("webhook: invalid signature: %v", err)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("webhook: parse error: %v", err)
		http.Error(w, "parse error", http.StatusBadRequest)
		return
	}

	switch e := event.(type) {
	case *github.PushEvent:
		// Process push events asynchronously
		go func() {
			if err := h.handlePush(e); err != nil {
				log.Printf("webhook: push handler error: %v", err)
			}
		}()
		w.WriteHeader(http.StatusAccepted)

	case *github.InstallationEvent:
		log.Printf("webhook: installation event: action=%s, id=%d",
			e.GetAction(), e.GetInstallation().GetID())
		w.WriteHeader(http.StatusOK)

	default:
		log.Printf("webhook: unhandled event type: %s", github.WebHookType(r))
		w.WriteHeader(http.StatusOK)
	}
}
