package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type BuildJob struct {
	ID           string     `json:"id"`
	FunctionName string     `json:"function_name"`
	UserID       string     `json:"user_id"`
	Namespace    string     `json:"namespace"`
	Runtime      string     `json:"runtime"`
	SourceCode   string     `json:"source_code"`
	Status       string     `json:"status"` // pending, queued, building, pushing, success, failed
	Image        string     `json:"image,omitempty"`
	Error        string     `json:"error,omitempty"`
	Logs         string     `json:"logs,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type BuildJobRepository struct {
	db        *DB
	publisher Publisher
}

// Publisher interface for publishing events
type Publisher interface {
	Publish(eventType, function string, payload map[string]interface{}) error
}

func NewBuildJobRepository(db *DB, publisher Publisher) *BuildJobRepository {
	return &BuildJobRepository{
		db:        db,
		publisher: publisher,
	}
}

// Create creates a new build job
func (r *BuildJobRepository) Create(ctx context.Context, functionName, userID, namespace, runtime, sourceCode string) (*BuildJob, error) {
	job := &BuildJob{
		ID:           uuid.New().String(),
		FunctionName: functionName,
		UserID:       userID,
		Namespace:    namespace,
		Runtime:      runtime,
		SourceCode:   sourceCode,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO build_jobs (id, function_name, user_id, namespace, runtime, source_code, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool().Exec(ctx, query,
		job.ID, job.FunctionName, job.UserID, job.Namespace,
		job.Runtime, job.SourceCode, job.Status,
		job.CreatedAt, job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create build job: %w", err)
	}

	// Publish build event
	if r.publisher != nil {
		err = r.publisher.Publish("build.created", job.FunctionName, map[string]interface{}{
			"build_id":      job.ID,
			"function_name": job.FunctionName,
			"namespace":     job.Namespace,
			"user_id":       job.UserID,
			"runtime":       job.Runtime,
		})
		if err != nil {
			// Log error but don't fail the request - worker will pick it up via fallback
			fmt.Printf("Warning: failed to publish build event: %v\n", err)
		}
	}

	return job, nil
}

// Get retrieves a build job by ID
func (r *BuildJobRepository) Get(ctx context.Context, id string) (*BuildJob, error) {
	job := &BuildJob{}

	query := `
		SELECT id, function_name, user_id, namespace, runtime, source_code,
		       status, image, error, logs, started_at, completed_at, created_at, updated_at
		FROM build_jobs
		WHERE id = $1
	`

	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&job.ID, &job.FunctionName, &job.UserID, &job.Namespace,
		&job.Runtime, &job.SourceCode, &job.Status, &job.Image,
		&job.Error, &job.Logs, &job.StartedAt, &job.CompletedAt,
		&job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get build job: %w", err)
	}

	return job, nil
}

// GetByFunction retrieves build jobs for a function
func (r *BuildJobRepository) GetByFunction(ctx context.Context, functionName, namespace string) ([]*BuildJob, error) {
	query := `
		SELECT id, function_name, user_id, namespace, runtime, source_code,
		       status, image, error, logs, started_at, completed_at, created_at, updated_at
		FROM build_jobs
		WHERE function_name = $1 AND namespace = $2
		ORDER BY created_at DESC
		LIMIT 10
	`

	rows, err := r.db.Pool().Query(ctx, query, functionName, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to query build jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*BuildJob
	for rows.Next() {
		job := &BuildJob{}
		err := rows.Scan(
			&job.ID, &job.FunctionName, &job.UserID, &job.Namespace,
			&job.Runtime, &job.SourceCode, &job.Status, &job.Image,
			&job.Error, &job.Logs, &job.StartedAt, &job.CompletedAt,
			&job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan build job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// ListPending retrieves all pending build jobs
func (r *BuildJobRepository) ListPending(ctx context.Context) ([]*BuildJob, error) {
	query := `
		SELECT id, function_name, user_id, namespace, runtime, source_code,
		       status, image, error, logs, started_at, completed_at, created_at, updated_at
		FROM build_jobs
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT 10
	`

	rows, err := r.db.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*BuildJob
	for rows.Next() {
		job := &BuildJob{}
		err := rows.Scan(
			&job.ID, &job.FunctionName, &job.UserID, &job.Namespace,
			&job.Runtime, &job.SourceCode, &job.Status, &job.Image,
			&job.Error, &job.Logs, &job.StartedAt, &job.CompletedAt,
			&job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan build job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// UpdateStatus updates the status of a build job
func (r *BuildJobRepository) UpdateStatus(ctx context.Context, id, status string, image, errorMsg, logs string) error {
	now := time.Now()

	query := `
		UPDATE build_jobs
		SET status = $1, image = $2, error = $3, logs = $4, updated_at = $5
	`
	args := []interface{}{status, image, errorMsg, logs, now}

	// Set started_at for building status
	if status == "building" {
		query += `, started_at = $6`
		args = append(args, now)
	}

	// Set completed_at for terminal statuses
	if status == "success" || status == "failed" {
		if status == "building" {
			query += `, completed_at = $7 WHERE id = $8`
			args = append(args, now, id)
		} else {
			query += `, completed_at = $6 WHERE id = $7`
			args = append(args, now, id)
		}
	} else {
		query += ` WHERE id = $6`
		args = append(args, id)
	}

	_, err := r.db.Pool().Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update build job status: %w", err)
	}

	return nil
}

// AppendLogs appends logs to a build job
func (r *BuildJobRepository) AppendLogs(ctx context.Context, id, newLogs string) error {
	query := `
		UPDATE build_jobs
		SET logs = COALESCE(logs, '') || $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Pool().Exec(ctx, query, newLogs, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to append logs: %w", err)
	}

	return nil
}

// Delete deletes a build job
func (r *BuildJobRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM build_jobs WHERE id = $1`

	_, err := r.db.Pool().Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete build job: %w", err)
	}

	return nil
}
