package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flaresolverr-gateway/solver/internal/ratelimit"
)

type contextKey string
const ctxNamespace contextKey = "namespace"

var globalLimiter = ratelimit.New(10, 20)

func APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") { key = strings.TrimPrefix(auth, "Bearer ") }
		}
		expectedKey := os.Getenv("SOLVER_API_KEY")
		if expectedKey != "" {
			if key == "" { writeError(w, http.StatusUnauthorized, "missing API key (X-API-Key or Bearer)"); return }
			valid := false
			for _, ek := range strings.Split(expectedKey, ",") {
				if strings.TrimSpace(ek) == key { valid = true; break }
			}
			if !valid { writeError(w, http.StatusUnauthorized, "invalid API key"); return }
		}
		namespace := "public"
		if key != "" {
			h := sha256.Sum256([]byte(key))
			namespace = hex.EncodeToString(h[:8])
		}
		if !globalLimiter.Allow(namespace) {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded"); return
		}
		ctx := context.WithValue(r.Context(), ctxNamespace, namespace)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NamespaceFromCtx(r *http.Request) string {
	ns, _ := r.Context().Value(ctxNamespace).(string)
	if ns == "" { return "public" }
	return ns
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[http] %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[http] PANIC: %v", rec)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
