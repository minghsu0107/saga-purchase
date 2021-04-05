package infra

import (
	"context"
	"fmt"

	infra_http "github.com/minghsu0107/saga-purchase/infra/http"
	log "github.com/sirupsen/logrus"
)

// Server wraps http and grpc server
type Server struct {
	HTTPServer *infra_http.Server
}

func NewServer(httpServer *infra_http.Server) *Server {
	return &Server{
		HTTPServer: httpServer,
	}
}

// Run server
func (s *Server) Run() error {
	if err := s.HTTPServer.Run(); err != nil {
		return err
	}
	return nil
}

// GracefulStop server
func (s *Server) GracefulStop(ctx context.Context) error {
	if err := s.HTTPServer.GracefulStop(ctx); err != nil {
		return fmt.Errorf("error server shutdown: %v", err)
	}
	log.Info("gracefully shutdowned")
	return nil
}
