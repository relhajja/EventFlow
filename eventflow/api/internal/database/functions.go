package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eventflow/api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type FunctionRepository struct {
	db *DB
}

func NewFunctionRepository(db *DB) *FunctionRepository {
	return &FunctionRepository{db: db}
}

// Create inserts a new function
func (r *FunctionRepository) Create(ctx context.Context, name, image string, replicas int32, env map[string]string, command []string) (*models.Function, error) {
	var envJSON []byte
	var err error
	if len(env) > 0 {
		envJSON, err = json.Marshal(env)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal env: %w", err)
		}
	}

	// PostgreSQL stores `command` as text[]; pass Go []string directly so pgx
	// encodes it as a PostgreSQL array literal. Do not JSON-marshal the slice
	// (that produces a JSON string which is invalid for text[] columns).
	var commandParam interface{}
	if len(command) > 0 {
		commandParam = command
	} else {
		commandParam = nil
	}

	query := `
		INSERT INTO functions (name, image, replicas, env, command, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, name, image, replicas, created_at, updated_at
	`

	var fn models.Function
	var id uuid.UUID

	err = r.db.pool.QueryRow(ctx, query, name, image, replicas, envJSON, commandParam).
		Scan(&id, &fn.Name, &fn.Image, &fn.Replicas, &fn.CreatedAt, &fn.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create function: %w", err)
	}

	fn.Env = env
	fn.Command = command

	return &fn, nil
}

// Get retrieves a function by name
func (r *FunctionRepository) Get(ctx context.Context, name string) (*models.Function, error) {
	query := `
		SELECT name, image, replicas, env, command, created_at, updated_at
		FROM functions
		WHERE name = $1 AND deleted_at IS NULL
	`

	var fn models.Function
	var envJSON []byte
	var commandArray []string
	
	err := r.db.pool.QueryRow(ctx, query, name).
		Scan(&fn.Name, &fn.Image, &fn.Replicas, &envJSON, &commandArray, &fn.CreatedAt, &fn.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("function not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}

	// Unmarshal env JSON
	if len(envJSON) > 0 {
		if err := json.Unmarshal(envJSON, &fn.Env); err != nil {
			return nil, fmt.Errorf("failed to unmarshal env: %w", err)
		}
	}
	
	// Assign command array
	fn.Command = commandArray

	return &fn, nil
}

// List retrieves all functions
func (r *FunctionRepository) List(ctx context.Context) ([]*models.Function, error) {
	query := `
		SELECT name, image, replicas, env, command, created_at, updated_at
		FROM functions
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}
	defer rows.Close()

	var functions []*models.Function
	for rows.Next() {
		var fn models.Function
		var envJSON []byte
		var commandArray []string
		
		err := rows.Scan(&fn.Name, &fn.Image, &fn.Replicas, &envJSON, &commandArray, &fn.CreatedAt, &fn.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan function: %w", err)
		}

		// Unmarshal env JSON
		if len(envJSON) > 0 {
			if err := json.Unmarshal(envJSON, &fn.Env); err != nil {
				return nil, fmt.Errorf("failed to unmarshal env: %w", err)
			}
		}
		
		// Assign command array
		fn.Command = commandArray

		functions = append(functions, &fn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating functions: %w", err)
	}

	return functions, nil
}

// Delete soft-deletes a function
func (r *FunctionRepository) Delete(ctx context.Context, name string) error {
	query := `
		UPDATE functions
		SET deleted_at = $1
		WHERE name = $2 AND deleted_at IS NULL
	`

	result, err := r.db.pool.Exec(ctx, query, time.Now(), name)
	if err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("function not found: %s", name)
	}

	return nil
}

// Update modifies function configuration
func (r *FunctionRepository) Update(ctx context.Context, name string, replicas int32) error {
	query := `
		UPDATE functions
		SET replicas = $1, updated_at = NOW()
		WHERE name = $2 AND deleted_at IS NULL
	`

	result, err := r.db.pool.Exec(ctx, query, replicas, name)
	if err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("function not found: %s", name)
	}

	return nil
}

// RecordInvocation logs a function invocation
func (r *FunctionRepository) RecordInvocation(ctx context.Context, functionName, eventID, eventType string, payload map[string]interface{}) error {
	// Get function ID
	var functionID uuid.UUID
	err := r.db.pool.QueryRow(ctx, "SELECT id FROM functions WHERE name = $1 AND deleted_at IS NULL", functionName).Scan(&functionID)
	if err != nil {
		return fmt.Errorf("function not found: %w", err)
	}

	payloadJSON, _ := json.Marshal(payload)

	query := `
		INSERT INTO invocations (function_id, event_id, event_type, payload, status)
		VALUES ($1, $2, $3, $4, 'pending')
	`

	_, err = r.db.pool.Exec(ctx, query, functionID, eventID, eventType, payloadJSON)
	if err != nil {
		return fmt.Errorf("failed to record invocation: %w", err)
	}

	return nil
}
