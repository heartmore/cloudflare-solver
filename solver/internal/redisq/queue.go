package redisq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/flaresolverr-gateway/solver/internal/model"
)

const (queueKey = "solver:queue"; processingKey = "solver:processing")

type Queue struct{ rdb *redis.Client }

func New(rdb *redis.Client) *Queue { return &Queue{rdb: rdb} }

func (q *Queue) Enqueue(ctx context.Context, task *model.Task) error {
	data, _ := json.Marshal(task)
	return q.rdb.LPush(ctx, queueKey, data).Err()
}

func (q *Queue) Dequeue(ctx context.Context) (*model.Task, error) {
	result, err := q.rdb.BRPop(ctx, 0, queueKey).Result()
	if err != nil { return nil, err }
	var task model.Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil { return nil, err }
	q.rdb.HSet(ctx, processingKey, task.ID, time.Now().Unix())
	return &task, nil
}

func (q *Queue) MarkDone(ctx context.Context, taskID string) { q.rdb.HDel(ctx, processingKey, taskID) }
func (q *Queue) Len(ctx context.Context) int64 { return q.rdb.LLen(ctx, queueKey).Val() }

type Worker struct {
	queue   *Queue
	process func(*model.Task)
	id      int
	stopCh  chan struct{}
}

func NewWorker(id int, q *Queue, processFn func(*model.Task)) *Worker {
	return &Worker{queue: q, process: processFn, id: id, stopCh: make(chan struct{})}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-w.stopCh: return
			default:
				task, err := w.queue.Dequeue(ctx)
				if err != nil { time.Sleep(time.Second); continue }
				w.process(task)
				w.queue.MarkDone(ctx, task.ID)
			}
		}
	}()
}

func (w *Worker) Stop() { close(w.stopCh) }
