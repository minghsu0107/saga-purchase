package repo

import (
	"context"
	"time"

	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/domain/model"
	"github.com/minghsu0107/saga-purchase/infra/grpc"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

// AuthRepository is the auth repository interface
type AuthRepository interface {
	Auth(accessToken string) (*model.AuthResult, error)
}

// AuthRepositoryImpl is the implementation of AuthRepository
type AuthRepositoryImpl struct {
	ctx  context.Context
	auth endpoint.Endpoint
}

// NewAuthRepository is the factory of AuthRepository
func NewAuthRepository(conn *grpc.AuthConn, config *conf.Config) AuthRepository {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), config.ServiceOptions.Rps))

	var auth endpoint.Endpoint
	{
		auth = grpctransport.NewClient(
			conn.Conn,
			"auth.AuthService",
			"Auth",
			encodeGRPCRequest,
			decodeGRPCResponse,
			&pb.AuthResponse{},
		).Endpoint()
		auth = limiter(auth)
		auth = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "auth",
			Timeout: config.ServiceOptions.Timeout,
		}))(auth)
	}

	return &AuthRepositoryImpl{
		ctx:  context.Background(),
		auth: auth,
	}
}

// Auth method implements AuthRepository interface
func (repo *AuthRepositoryImpl) Auth(accessToken string) (*model.AuthResult, error) {
	res, err := repo.auth(repo.ctx, &pb.AuthPayload{
		AccessToken: accessToken,
	})
	if err != nil {
		return nil, err
	}
	response := res.(*pb.AuthResponse)
	return &model.AuthResult{
		CustomerID: response.CustomerId,
		Active:     response.Active,
		Expired:    response.Expired,
	}, nil
}
