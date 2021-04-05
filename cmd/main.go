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
		ocagent.WithServiceName("purchase"))
	if err != nil {
		log.Fatalf("failed to create ocagent-exporter: %v", err)
	}
	trace.RegisterExporter(oce)

	errs := make(chan error, 1)
	if promPort != "" {
		// Start prometheus server
		go func() {
			log.Infof("starting prom metrics on PROM_PORT=[%s]", promPort)
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
		errs <- server.Run()
	}()

	// Catch shutdown
	done := make(chan bool, 1)
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
		server.GracefulStop(ctx, done)
	}()

	err = <-errs
	if err != nil {
		log.Fatal(err)
	}

	// wait for graceful shutdown
	<-done
}
