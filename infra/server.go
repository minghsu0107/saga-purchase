package infra

import (
	"context"

	infra_http "github.com/minghsu0107/saga-purchase/infra/http"
	infra_observe "github.com/minghsu0107/saga-purchase/infra/observe"
	log "github.com/sirupsen/logrus"
)

// Server wraps http and grpc server
type Server struct {
	HTTPServer  *infra_http.Server
	ObsInjector *infra_observe.ObservibilityInjector
}

func NewServer(httpServer *infra_http.Server, obsInjector *infra_observe.ObservibilityInjector) *Server {
	return &Server{
		HTTPServer:  httpServer,
		ObsInjector: obsInjector,
	}
}

// Run server
func (s *Server) Run() error {
	errs := make(chan error, 1)
	s.ObsInjector.Register(errs)
	go func() {
		errs <- s.HTTPServer.Run()
	}()
	err := <-errs
	if err != nil {
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
