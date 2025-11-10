package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eventflow/builder"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

type BuildJob struct {
	ID           string
	FunctionName string
	UserID       string
	Namespace    string
	Runtime      string
	SourceCode   string
	Status       string
}

type BuildEvent struct {
	BuildID      string `json:"build_id"`
	FunctionName string `json:"function_name"`
	Namespace    string `json:"namespace"`
	UserID       string `json:"user_id"`
	Runtime      string `json:"runtime"`
}

type Worker struct {
	db       *sql.DB
	nc       *nats.Conn
	sub      *nats.Subscription
	builder  *builder.Builder
	registry string
	interval time.Duration
}

func NewWorker(databaseURL, natsURL, registry string) (*Worker, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &Worker{
		db:       db,
		nc:       nc,
		builder:  builder.NewBuilder(),
		registry: registry,
		interval: 30 * time.Second, // Fallback polling - less frequent now
	}, nil
}

func (w *Worker) Start(ctx context.Context) error {
	log.Println("🚀 Builder worker started (event-driven mode)")
	log.Printf("   Registry: %s", w.registry)
	log.Printf("   Fallback poll interval: %v", w.interval)

	// Subscribe to build events
	sub, err := w.nc.Subscribe("eventflow.events", func(msg *nats.Msg) {
		w.handleBuildEvent(ctx, msg.Data)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to build events: %w", err)
	}
	w.sub = sub
	defer sub.Unsubscribe()

	log.Println("✅ Subscribed to eventflow.events")

	// Fallback ticker for missed events
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down worker...")
			return ctx.Err()

		case <-ticker.C:
			// Process any pending jobs that might have been missed
			if err := w.processPendingJobs(ctx); err != nil {
				log.Printf("Error processing pending jobs: %v", err)
			}
		}
	}
}

func (w *Worker) handleBuildEvent(ctx context.Context, data []byte) {
	var event struct {
		Type    string                 `json:"type"`
		Payload map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return
	}

	// Only process build.created events
	if event.Type != "build.created" {
		return
	}

	buildID, ok := event.Payload["build_id"].(string)
	if !ok {
		log.Printf("Invalid build_id in event payload")
		return
	}

	log.Printf("📨 Received build event for job: %s", buildID)

	// Get job details from database
	job, err := w.getJobByID(ctx, buildID)
	if err != nil {
		log.Printf("Failed to get job %s: %v", buildID, err)
		return
	}

	// Process the job
	if err := w.processJob(ctx, job); err != nil {
		log.Printf("❌ Job %s failed: %v", job.ID, err)
		w.updateJobStatus(ctx, job.ID, "failed", "", err.Error(), "")
	}
}

func (w *Worker) processPendingJobs(ctx context.Context) error {
	// Get pending jobs (fallback for missed events)
	jobs, err := w.getPendingJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	if len(jobs) > 0 {
		log.Printf("⚠️  Found %d pending job(s) in fallback poll", len(jobs))
	}

	// Process each job
	for _, job := range jobs {
		if err := w.processJob(ctx, job); err != nil {
			log.Printf("❌ Job %s failed: %v", job.ID, err)
			w.updateJobStatus(ctx, job.ID, "failed", "", err.Error(), "")
		}
	}

	return nil
}

func (w *Worker) processJobs(ctx context.Context) error {
	// Get pending jobs
	jobs, err := w.getPendingJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil
	}

	log.Printf("📦 Found %d pending build job(s)", len(jobs))

	// Process each job
	for _, job := range jobs {
		if err := w.processJob(ctx, job); err != nil {
			log.Printf("❌ Job %s failed: %v", job.ID, err)
			w.updateJobStatus(ctx, job.ID, "failed", "", err.Error(), "")
		}
	}

	return nil
}

func (w *Worker) processJob(ctx context.Context, job BuildJob) error {
	log.Printf("🔨 Building %s/%s (runtime: %s, job: %s)",
		job.Namespace, job.FunctionName, job.Runtime, job.ID)

	// Update status to building
	if err := w.updateJobStatus(ctx, job.ID, "building", "", "", "Build started...\n"); err != nil {
		return err
	}

	// Build the image
	req := builder.BuildRequest{
		FunctionName: job.FunctionName,
		Runtime:      job.Runtime,
		SourceCode:   job.SourceCode,
		UserID:       job.UserID,
		Namespace:    job.Namespace,
		Registry:     w.registry,
	}

	image, err := w.builder.Build(ctx, req)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Update status to pushing
	if err := w.updateJobStatus(ctx, job.ID, "pushing", image, "", "Pushing image...\n"); err != nil {
		return err
	}

	// Update status to success
	if err := w.updateJobStatus(ctx, job.ID, "success", image, "", "✅ Build completed successfully\n"); err != nil {
		return err
	}

	log.Printf("✅ Job %s completed: %s", job.ID, image)

	// TODO: Trigger function deployment here
	// This could call the API to create the Function CR with the built image

	return nil
}

func (w *Worker) getJobByID(ctx context.Context, id string) (BuildJob, error) {
	query := `
		SELECT id, function_name, user_id, namespace, runtime, source_code, status
		FROM build_jobs
		WHERE id = $1
	`

	var job BuildJob
	err := w.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.FunctionName, &job.UserID,
		&job.Namespace, &job.Runtime, &job.SourceCode, &job.Status,
	)
	if err != nil {
		return BuildJob{}, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

func (w *Worker) getPendingJobs(ctx context.Context) ([]BuildJob, error) {
	query := `
		SELECT id, function_name, user_id, namespace, runtime, source_code, status
		FROM build_jobs
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT 5
	`

	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []BuildJob
	for rows.Next() {
		var job BuildJob
		if err := rows.Scan(&job.ID, &job.FunctionName, &job.UserID,
			&job.Namespace, &job.Runtime, &job.SourceCode, &job.Status); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

func (w *Worker) updateJobStatus(ctx context.Context, id, status, image, errorMsg, logs string) error {
	query := `
		UPDATE build_jobs
		SET status = $1, image = $2, error = $3, 
		    logs = COALESCE(logs, '') || $4, 
		    updated_at = NOW()
	`
	args := []interface{}{status, image, errorMsg, logs}

	if status == "building" {
		query += `, started_at = NOW()`
	}

	if status == "success" || status == "failed" {
		query += `, completed_at = NOW()`
	}

	query += ` WHERE id = $5`
	args = append(args, id)

	_, err := w.db.ExecContext(ctx, query, args...)
	return err
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://eventflow:eventflow123@postgres.eventflow.svc.cluster.local:5432/eventflow?sslmode=disable"
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats.eventflow.svc.cluster.local:4222"
	}

	registry := os.Getenv("REGISTRY_URL")
	if registry == "" {
		registry = "docker-registry.eventflow.svc.cluster.local:5000"
	}

	worker, err := NewWorker(databaseURL, natsURL, registry)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.nc.Close()

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	if err := worker.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Worker error: %v", err)
	}

	log.Println("Worker stopped")
}
