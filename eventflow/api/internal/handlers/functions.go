package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eventflow/api/internal/database"
	"github.com/eventflow/api/internal/events"
	"github.com/eventflow/api/internal/k8s"
	"github.com/eventflow/api/internal/metrics"
	"github.com/eventflow/api/internal/models"
	"github.com/go-chi/chi/v5"
)

type FunctionHandler struct {
	k8sClient    *k8s.Client
	publisher    *events.Publisher
	functionRepo *database.FunctionRepository
}

func NewFunctionHandler(k8sClient *k8s.Client, publisher *events.Publisher, functionRepo *database.FunctionRepository) *FunctionHandler {
	return &FunctionHandler{
		k8sClient:    k8sClient,
		publisher:    publisher,
		functionRepo: functionRepo,
	}
}

// CreateFunction handles POST /v1/functions
func (h *FunctionHandler) CreateFunction(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Validate request
	if req.Name == "" || req.Image == "" {
		respondError(w, http.StatusBadRequest, "name and image are required", nil)
		return
	}

	if req.Replicas == 0 {
		req.Replicas = 1
	}

	// Save to database first
	function, err := h.functionRepo.Create(r.Context(), req.Name, req.Image, req.Replicas, req.Env, req.Command)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create function in database", err)
		return
	}

	// Create deployment in Kubernetes if available
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		err = h.k8sClient.CreateDeployment(
			r.Context(),
			req.Name,
			req.Image,
			req.Replicas,
			req.Env,
			req.Command,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create function in Kubernetes", err)
			return
		}

		metrics.ActiveFunctions.WithLabelValues(h.k8sClient.GetNamespace()).Inc()
	}

	// Return as FunctionStatus
	status := models.FunctionStatus{
		Name:              function.Name,
		Image:             function.Image,
		Replicas:          function.Replicas,
		AvailableReplicas: 0,
		ReadyReplicas:     0,
		UpdatedReplicas:   0,
		Status:            "Pending",
		CreatedAt:         function.CreatedAt,
	}

	respondJSON(w, http.StatusCreated, status)
}

// ListFunctions handles GET /v1/functions
func (h *FunctionHandler) ListFunctions(w http.ResponseWriter, r *http.Request) {
	// Get functions from database
	functions, err := h.functionRepo.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list functions", err)
		return
	}

	// If Kubernetes is available, enrich with deployment status
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		statusList := make([]models.FunctionStatus, 0, len(functions))

		for _, fn := range functions {
			status := models.FunctionStatus{
				Name:              fn.Name,
				Image:             fn.Image,
				Replicas:          fn.Replicas,
				AvailableReplicas: 0,
				ReadyReplicas:     0,
				UpdatedReplicas:   0,
				Status:            "Pending",
				CreatedAt:         fn.CreatedAt,
			}

			// Try to get deployment status from Kubernetes
			deployment, err := h.k8sClient.GetDeployment(r.Context(), fn.Name)
			if err == nil && deployment != nil {
				status.AvailableReplicas = deployment.Status.AvailableReplicas
				status.ReadyReplicas = deployment.Status.ReadyReplicas
				status.UpdatedReplicas = deployment.Status.UpdatedReplicas

				// Determine status
				if deployment.Status.ReadyReplicas == fn.Replicas {
					status.Status = "Running"
				} else if deployment.Status.UnavailableReplicas > 0 {
					status.Status = "Failed"
				}
			}

			statusList = append(statusList, status)
		}

		respondJSON(w, http.StatusOK, statusList)
		return
	}

	// No Kubernetes - return basic function info as status
	statusList := make([]models.FunctionStatus, 0, len(functions))
	for _, fn := range functions {
		statusList = append(statusList, models.FunctionStatus{
			Name:              fn.Name,
			Image:             fn.Image,
			Replicas:          fn.Replicas,
			AvailableReplicas: fn.Replicas, // Assume available in demo mode
			ReadyReplicas:     fn.Replicas,
			UpdatedReplicas:   fn.Replicas,
			Status:            "Running", // Demo mode - always running
			CreatedAt:         fn.CreatedAt,
		})
	}

	respondJSON(w, http.StatusOK, statusList)
}

// GetFunction handles GET /v1/functions/{name}
func (h *FunctionHandler) GetFunction(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	// Get from database
	function, err := h.functionRepo.Get(r.Context(), name)
	if err != nil {
		respondError(w, http.StatusNotFound, "function not found", err)
		return
	}

	// If Kubernetes is available, enrich with deployment status
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		status := models.FunctionStatus{
			Name:              function.Name,
			Image:             function.Image,
			Replicas:          function.Replicas,
			AvailableReplicas: 0,
			ReadyReplicas:     0,
			UpdatedReplicas:   0,
			Status:            "Pending",
			CreatedAt:         function.CreatedAt,
		}

		// Try to get deployment status from Kubernetes
		deployment, err := h.k8sClient.GetDeployment(r.Context(), name)
		if err == nil && deployment != nil {
			status.AvailableReplicas = deployment.Status.AvailableReplicas
			status.ReadyReplicas = deployment.Status.ReadyReplicas
			status.UpdatedReplicas = deployment.Status.UpdatedReplicas

			// Determine status
			if deployment.Status.ReadyReplicas == function.Replicas {
				status.Status = "Running"
			} else if deployment.Status.UnavailableReplicas > 0 {
				status.Status = "Failed"
			}
		}

		respondJSON(w, http.StatusOK, status)
		return
	}

	// No Kubernetes - return basic function info
	status := models.FunctionStatus{
		Name:              function.Name,
		Image:             function.Image,
		Replicas:          function.Replicas,
		AvailableReplicas: function.Replicas,
		ReadyReplicas:     function.Replicas,
		UpdatedReplicas:   function.Replicas,
		Status:            "Running", // Demo mode - always running
		CreatedAt:         function.CreatedAt,
	}

	respondJSON(w, http.StatusOK, status)
}

