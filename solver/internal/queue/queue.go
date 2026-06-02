package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/flaresolverr-gateway/solver/internal/flaresolverr"
	"github.com/flaresolverr-gateway/solver/internal/model"
	"github.com/flaresolverr-gateway/solver/internal/store"
)

type Dispatcher struct {
	fsClient *flaresolverr.Client
	store    store.TaskStore
	jobCh    chan *model.Task
	workers  int
	wg       sync.WaitGroup
}

func New(fsClient *flaresolverr.Client, st store.TaskStore, workers int) *Dispatcher {
	if workers <= 0 { workers = 2 }
	return &Dispatcher{fsClient: fsClient, store: st, jobCh: make(chan *model.Task, 100), workers: workers}
}

func (d *Dispatcher) Start() {
	for i := 0; i < d.workers; i++ { d.wg.Add(1); go d.worker(i) }
	log.Printf("[queue] %d workers", d.workers)
}

func (d *Dispatcher) Submit(task *model.Task) {
	task.Status = "pending"
	d.store.Put(context.Background(), "public", task)
	d.jobCh <- task
}

func (d *Dispatcher) Stop() { close(d.jobCh); d.wg.Wait() }

func (d *Dispatcher) worker(id int) {
	defer d.wg.Done()
	for task := range d.jobCh {
		task.Status = "running"
		task.StartedAt = time.Now()
		result, err := d.fsClient.Solve(task.Req)
		task.EndedAt = time.Now()
		d.store.UpdateResult(context.Background(), "public", task.ID, result, err)
		if err != nil { log.Printf("[w%d] FAIL %s", id, task.Req.URL) }
	}
}
