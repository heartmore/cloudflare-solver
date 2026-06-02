package store

import (
	"context"
	"strings"
	"sync"

	"github.com/flaresolverr-gateway/solver/internal/model"
)

type MemoryStore struct {
	mu    sync.RWMutex
	tasks map[string]*model.Task
}

func New() *MemoryStore { return &MemoryStore{tasks: make(map[string]*model.Task)} }

func (s *MemoryStore) Put(_ context.Context, namespace string, task *model.Task) error {
	s.mu.Lock(); defer s.mu.Unlock()
	s.tasks[namespace+":"+task.ID] = task
	return nil
}

func (s *MemoryStore) Get(_ context.Context, namespace, taskID string) (*model.Task, error) {
	s.mu.RLock(); defer s.mu.RUnlock()
	t, ok := s.tasks[namespace+":"+taskID]
	if !ok { return nil, nil }
	return t, nil
}

func (s *MemoryStore) UpdateResult(_ context.Context, namespace, taskID string, result *model.FlareSolverrResult, err error) error {
	s.mu.Lock(); defer s.mu.Unlock()
	t, ok := s.tasks[namespace+":"+taskID]
	if !ok { return nil }
	if err != nil { t.Status = "failed"; t.Error = err.Error() } else { t.Status = "completed"; t.Result = result }
	return nil
}

func (s *MemoryStore) List(_ context.Context, namespace string, limit int) ([]string, error) {
	s.mu.RLock(); defer s.mu.RUnlock()
	prefix := namespace + ":"
	keys := make([]string, 0)
	for k := range s.tasks {
		if strings.HasPrefix(k, prefix) { keys = append(keys, strings.TrimPrefix(k, prefix)) }
	}
	if limit > 0 && len(keys) > limit { keys = keys[:limit] }
	return keys, nil
}

func (s *MemoryStore) Stats(_ context.Context, namespace string) map[string]int64 {
	s.mu.RLock(); defer s.mu.RUnlock()
	stats := map[string]int64{}
	prefix := namespace + ":"
	for k, t := range s.tasks {
		if strings.HasPrefix(k, prefix) {
			stats["total"]++
			switch t.Status {
			case "completed": stats["completed"]++
			case "failed": stats["failed"]++
			default: stats["pending"]++
			}
		}
	}
	return stats
}
