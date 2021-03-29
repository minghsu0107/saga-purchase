package result

import (
	"net/http"

	pb "github.com/minghsu0107/saga-pb"
	"github.com/minghsu0107/saga-purchase/domain/event"
)

// PurchaseResultService is the interface of purchase result service
type PurchaseResultService interface {
	MapPurchaseResult(purchaseResult *pb.PurchaseResult) *event.PurchaseResult
	GetPurchaseResult(req *http.Request) (*event.PurchaseResult, error)
}
