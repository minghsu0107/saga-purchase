package grpc

import (
	"context"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	conf "github.com/minghsu0107/saga-purchase/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
	conn, err := newGRPCConn(config.RPCEndpoints.AuthServerURL)
	if err != nil {
		return nil, err
	}
	return &AuthConn{
		Conn: conn,
	}, nil
}

// NewProductConn returns a grpc client connection for ProductRepository
func NewProductConn(config *conf.Config) (*ProductConn, error) {
	conn, err := newGRPCConn(config.RPCEndpoints.ProductServerURL)
	if err != nil {
		return nil, err
	}
	return &ProductConn{
		Conn: conn,
	}, nil
}

func newGRPCConn(serverURL string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	retryOpts := []grpc_retry.CallOption{
		// generate waits between 900ms to 1100ms
		grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(1*time.Second, 0.1)),
		// retry only on NotFound and Unavailable
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
	}
	conn, err := grpc.DialContext(
		ctx,
		serverURL,
		grpc.WithInsecure(),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
