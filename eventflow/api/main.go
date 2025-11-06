package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eventflow/api/internal/config"
	"github.com/eventflow/api/internal/database"
	"github.com/eventflow/api/internal/k8s"
	"github.com/eventflow/api/internal/metrics"
	"github.com/eventflow/api/internal/server"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg := config.Load()

	// Initialize database if URL provided
	var db *database.DB
	if cfg.DatabaseURL != "" {
		log.Println("Connecting to database...")
		var err error
		db, err = database.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()
		log.Println("✅ Database connected")
	} else {
		log.Println("⚠️  No DATABASE_URL provided, running without persistence")
	}

	// Initialize Kubernetes client
	k8sClient, err := k8s.NewClient(cfg.Namespace)
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	// Initialize metrics
	metrics.Init()

	// Create HTTP server
	srv := server.New(cfg, k8sClient, db)

	// Start server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      srv.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting EventFlow API server on port %d", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
