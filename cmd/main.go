package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/minghsu0107/saga-purchase/dep"
	"github.com/minghsu0107/saga-purchase/infra/broker"
	infrahttp "github.com/minghsu0107/saga-purchase/infra/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opencensus.io/trace"
)

var (
	promPort    = os.Getenv("PROM_PORT")
	ocagentHost = os.Getenv("OC_AGENT_HOST")
)

func main() {
	oce, err := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithReconnectionPeriod(5*time.Second),
		ocagent.WithAddress(ocagentHost),
		ocagent.WithServiceName("voting"))
	if err != nil {
		log.Fatalf("Failed to create ocagent-exporter: %v", err)
	}
	trace.RegisterExporter(oce)

	errs := make(chan error, 1)
	if promPort != "" {
		// Start prometheus server
		go func() {
			log.Printf("Starting prom metrics on PROM_PORT=[%s]", promPort)
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(fmt.Sprintf(":%s", promPort), nil)
			errs <- err
		}()
	}

	server, err := dep.InitializeServer()
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Publisher.Close()
	defer broker.Subscriber.Close()
	go func() {
		err := server.Run()
		if err != nil {
			errs <- err
		}
	}()

	// Catch shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		// kill (no param) default send syscall.SIGTERM
		// kill -2 is syscall.SIGINT
		// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		// graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		errs <- gracefulShutdown(ctx, server)
	}()

	log.Fatal(<-errs)
}

func gracefulShutdown(ctx context.Context, server *infrahttp.Server) error {
	if err := server.Svr.Shutdown(ctx); err != nil {
		return fmt.Errorf("error server shutdown: %v", err)
	}
	return fmt.Errorf("gracefully shutdowned")
}
