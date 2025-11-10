package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eventflow/api/internal/auth"
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
	// Ensure request body is closed to prevent file descriptor leaks
	defer r.Body.Close()

	// Extract user from JWT token
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

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

	// Auto-generate namespace from user ID
	req.Namespace = claims.Namespace // tenant-{userID}

	if req.Replicas == 0 {
		req.Replicas = 1
	}

	// Ensure tenant namespace exists
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		if err := h.k8sClient.EnsureNamespace(r.Context(), req.Namespace); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create tenant namespace", err)
			return
		}
	}

	// Save to database first
	function, err := h.functionRepo.Create(r.Context(), claims.UserID, req.Name, req.Namespace, req.Image, req.Replicas, req.Env, req.Command)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create function in database", err)
		return
	}

	// Create Function CR in Kubernetes if available
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		err = h.k8sClient.CreateFunctionCR(r.Context(), req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create function in Kubernetes", err)
			return
		}
		metrics.ActiveFunctions.WithLabelValues(req.Namespace).Inc()
	}

	// Return as FunctionStatus
	status := models.FunctionStatus{
		Name:              function.Name,
		Namespace:         function.Namespace,
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
	// Extract user from JWT token
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	// Get functions from database (scoped to user)
	functions, err := h.functionRepo.List(r.Context(), claims.UserID)
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
				Namespace:         fn.Namespace,
				Image:             fn.Image,
				Replicas:          fn.Replicas,
				AvailableReplicas: 0,
				ReadyReplicas:     0,
				UpdatedReplicas:   0,
				Status:            "Pending",
				CreatedAt:         fn.CreatedAt,
			}

			// Try to get deployment status from Kubernetes
			deployment, err := h.k8sClient.GetDeployment(r.Context(), fn.Namespace, fn.Name)
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
	// Extract user from JWT token
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	name := chi.URLParam(r, "name")

	// Get from database (scoped to user and their namespace)
	function, err := h.functionRepo.Get(r.Context(), claims.UserID, name, claims.Namespace)
	if err != nil {
		respondError(w, http.StatusNotFound, "function not found", err)
		return
	}

	// If Kubernetes is available, enrich with deployment status
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		status := models.FunctionStatus{
			Name:              function.Name,
			Namespace:         function.Namespace,
			Image:             function.Image,
			Replicas:          function.Replicas,
			AvailableReplicas: 0,
			ReadyReplicas:     0,
			UpdatedReplicas:   0,
			Status:            "Pending",
			CreatedAt:         function.CreatedAt,
		}

		// Try to get deployment status from Kubernetes
		deployment, err := h.k8sClient.GetDeployment(r.Context(), function.Namespace, name)
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

// InvokeFunction handles POST /v1/functions/{name}/invoke
func (h *FunctionHandler) InvokeFunction(w http.ResponseWriter, r *http.Request) {
	// Ensure request body is closed to prevent file descriptor leaks
	defer r.Body.Close()

	// This is a placeholder for an alternative function invocation method

	// Extract user from JWT token
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}
	// Get function name from payload
	functionName := chi.URLParam(r, "name")

	// Check if function exists in database (scoped to user)
	function, err := h.functionRepo.Get(r.Context(), claims.UserID, functionName, claims.Namespace)
	if err != nil {
		respondError(w, http.StatusNotFound, "function not found", err)
		return
	}

	// Ensure tenant namespace exists
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		if err := h.k8sClient.EnsureNamespace(r.Context(), claims.Namespace); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create tenant namespace", err)
			return
		}
		var functionRequest models.CreateFunctionRequest
		functionRequest.Name = function.Name
		functionRequest.Namespace = function.Namespace
		functionRequest.Image = function.Image
		functionRequest.Replicas = function.Replicas
		functionRequest.Env = function.Env
		functionRequest.Command = function.Command
		// Create Function CR in Kubernetes if available
		err = h.k8sClient.CreateFunctionCR(r.Context(), functionRequest)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create function in database", err)
			return
		}
		metrics.ActiveFunctions.WithLabelValues(function.Namespace).Inc()
	}
}

// DeleteFunction handles DELETE /v1/functions/{name}
func (h *FunctionHandler) DeleteFunction(w http.ResponseWriter, r *http.Request) {
	// Extract user from JWT token
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	name := chi.URLParam(r, "name")

	// Delete from database (scoped to user)
	err := h.functionRepo.Delete(r.Context(), name, claims.Namespace, claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete function from database", err)
		return
	}

	// Delete from Kubernetes if available
	if h.k8sClient != nil && h.k8sClient.HasKubernetes() {
		err = h.k8sClient.DeleteFunctionCR(r.Context(), name, claims.Namespace, claims.UserID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to delete function from Kubernetes", err)
			return
		}

		metrics.ActiveFunctions.WithLabelValues(name).Dec()
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "function deleted successfully",
		"name":    name,
	})
}

// GetFunctionLogs handles GET /v1/functions/{name}/logs
func (h *FunctionHandler) GetFunctionLogs(w http.ResponseWriter, r *http.Request) {
	// Extract user from JWT token
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	name := chi.URLParam(r, "name")
	follow := r.URL.Query().Get("follow") == "true"

	// Check if Kubernetes is available
	if h.k8sClient == nil || !h.k8sClient.HasKubernetes() {
		respondError(w, http.StatusNotImplemented, "logs not available in demo mode", nil)
		return
	}

	logStream, err := h.k8sClient.GetPodLogs(r.Context(), claims.Namespace, name, follow)
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

func (h *FunctionHandler) UndeployFunction(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}
	functionName := chi.URLParam(r, "name")
	if functionName == "" {
		respondError(w, http.StatusBadRequest, "function name is required", nil)
		return
	}
	if h.k8sClient == nil || !h.k8sClient.HasKubernetes() {
		respondError(w, http.StatusNotImplemented, "undeploy not available in demo mode", nil)
		return
	}

	// Get function from database to verify ownership
	fmt.Println("Undeploying function:", functionName, claims.UserID, claims.Namespace)
	function, err := h.functionRepo.Get(r.Context(), claims.UserID, functionName, claims.Namespace)
	if err != nil {
		respondError(w, http.StatusNotFound, "function not found", err)
		return
	}
	err = h.k8sClient.DeleteFunctionCR(r.Context(), function.Name, claims.Namespace, claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to undeploy function", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "function undeployed successfully",
		"name":    functionName,
	})

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
