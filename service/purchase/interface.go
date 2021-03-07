package purchase

import (
	"github.com/minghsu0107/saga-purchase/domain/model"
	"github.com/minghsu0107/saga-purchase/infra/http/presenter"
)

// PurchasingService is the interface of purchasing service
type PurchasingService interface {
	CheckProduct(cartItems *[]model.CartItem) error
	CreatePurchase(customerID uint64, purchase *presenter.Purchase) error
}
