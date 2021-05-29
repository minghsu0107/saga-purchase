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

// ProductRepository is the product repository interface
type ProductRepository interface {
	CheckProducts(ctx context.Context, cartItems *[]model.CartItem) (*[]model.ProductStatus, error)
}

// ProductRepositoryImpl is the implementation of ProductRepository
type ProductRepositoryImpl struct {
	checkProducts endpoint.Endpoint
}

// NewProductRepository is the factory of AuthRepository
func NewProductRepository(conn *grpc.ProductConn, config *conf.Config) ProductRepository {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), config.ServiceOptions.Rps))

	var options []grpctransport.ClientOption

	var checkProducts endpoint.Endpoint
	{
		svcName := "product.ProductService"
		checkProducts = grpctransport.NewClient(
			conn.Conn,
			svcName,
			"CheckProducts",
			encodeGRPCRequest,
			decodeGRPCResponse,
			&pb.CheckProductsResponse{},
			append(options, grpctransport.ClientBefore(grpctransport.SetRequestHeader(ServiceNameHeader, svcName)))...,
		).Endpoint()
		checkProducts = limiter(checkProducts)
		checkProducts = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "product",
			Timeout: config.ServiceOptions.Timeout,
		}))(checkProducts)
	}

	return &ProductRepositoryImpl{
		checkProducts: checkProducts,
	}
}

// CheckProduct method implements ProductRepository interface
func (r *ProductRepositoryImpl) CheckProducts(ctx context.Context, cartItems *[]model.CartItem) (*[]model.ProductStatus, error) {
	var pbCartItems []*pb.CartItem
	for _, cartItem := range *cartItems {
		pbCartItems = append(pbCartItems, &pb.CartItem{
			ProductId: cartItem.ProductID,
			Amount:    cartItem.Amount,
		})
	}
	res, err := r.checkProducts(ctx, &pb.CheckProductsRequest{
		CartItems: pbCartItems,
	})
	if err != nil {
		return nil, err
	}
	checkProductsResponse := res.(*pb.CheckProductsResponse)
	var productStatuses []model.ProductStatus
	for _, productStatus := range checkProductsResponse.ProductStatuses {
		productStatuses = append(productStatuses, model.ProductStatus{
			ProductID: productStatus.ProductId,
			Price:     productStatus.Price,
			Status:    getProductStatus(productStatus.Status),
		})
	}
	return &productStatuses, nil
}

func getProductStatus(status pb.Status) model.Status {
	switch status {
	case pb.Status_STATUS_OK:
		return model.ProductOk
	case pb.Status_STATUS_NOT_FOUND:
		return model.ProductNotFound
	}
	return -1
}
