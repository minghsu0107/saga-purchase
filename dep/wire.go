//+build wireinject

// The build tag makes sure the stub is not built in the final build.
package dep

import (
	"github.com/google/wire"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/infra/broker"
	"github.com/minghsu0107/saga-purchase/infra/cache"
	"github.com/minghsu0107/saga-purchase/infra/grpc"
	"github.com/minghsu0107/saga-purchase/infra/http"
	"github.com/minghsu0107/saga-purchase/infra/http/middleware"
	"github.com/minghsu0107/saga-purchase/repo"
	"github.com/minghsu0107/saga-purchase/service/purchase"
	"github.com/minghsu0107/saga-purchase/service/result"
)

func InitializeServer() (*http.Server, error) {
	wire.Build(
		conf.NewConfig,

		http.NewServer,
		http.NewEngine,
		http.NewRouter,
		http.NewPurchaseResultStreamHandler,
		http.NewPurchasingHandler,

		middleware.NewJWTAuthChecker,

		grpc.NewAuthConn,
		grpc.NewProductConn,

		cache.NewLocalCache,

		broker.NewSSERouter,
		broker.NewNATSSubscriber,
		broker.NewNATSPublisher,

		result.NewPurchaseResultService,
		purchase.NewPurchasingService,

		repo.NewAuthRepository,
		repo.NewPurchaseResultRepository,
		repo.NewPurchasingRepository,
		repo.NewProductRepository,
	)
	return &http.Server{}, nil
}
