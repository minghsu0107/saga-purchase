package repo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/domain/model"

	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PurchasingRepository is the repository interface of purchase aggregate
type PurchasingRepository interface {
	CreatePurchase(ctx context.Context, purchase *model.Purchase) error
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
func (r *PurchasingRepositoryImpl) CreatePurchase(ctx context.Context, purchase *model.Purchase) error {
	tr := otel.Tracer("createPurchase")
	ctx, span := tr.Start(ctx, "event.CreatePurchase")
	defer span.End()

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
			Amount:       purchase.Payment.Amount,
		},
	}

	curTime := timestamppb.New(time.Now())
	createPurchaseCommand := &pb.CreatePurchaseCmd{
		PurchaseId: purchase.ID,
		Purchase:   pbPurchase,
		Timestamp:  curTime,
	}
	payload, err := json.Marshal(createPurchaseCommand)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set(conf.SpanContextKey, spanContextToW3C(ctx))
	middleware.SetCorrelationID(watermill.NewUUID(), msg)

	if err := r.publisher.Publish(conf.PurchaseTopic, msg); err != nil {
		return err
	}
	return nil
}
