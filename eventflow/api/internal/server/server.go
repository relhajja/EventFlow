package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/eventflow/api/internal/auth"
	"github.com/eventflow/api/internal/config"
	"github.com/eventflow/api/internal/database"
	"github.com/eventflow/api/internal/events"
	"github.com/eventflow/api/internal/handlers"
	"github.com/eventflow/api/internal/k8s"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	config    *config.Config
	k8sClient *k8s.Client
	db        *database.DB
	auth      *auth.Authenticator
	router    *chi.Mux
	publisher *events.Publisher
}

func New(cfg *config.Config, k8sClient *k8s.Client, db *database.DB) *Server {
	// Initialize NATS publisher (optional)
	var publisher *events.Publisher
	natsURL := os.Getenv("NATS_URL")
	if natsURL != "" {
		pub, err := events.NewPublisher(natsURL)
		if err == nil {
			publisher = pub
		}
	}

	s := &Server{
		config:    cfg,
		k8sClient: k8sClient,
		db:        db,
		auth:      auth.NewAuthenticator(cfg.JWTSecret),
		router:    chi.NewRouter(),
		publisher: publisher,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// CORS
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes() {
	// Initialize function repository
	functionRepo := database.NewFunctionRepository(s.db)
	functionHandler := handlers.NewFunctionHandler(s.k8sClient, s.publisher, functionRepo)

	// Public routes
	s.router.Get("/healthz", s.healthHandler)
	s.router.Get("/readyz", s.readyHandler)
	s.router.Get("/metrics", promhttp.Handler().ServeHTTP)

	// Dev auth endpoint (generate tokens for testing)
	s.router.Post("/auth/token", s.generateTokenHandler)

	// Protected API routes
	s.router.Route("/v1", func(r chi.Router) {
		r.Use(s.auth.Middleware)

		r.Route("/functions", func(r chi.Router) {
			r.Get("/", functionHandler.ListFunctions)
			r.Post("/", functionHandler.CreateFunction)
			r.Get("/{name}", functionHandler.GetFunction)
			r.Delete("/{name}", functionHandler.DeleteFunction)
			r.Post("/{name}:invoke", functionHandler.InvokeFunction)
			r.Post("/{name}/undeploy", functionHandler.UndeployFunction)
			r.Get("/{name}/logs", functionHandler.GetFunctionLogs)
		})
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	// Could add checks for Kubernetes connectivity here
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

func (s *Server) generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	// For development only - generate a test token
	// In production, this would authenticate against an identity provider

	var req struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	// Parse request body for user selection
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to demo user if no body provided
		req.UserID = "demo-user"
		req.Username = "Demo User"
		req.Email = "demo@eventflow.io"
	}

	// Validate required fields
	if req.UserID == "" {
		req.UserID = "demo-user"
	}
	if req.Username == "" {
		req.Username = req.UserID
	}

	token, err := s.auth.GenerateToken(req.UserID, req.Username, req.Email)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":     token,
		"user_id":   req.UserID,
		"username":  req.Username,
		"email":     req.Email,
		"namespace": fmt.Sprintf("tenant-%s", req.UserID),
	})
}

func (s *Server) Router() http.Handler {
	return s.router
}