// InvokeFunction handles POST /v1/functions/{name}:invoke
func (h *FunctionHandler) InvokeFunction(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	// Parse event payload
	var req models.InvokeFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is ok
		req.Payload = make(map[string]interface{})
	}

	start := time.Now()
	eventID := fmt.Sprintf("inv-%d", time.Now().UnixNano())

	// Check if function exists in database
	function, err := h.functionRepo.Get(r.Context(), name)
	if err != nil {
		respondError(w, http.StatusNotFound, "function not found", err)
		return
	}

	// Publish event to NATS for async processing
	if h.publisher != nil {
		err := h.publisher.PublishWithMetadata("http.invoke", name, function.Image, function.Command, req.Payload)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to publish event", err)
			return
		}

		// Record invocation in database
		_ = h.functionRepo.RecordInvocation(r.Context(), name, eventID, "http.invoke", req.Payload)

		metrics.FunctionInvocations.WithLabelValues(name, h.k8sClient.GetNamespace()).Inc()

		respondJSON(w, http.StatusAccepted, map[string]interface{}{
			"message":  "function invocation queued",
			"name":     name,
			"status":   "pending",
			"event_id": eventID,
			"image":    function.Image,
			"replicas": function.Replicas,
		})
		return
	}

	// Synchronous invocation if Kubernetes is available
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		pod, response, err := h.k8sClient.InvokeFunction(r.Context(), name, "POST", "/", r.Body)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to invoke function", err)
			return
		}

		duration := time.Since(start).Seconds()

		// Record invocation in database
		_ = h.functionRepo.RecordInvocation(r.Context(), name, eventID, "http.invoke", req.Payload)

		metrics.FunctionInvocations.WithLabelValues(name, h.k8sClient.GetNamespace()).Inc()
		metrics.FunctionDuration.WithLabelValues(name, h.k8sClient.GetNamespace()).Observe(duration)

		var podName string
		if pod != nil {
			podName = pod.Name
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message":     "function invoked successfully",
			"name":        name,
			"event_id":    eventID,
			"pod":         podName,
			"response":    string(response),
			"duration_ms": duration * 1000,
		})
		return
	}

	// Demo mode - simulate invocation
	duration := time.Since(start).Seconds()

	_ = h.functionRepo.RecordInvocation(r.Context(), name, eventID, "http.invoke", req.Payload)

	metrics.FunctionInvocations.WithLabelValues(name, "default").Inc()
	metrics.FunctionDuration.WithLabelValues(name, "default").Observe(duration)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":     "function invoked (demo mode)",
		"name":        name,
		"event_id":    eventID,
		"image":       function.Image,
		"replicas":    function.Replicas,
		"duration_ms": duration * 1000,
	})
}

// DeleteFunction handles DELETE /v1/functions/{name}
func (h *FunctionHandler) DeleteFunction(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	// Delete from database
	err := h.functionRepo.Delete(r.Context(), name)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete function from database", err)
		return
	}

	// Delete from Kubernetes if available
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		err = h.k8sClient.DeleteDeployment(r.Context(), name)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to delete function from Kubernetes", err)
			return
		}

		metrics.ActiveFunctions.WithLabelValues(h.k8sClient.GetNamespace()).Dec()
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "function deleted successfully",
		"name":    name,
	})
}

// GetFunctionLogs handles GET /v1/functions/{name}/logs
func (h *FunctionHandler) GetFunctionLogs(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	follow := r.URL.Query().Get("follow") == "true"

	// Check if Kubernetes is available
	if h.k8sClient == nil || !h.k8sClient.HasKubernetes() {
		respondError(w, http.StatusNotImplemented, "logs not available in demo mode", nil)
		return
	}

	logStream, err := h.k8sClient.GetPodLogs(r.Context(), name, follow)
	if err != nil {
		respondError(w, http.StatusNotFound, "failed to get logs", err)
		return
	}
	defer logStream.Close()

	// Set headers for streaming
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	// Stream logs to client
	if flusher, ok := w.(http.Flusher); ok {
		io.Copy(w, logStream)
		flusher.Flush()
	} else {
		io.Copy(w, logStream)
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string, err error) {
	errMsg := message
	if err != nil {
		errMsg = fmt.Sprintf("%s: %v", message, err)
	}

	respondJSON(w, status, models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: errMsg,
	})
}
