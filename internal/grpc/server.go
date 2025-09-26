package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// Server wraps the gRPC server configuration.
type Server struct {
	server *grpc.Server
	addr   string
}

// NewServer creates a new gRPC server with unary interceptors (tracing, auth etc.).
func NewServer(addr string, opts ...grpc.ServerOption) *Server {
	return &Server{
		server: grpc.NewServer(opts...),
		addr:   addr,
	}
}

// Serve listens on the configured address and blocks until shutdown.
func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return s.server.Serve(lis)
}

// GracefulStop gracefully terminates the server.
func (s *Server) GracefulStop(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		s.server.Stop()
	case <-done:
	}
}

// Underlying exposes the raw gRPC server for service registration.
func (s *Server) Underlying() *grpc.Server {
	return s.server
}
