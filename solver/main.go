package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"github.com/flaresolverr-gateway/solver/internal/api"
	"github.com/flaresolverr-gateway/solver/internal/flaresolverr"
	"github.com/flaresolverr-gateway/solver/internal/model"
	"github.com/flaresolverr-gateway/solver/internal/redisq"
	"github.com/flaresolverr-gateway/solver/internal/redisstore"
	"github.com/flaresolverr-gateway/solver/internal/store"
	"github.com/flaresolverr-gateway/solver/internal/web"
)

func main() {
	fsURL := getEnv("FLARESOLVERR_URL", "http://localhost:8191")
	port := getEnv("PORT", "8080")
	workers, _ := strconv.Atoi(getEnv("WORKERS", "2"))
	redisURL := os.Getenv("REDIS_URL")

	var rdb *redis.Client
	if redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("[main] invalid REDIS_URL: %v", err)
		}
		rdb = redis.NewClient(opts)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Printf("[main] WARNING: Redis unreachable (%v)", err)
			rdb = nil
		}
	}

	fsClient := flaresolverr.New(fsURL)

	var taskStore store.TaskStore
	if rdb != nil {
		taskStore = redisstore.New(rdb)
	} else {
		log.Println("[main] No Redis — in-memory store")
		taskStore = store.New()
	}

	var enqueue func(*model.Task) error
	var queueLen func() int64

	if rdb != nil {
		q := redisq.New(rdb)
		for i := 0; i < workers; i++ {
			w := redisq.NewWorker(i, q, func(task *model.Task) {
				processTask(fsClient, taskStore, task)
			})
			w.Start(context.Background())
			defer w.Stop()
		}
		enqueue = func(t *model.Task) error { return q.Enqueue(context.Background(), t) }
		queueLen = func() int64 { return q.Len(context.Background()) }
	} else {
		jobCh := make(chan *model.Task, 100)
		for i := 0; i < workers; i++ {
			go func(id int) {
				for task := range jobCh {
					processTask(fsClient, taskStore, task)
				}
			}(i)
		}
		enqueue = func(t *model.Task) error { jobCh <- t; return nil }
		queueLen = func() int64 { return int64(len(jobCh)) }
	}

	apiHandler := api.NewHandler(taskStore, enqueue, queueLen)

	r := chi.NewRouter()
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("pong")) })
	apiHandler.RegisterRoutes(r)
	r.Get("/v1/openapi.yaml", web.OpenAPIHandler)
	r.Get("/docs", web.DocsHandler)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})
	r.Handle("/static/*", http.StripPrefix("/static/", web.Handler()))

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		srv.Close()
	}()

	log.Printf("[main] listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[main] %v", err)
	}
}

func processTask(fs *flaresolverr.Client, st store.TaskStore, task *model.Task) {
	task.Status = "running"
	task.StartedAt = time.Now()
	st.Put(context.Background(), "public", task)
	result, err := fs.Solve(task.Req)
	task.EndedAt = time.Now()
	st.UpdateResult(context.Background(), "public", task.ID, result, err)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" { return v }
	return fallback
}
