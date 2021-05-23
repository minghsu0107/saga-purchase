package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/minghsu0107/saga-purchase/dep"
	"github.com/minghsu0107/saga-purchase/infra/broker"
	"github.com/minghsu0107/saga-purchase/infra/grpc"
)

func main() {
	errs := make(chan error, 1)

	server, err := dep.InitializeServer()
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Publisher.Close()
	defer broker.Subscriber.Close()
	defer grpc.AuthClientConn.Conn.Close()
	defer grpc.ProductClientConn.Conn.Close()

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
