package purchase

import (
	"context"

	"github.com/minghsu0107/saga-purchase/domain/model"
	"github.com/minghsu0107/saga-purchase/infra/http/presenter"
)

// PurchasingService is the interface of purchasing service
type PurchasingService interface {
	CheckProducts(ctx context.Context, cartItems *[]model.CartItem) (*[]model.ProductStatus, error)
	CreatePurchase(ctx context.Context, customerID uint64, purchase *presenter.Purchase) error
}
