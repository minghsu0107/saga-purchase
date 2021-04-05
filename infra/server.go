package infra

import (
	"context"

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
func (s *Server) GracefulStop(ctx context.Context, done chan bool) {
	errs := make(chan error, 1)
	go func() {
		errs <- s.HTTPServer.GracefulStop(ctx)
	}()
	err := <-errs
	if err != nil {
		log.Error(err)
	}
	log.Info("gracefully shutdowned")
	done <- true
}
