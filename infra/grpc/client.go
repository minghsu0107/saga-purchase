package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/sercand/kuberesolver/v3"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
)

var (
	// AuthClientConn grpc connection
	AuthClientConn *AuthConn
	// ProductClientConn grpc connection
	ProductClientConn *ProductConn

	// KubernetesResolver name
	KubernetesResolver string = "kubernetes"
	once               sync.Once
)

// AuthConn is a wrapper for Auth grpc connection
type AuthConn struct {
	Conn *grpc.ClientConn
}

// ProductConn is a wrapper for Product grpc connection
type ProductConn struct {
	Conn *grpc.ClientConn
}

// NewAuthConn returns a grpc client connection for AuthRepository
func NewAuthConn(config *conf.Config) (*AuthConn, error) {
	log.Info("connecting to grpc auth service...")
	conn, err := newGRPCConn(config.Resolver, config.RPCEndpoints.AuthSvcHost, config.OcAgentHost)
	if err != nil {
		return nil, err
	}
	AuthClientConn = &AuthConn{
		Conn: conn,
	}
	return AuthClientConn, nil
}

// NewProductConn returns a grpc client connection for ProductRepository
func NewProductConn(config *conf.Config) (*ProductConn, error) {
	log.Info("connecting to grpc product service...")
	conn, err := newGRPCConn(config.Resolver, config.RPCEndpoints.ProductSvcHost, config.OcAgentHost)
	if err != nil {
		return nil, err
	}
	ProductClientConn = &ProductConn{
		Conn: conn,
	}
	return ProductClientConn, nil
}

func newGRPCConn(resolver, svcHost, ocAgentHost string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var scheme string

	if resolver == KubernetesResolver {
		once.Do(kuberesolver.RegisterInCluster)
		scheme = "kubernetes"
	} else {
		scheme = "dns"
	}

	retryOpts := []grpc_retry.CallOption{
		// generate waits between 900ms to 1100ms
		grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(1*time.Second, 0.1)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
	}

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	if ocAgentHost != "" {
		dialOpts = append(dialOpts, grpc.WithStatsHandler(new(ocgrpc.ClientHandler)))
	}

	dialOpts = append(dialOpts,
		grpc.WithDisableServiceConfig(),
		grpc.WithDefaultServiceConfig(`{
			"loadBalancingPolicy": "round_robin"
		}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		}),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
		//grpc.WithBlock(),
	)

	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s:///%s", scheme, svcHost),
		dialOpts...,
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
