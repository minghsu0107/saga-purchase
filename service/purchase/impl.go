package purchase

import (
	"context"

	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/domain/model"
	"github.com/minghsu0107/saga-purchase/infra/http/presenter"
	"github.com/minghsu0107/saga-purchase/repo"
	log "github.com/sirupsen/logrus"
)

// PurchasingServiceImpl implements PurchasingService interface
type PurchasingServiceImpl struct {
	logger         *log.Entry
	purchasingRepo repo.PurchasingRepository
	productRepo    repo.ProductRepository
}

// NewPurchasingService is the factory of PurchasingService
func NewPurchasingService(config *conf.Config, purchasingRepo repo.PurchasingRepository, productRepo repo.ProductRepository) PurchasingService {
	return &PurchasingServiceImpl{
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:PurchasingService",
		}),
		purchasingRepo: purchasingRepo,
		productRepo:    productRepo,
	}
}

// CheckProduct checks the product status
func (svc *PurchasingServiceImpl) CheckProducts(ctx context.Context, cartItems *[]model.CartItem) (*[]model.ProductStatus, error) {
	for _, cartcartItem := range *cartItems {
		if cartcartItem.Amount <= 0 {
			return nil, ErrInvalidCartItemAmount
		}
	}
	productStatuses, err := svc.productRepo.CheckProducts(ctx, cartItems)
	if err != nil {
		svc.logger.Error(err.Error())
		return nil, err
	}
	for _, productStatus := range *productStatuses {
		switch productStatus.Status {
		case model.ProductOk:
			continue
		case model.ProductNotFound:
			return nil, ErrProductNotfound
		default:
			return nil, ErrUnkownProductStatus
		}
	}
	return productStatuses, nil
}

// CreatePurchase passes a CreatePurchase command to orchestrator
func (svc *PurchasingServiceImpl) CreatePurchase(ctx context.Context, customerID uint64, purchase *presenter.Purchase) error {
	var cartItems []model.CartItem
	for _, cartItem := range *(purchase.CartItems) {
		cartItems = append(cartItems, model.CartItem{
			ProductID: cartItem.ProductID,
			Amount:    cartItem.Amount,
		})
	}
	productStatuses, err := svc.CheckProducts(ctx, &cartItems)
	if err != nil {
		return err
	}
	var amount int64 = 0
	for i, productStatus := range *productStatuses {
		amount += cartItems[i].Amount * productStatus.Price
	}
	newPurchase := &model.Purchase{
		Order: &model.Order{
			CustomerID: customerID,
			CartItems:  &cartItems,
		},
		Payment: &model.Payment{
			CurrencyCode: purchase.Payment.CurrencyCode,
			Amount:       amount,
		},
	}

	if err := svc.purchasingRepo.CreatePurchase(newPurchase); err != nil {
		svc.logger.Error(err.Error())
		return err
	}
	return nil
}
