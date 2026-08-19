// Package http holds delivery handlers that are not part of the
// gRPC-Gateway-generated mux: health checks and (in a fuller build) things
// like SSE streams or feature-flag admin endpoints.
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fauzie/golang-sekeleton/internal/repository"
)

// HealthHandler serves liveness/readiness probes.
type HealthHandler struct {
	repo *repository.Repository
}

// NewHealthHandler builds a HealthHandler.
func NewHealthHandler(repo *repository.Repository) *HealthHandler {
	return &HealthHandler{repo: repo}
}

// Live reports whether the process is up. It never checks dependencies —
// that's Ready's job — so a slow database doesn't get this pod killed by
// the liveness probe.
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "live"})
}

// Ready reports whether the service can currently serve traffic: database,
// cache, and message broker must all answer a ping.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	healthy := true

	if h.repo.DBReadWriter != nil {
		if err := h.repo.DBReadWriter.Ping(ctx); err != nil {
			checks["database"] = err.Error()
			healthy = false
		} else {
			checks["database"] = "ok"
		}
	}
	if h.repo.Cache != nil {
		if err := h.repo.Cache.Ping(ctx); err != nil {
			checks["cache"] = err.Error()
			healthy = false
		} else {
			checks["cache"] = "ok"
		}
	}
	if h.repo.MessagePublisher != nil {
		if err := h.repo.MessagePublisher.Ping(ctx); err != nil {
			checks["message_broker"] = err.Error()
			healthy = false
		} else {
			checks["message_broker"] = "ok"
		}
	}

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]interface{}{"status": readyStatus(healthy), "checks": checks})
}

func readyStatus(healthy bool) string {
	if healthy {
		return "ready"
	}
	return "not_ready"
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
