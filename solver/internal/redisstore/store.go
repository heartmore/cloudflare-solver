package redisstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/flaresolverr-gateway/solver/internal/model"
)

const (keyPrefix = "solver:task:"; defaultTTL = 24 * time.Hour)

type Store struct{ rdb *redis.Client }

func New(rdb *redis.Client) *Store { return &Store{rdb: rdb} }

func (s *Store) Put(ctx context.Context, namespace string, task *model.Task) error {
	data, _ := json.Marshal(task)
	return s.rdb.Set(ctx, keyPrefix+namespace+":"+task.ID, data, defaultTTL).Err()
}

func (s *Store) Get(ctx context.Context, namespace, taskID string) (*model.Task, error) {
	data, err := s.rdb.Get(ctx, keyPrefix+namespace+":"+taskID).Bytes()
	if err == redis.Nil { return nil, nil }
	if err != nil { return nil, err }
	var task model.Task
	json.Unmarshal(data, &task)
	return &task, nil
}

func (s *Store) UpdateResult(ctx context.Context, namespace, taskID string, result *model.FlareSolverrResult, err error) error {
	task, _ := s.Get(ctx, namespace, taskID)
	if task == nil { return nil }
	if err != nil { task.Status = "failed"; task.Error = err.Error() } else { task.Status = "completed"; task.Result = result }
	task.EndedAt = time.Now()
	return s.Put(ctx, namespace, task)
}

func (s *Store) List(ctx context.Context, namespace string, limit int) ([]string, error) {
	keys, _ := s.rdb.Keys(ctx, keyPrefix+namespace+":*").Result()
	if limit > 0 && len(keys) > limit { keys = keys[:limit] }
	return keys, nil
}

func (s *Store) Stats(ctx context.Context, namespace string) map[string]int64 {
	stats := map[string]int64{}
	keys, _ := s.List(ctx, namespace, 1000)
	stats["total"] = int64(len(keys))
	for _, k := range keys {
		data, _ := s.rdb.Get(ctx, k).Bytes()
		var task model.Task
		if json.Unmarshal(data, &task) != nil { continue }
		switch task.Status {
		case "completed": stats["completed"]++
		case "failed": stats["failed"]++
		default: stats["pending"]++
		}
	}
	return stats
}
