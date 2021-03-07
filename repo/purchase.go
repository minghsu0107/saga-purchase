package repo

import (
	"encoding/json"
	"time"

	pb "github.com/minghsu0107/saga-pb"
	"github.com/minghsu0107/saga-purchase/config"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/golang/protobuf/ptypes"
	"github.com/minghsu0107/saga-purchase/domain/model"
)

// PurchasingRepository is the repository interface of purchase aggregate
type PurchasingRepository interface {
	CreatePurchase(purchase *model.Purchase) error
}

// PurchasingRepositoryImpl is the repository implementation of purchase aggregate
type PurchasingRepositoryImpl struct {
	publisher message.Publisher
}

// NewPurchasingRepository is the factory of PurchaseRepository
func NewPurchasingRepository(publisher message.Publisher) PurchasingRepository {
	return &PurchasingRepositoryImpl{
		publisher: publisher,
	}
}

// CreatePurchase publish a CreatePurchase command to the message broker
func (r *PurchasingRepositoryImpl) CreatePurchase(purchase *model.Purchase) error {
	var purchasedItems []*pb.PurchasedItem
	for _, cartItem := range *purchase.Order.CartItems {
		purchasedItems = append(purchasedItems, &pb.PurchasedItem{
			ProductId: cartItem.ProductID,
			Amount:    cartItem.Amount,
		})
	}
	pbPurchase := &pb.Purchase{
		Order: &pb.Order{
			CustomerId:     purchase.Order.CustomerID,
			PurchasedItems: purchasedItems,
		},
		Payment: &pb.Payment{
			CurrencyCode: purchase.Payment.CurrencyCode,
		},
	}
	curTime, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return err
	}
	createPurchaseCommand := &pb.CreatePurchase{
		Purchase:  pbPurchase,
		Timestamp: curTime,
	}
	payload, err := json.Marshal(createPurchaseCommand)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(watermill.NewUUID(), msg)

	if err := r.publisher.Publish(config.PurchaseTopic, msg); err != nil {
		return err
	}
	return nil
}
