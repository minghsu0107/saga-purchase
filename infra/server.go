package infra

import (
	"context"

	infra_broker "github.com/minghsu0107/saga-purchase/infra/broker"
	infra_grpc "github.com/minghsu0107/saga-purchase/infra/grpc"
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
	errs := make(chan error, 2)
	s.ObsInjector.Register(errs)
	err := <-errs
	if err != nil {
		return err
	}
	go func() {
		errs <- s.HTTPServer.Run()
	}()
	err = <-errs
	if err != nil {
		return err
	}
	return nil
}

// GracefulStop server
func (s *Server) GracefulStop(ctx context.Context, done chan bool) {
	err := s.HTTPServer.GracefulStop(ctx)
	if err != nil {
		log.Error(err)
	}

	if infra_observe.TracerProvider != nil {
		err = infra_observe.TracerProvider.Shutdown(ctx)
		if err != nil {
			log.Error(err)
		}
	}
	if err = infra_broker.Publisher.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_broker.Subscriber.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_grpc.AuthClientConn.Conn.Close(); err != nil {
		log.Error(err)
	}
	if err = infra_grpc.ProductClientConn.Conn.Close(); err != nil {
		log.Error(err)
	}

	log.Info("gracefully shutdowned")
	done <- true
}
