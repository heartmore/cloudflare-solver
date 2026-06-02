package store

import (
	"context"
	"github.com/flaresolverr-gateway/solver/internal/model"
)

type TaskStore interface {
	Put(ctx context.Context, namespace string, task *model.Task) error
	Get(ctx context.Context, namespace, taskID string) (*model.Task, error)
	UpdateResult(ctx context.Context, namespace, taskID string, result *model.FlareSolverrResult, err error) error
	List(ctx context.Context, namespace string, limit int) ([]string, error)
	Stats(ctx context.Context, namespace string) map[string]int64
}
