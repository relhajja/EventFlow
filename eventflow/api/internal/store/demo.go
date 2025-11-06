package store

import (
	"sync"
	"time"

	"github.com/eventflow/api/internal/models"
)

// DemoStore is an in-memory store for demo mode
type DemoStore struct {
	mu        sync.RWMutex
	functions map[string]*models.FunctionStatus
}

var instance *DemoStore
var once sync.Once

// GetDemoStore returns the singleton demo store
func GetDemoStore() *DemoStore {
	once.Do(func() {
		instance = &DemoStore{
			functions: make(map[string]*models.FunctionStatus),
		}
	})
	return instance
}

// CreateFunction adds a function to the store
func (s *DemoStore) CreateFunction(name, image string, replicas int32, env map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.functions[name] = &models.FunctionStatus{
		Name:              name,
		Image:             image,
		Replicas:          replicas,
		AvailableReplicas: replicas,
		ReadyReplicas:     replicas,
		UpdatedReplicas:   replicas,
		Status:            "Running",
		CreatedAt:         time.Now(),
	}
}

// GetFunction retrieves a function by name
func (s *DemoStore) GetFunction(name string) (*models.FunctionStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fn, exists := s.functions[name]
	return fn, exists
}

// ListFunctions returns all functions
func (s *DemoStore) ListFunctions() []*models.FunctionStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.FunctionStatus, 0, len(s.functions))
	for _, fn := range s.functions {
		result = append(result, fn)
	}
	return result
}

// DeleteFunction removes a function from the store
func (s *DemoStore) DeleteFunction(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.functions[name]; exists {
		delete(s.functions, name)
		return true
	}
	return false
}
