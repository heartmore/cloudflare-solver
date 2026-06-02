package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/flaresolverr-gateway/solver/internal/model"
	"github.com/flaresolverr-gateway/solver/internal/store"
)

type Handler struct {
	store    store.TaskStore
	enqueue  func(*model.Task) error
	queueLen func() int64
}

func NewHandler(st store.TaskStore, enq func(*model.Task) error, ql func() int64) *Handler {
	return &Handler{store: st, enqueue: enq, queueLen: ql}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		r.Use(RecoverMiddleware)
		r.Use(LoggerMiddleware)
		r.Get("/health", h.Health)
		r.Group(func(r chi.Router) {
			r.Use(APIKeyMiddleware)
			r.Post("/solve", h.Solve)
			r.Get("/result/{taskID}", h.GetResult)
			r.Get("/tasks", h.ListTasks)
			r.Get("/stats", h.Stats)
		})
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "queue": h.queueLen()})
}

func (h *Handler) Solve(w http.ResponseWriter, r *http.Request) {
	ns := NamespaceFromCtx(r)
	var req model.SolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}
	task := &model.Task{ID: uuid.NewString(), Status: "pending", Req: req, CreatedAt: time.Now()}
	if err := h.store.Put(r.Context(), ns, task); err != nil {
		writeError(w, http.StatusInternalServerError, "store error")
		return
	}
	if err := h.enqueue(task); err != nil {
		writeError(w, http.StatusInternalServerError, "enqueue error")
		return
	}
	log.Printf("[api] task %s -> %s (ns=%s)", task.ID, req.URL, ns)
	writeJSON(w, http.StatusAccepted, model.SolveResponse{TaskID: task.ID, Status: "pending"})
}

func (h *Handler) GetResult(w http.ResponseWriter, r *http.Request) {
	ns := NamespaceFromCtx(r)
	taskID := chi.URLParam(r, "taskID")
	task, err := h.store.Get(r.Context(), ns, taskID)
	if err != nil || task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	resp := model.ResultResponse{TaskID: task.ID, Status: task.Status, Error: task.Error}
	if task.Result != nil { resp.Solution = task.Result }
	if !task.EndedAt.IsZero() { resp.Duration = task.EndedAt.Sub(task.StartedAt).Milliseconds() }
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	ns := NamespaceFromCtx(r)
	keys, _ := h.store.List(r.Context(), ns, 50)
	tasks := make([]*model.ResultResponse, 0, len(keys))
	for _, key := range keys {
		taskID := key
		if strings.HasPrefix(key, "solver:task:") { taskID = key[len("solver:task:"):] }
		task, err := h.store.Get(r.Context(), ns, taskID)
		if err != nil || task == nil { continue }
		resp := &model.ResultResponse{TaskID: task.ID, Status: task.Status, Error: task.Error}
		if task.Result != nil { resp.Solution = task.Result }
		if !task.EndedAt.IsZero() { resp.Duration = task.EndedAt.Sub(task.StartedAt).Milliseconds() }
		tasks = append(tasks, resp)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks, "total": len(tasks)})
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	ns := NamespaceFromCtx(r)
	writeJSON(w, http.StatusOK, h.store.Stats(r.Context(), ns))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
