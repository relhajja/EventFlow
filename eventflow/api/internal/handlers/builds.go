package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/eventflow/api/internal/auth"
	"github.com/eventflow/api/internal/database"
	"github.com/go-chi/chi/v5"
)

type BuildHandler struct {
	buildRepo *database.BuildJobRepository
}

func NewBuildHandler(buildRepo *database.BuildJobRepository) *BuildHandler {
	return &BuildHandler{
		buildRepo: buildRepo,
	}
}

// GetBuildJob handles GET /v1/builds/{id}
func (h *BuildHandler) GetBuildJob(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	buildID := chi.URLParam(r, "id")

	job, err := h.buildRepo.Get(r.Context(), buildID)
	if err != nil {
		respondError(w, http.StatusNotFound, "build job not found", err)
		return
	}

	// Verify ownership
	if job.UserID != claims.UserID {
		respondError(w, http.StatusForbidden, "access denied", nil)
		return
	}

	respondJSON(w, http.StatusOK, job)
}

// GetFunctionBuilds handles GET /v1/functions/{name}/builds
func (h *BuildHandler) GetFunctionBuilds(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	functionName := chi.URLParam(r, "name")

	jobs, err := h.buildRepo.GetByFunction(r.Context(), functionName, claims.Namespace)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get builds", err)
		return
	}

	respondJSON(w, http.StatusOK, jobs)
}

// GetBuildLogs handles GET /v1/builds/{id}/logs
func (h *BuildHandler) GetBuildLogs(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	buildID := chi.URLParam(r, "id")

	job, err := h.buildRepo.Get(r.Context(), buildID)
	if err != nil {
		respondError(w, http.StatusNotFound, "build job not found", err)
		return
	}

	// Verify ownership
	if job.UserID != claims.UserID {
		respondError(w, http.StatusForbidden, "access denied", nil)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(job.Logs))
}

// StreamBuildLogs handles GET /v1/builds/{id}/logs/stream
func (h *BuildHandler) StreamBuildLogs(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	buildID := chi.URLParam(r, "id")

	job, err := h.buildRepo.Get(r.Context(), buildID)
	if err != nil {
		respondError(w, http.StatusNotFound, "build job not found", err)
		return
	}

	// Verify ownership
	if job.UserID != claims.UserID {
		respondError(w, http.StatusForbidden, "access denied", nil)
		return
	}

	// Set headers for streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		respondError(w, http.StatusInternalServerError, "streaming not supported", nil)
		return
	}

	// Send initial logs
	if job.Logs != "" {
		data, _ := json.Marshal(map[string]string{
			"type": "log",
			"data": job.Logs,
		})
		w.Write([]byte("data: " + string(data) + "\n\n"))
		flusher.Flush()
	}

	// Poll for updates
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastStatus := job.Status
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			job, err := h.buildRepo.Get(r.Context(), buildID)
			if err != nil {
				return
			}

			// Send status updates
			if job.Status != lastStatus {
				data, _ := json.Marshal(map[string]string{
					"type":   "status",
					"status": job.Status,
				})
				w.Write([]byte("data: " + string(data) + "\n\n"))
				flusher.Flush()
				lastStatus = job.Status
			}

			// Exit if terminal status
			if job.Status == "success" || job.Status == "failed" {
				if job.Status == "success" && job.Image != "" {
					data, _ := json.Marshal(map[string]string{
						"type":  "complete",
						"image": job.Image,
					})
					w.Write([]byte("data: " + string(data) + "\n\n"))
				}
				flusher.Flush()
				return
			}
		}
	}
}
