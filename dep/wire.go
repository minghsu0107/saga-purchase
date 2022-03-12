//+build wireinject

// The build tag makes sure the stub is not built in the final build.
package dep

import (
	"github.com/google/wire"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/infra"
	infra_broker "github.com/minghsu0107/saga-purchase/infra/broker"
	infra_grpc "github.com/minghsu0107/saga-purchase/infra/grpc"
	infra_http "github.com/minghsu0107/saga-purchase/infra/http"
	"github.com/minghsu0107/saga-purchase/infra/http/middleware"
	infra_observe "github.com/minghsu0107/saga-purchase/infra/observe"
	"github.com/minghsu0107/saga-purchase/pkg"
	"github.com/minghsu0107/saga-purchase/repo"
	"github.com/minghsu0107/saga-purchase/service/purchase"
	"github.com/minghsu0107/saga-purchase/service/result"
)

func InitializeServer() (*infra.Server, error) {
	wire.Build(
		conf.NewConfig,

		infra.NewServer,

		infra_http.NewServer,
		infra_http.NewEngine,
		infra_http.NewRouter,
		infra_http.NewPurchaseResultStreamHandler,
		infra_http.NewPurchasingHandler,

		infra_observe.NewObservabilityInjector,

		middleware.NewJWTAuthChecker,

		infra_grpc.NewAuthConn,
		infra_grpc.NewProductConn,

		infra_broker.NewSSERouter,
		infra_broker.NewRedisSubscriber,
		infra_broker.NewNATSPublisher,

		result.NewPurchaseResultService,
		purchase.NewPurchasingService,

		pkg.NewSonyFlake,

		repo.NewAuthRepository,
		repo.NewPurchasingRepository,
		repo.NewProductRepository,
	)
	return &infra.Server{}, nil
}
