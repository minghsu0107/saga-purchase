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
	infra_broker.Publisher.Close()
	infra_broker.Subscriber.Close()
	infra_grpc.AuthClientConn.Conn.Close()
	infra_grpc.ProductClientConn.Conn.Close()

	log.Info("gracefully shutdowned")
	done <- true
}
