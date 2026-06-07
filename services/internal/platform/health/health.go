package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Check сообщает, доступна ли зависимость. Обязана уважать ctx для таймаутов.
type Check func(context.Context) error

// Handler отдаёт liveness- и readiness-пробы. Liveness лишь говорит, что процесс
// жив; readiness прогоняет зарегистрированные проверки зависимостей.
type Handler struct {
	mu     sync.RWMutex
	checks map[string]Check
}

func New() *Handler {
	return &Handler{checks: make(map[string]Check)}
}

func (h *Handler) Register(name string, c Check) {
	h.mu.Lock()
	h.checks[name] = c
	h.mu.Unlock()
}

func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	h.mu.RLock()
	checks := make(map[string]Check, len(h.checks))
	for name, c := range h.checks {
		checks[name] = c
	}
	h.mu.RUnlock()

	results := make(map[string]string, len(checks))
	code := http.StatusOK
	for name, check := range checks {
		if err := check(ctx); err != nil {
			results[name] = err.Error()
			code = http.StatusServiceUnavailable
			continue
		}
		results[name] = "ok"
	}
	writeJSON(w, code, results)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
