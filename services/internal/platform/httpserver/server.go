package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Server struct {
	srv *http.Server
}

func New(addr string, h http.Handler) *Server {
	return &Server{srv: &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}}
}

// Run обслуживает запросы, пока не отменят ctx, затем доживает текущие в пределах
// timeout. Ошибка прослушивания (например, занятый порт) возвращается сразу.
func (s *Server) Run(ctx context.Context, timeout time.Duration) error {
	errc := make(chan error, 1)
	go func() {
		err := s.srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errc <- err
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return s.srv.Shutdown(shutCtx)
	}
}
