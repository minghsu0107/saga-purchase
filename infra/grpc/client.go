package grpc

import (
	"context"
	"fmt"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/sercand/kuberesolver/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/codes"
)

var (
	AuthClientConn    *AuthConn
	ProductClientConn *ProductConn

	KubernetesProvider string = "kubernetes"
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
	conn, err := newGRPCConn(config.Provider, config.RPCEndpoints.AuthSvcHost)
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
	conn, err := newGRPCConn(config.Provider, config.RPCEndpoints.ProductSvcHost)
	if err != nil {
		return nil, err
	}
	ProductClientConn = &ProductConn{
		Conn: conn,
	}
	return ProductClientConn, nil
}

func newGRPCConn(provider, serverURL string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var scheme string

	if provider == KubernetesProvider {
		kuberesolver.RegisterInCluster()
		scheme = "kubernetes"
	} else {
		scheme = "dns"
	}

	retryOpts := []grpc_retry.CallOption{
		// generate waits between 900ms to 1100ms
		grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(1*time.Second, 0.1)),
		// retry only on NotFound and Unavailable
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
	}

	conn, err := grpc.DialContext(
		ctx,
		fmt.Sprintf("%s///%s", scheme, serverURL),
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
		//grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
