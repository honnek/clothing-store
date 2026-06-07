package grpcserver

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	gs  *grpc.Server
	lis net.Listener
}

func New(lis net.Listener, opts ...grpc.ServerOption) *Server {
	return &Server{gs: grpc.NewServer(opts...), lis: lis}
}

// Grpc отдаёт нижележащий сервер, чтобы сервис зарегистрировал свои хендлеры
// до Run.
func (s *Server) Grpc() *grpc.Server { return s.gs }

// Run обслуживает запросы, пока не отменят ctx, затем мягко останавливается
// (даёт текущим RPC доработать).
func (s *Server) Run(ctx context.Context) error {
	errc := make(chan error, 1)
	go func() { errc <- s.gs.Serve(s.lis) }()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		s.gs.GracefulStop()
		return nil
	}
}
