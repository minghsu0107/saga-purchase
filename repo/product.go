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
	CheckProduct(cartItems *[]model.CartItem) (*[]model.ProductStatus, error)
}

// ProductRepositoryImpl is the implementation of ProductRepository
type ProductRepositoryImpl struct {
	ctx          context.Context
	checkProduct endpoint.Endpoint
}

// NewProductRepository is the factory of AuthRepository
func NewProductRepository(conn *grpc.ProductConn, config *conf.Config) ProductRepository {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), config.ServiceOptions.Rps))

	var options []grpctransport.ClientOption

	var checkProduct endpoint.Endpoint
	{
		svcName := "product.ProductService"
		checkProduct = grpctransport.NewClient(
			conn.Conn,
			svcName,
			"CheckProduct",
			encodeGRPCRequest,
			decodeGRPCResponse,
			&pb.CheckProductsResponse{},
			append(options, grpctransport.ClientBefore(grpctransport.SetRequestHeader(ServiceNameHeader, svcName)))...,
		).Endpoint()
		checkProduct = limiter(checkProduct)
		checkProduct = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "product",
			Timeout: config.ServiceOptions.Timeout,
		}))(checkProduct)
	}

	return &ProductRepositoryImpl{
		ctx:          context.Background(),
		checkProduct: checkProduct,
	}
}

// CheckProduct method implements ProductRepository interface
func (svc *ProductRepositoryImpl) CheckProduct(cartItems *[]model.CartItem) (*[]model.ProductStatus, error) {
	var pbCartItems []*pb.CartItem
	for _, cartItem := range *cartItems {
		pbCartItems = append(pbCartItems, &pb.CartItem{
			ProductId: cartItem.ProductID,
			Amount:    cartItem.Amount,
		})
	}
	res, err := svc.checkProduct(svc.ctx, &pb.CheckProductsRequest{
		CartItems: pbCartItems,
	})
	if err != nil {
		return nil, err
	}
	checkProductsResponse := res.(*pb.CheckProductsResponse)
	var productStatuses []model.ProductStatus
	for _, productStatus := range checkProductsResponse.ProductStatuses {
		productStatuses = append(productStatuses, getProductStatus(productStatus.Status))
	}
	return &productStatuses, nil
}

func getProductStatus(status pb.Status) model.ProductStatus {
	switch status {
	case pb.Status_STATUS_OK:
		return model.ProductOk
	case pb.Status_STATUS_NOT_FOUND:
		return model.ProductNotFound
	}
	return -1
}
