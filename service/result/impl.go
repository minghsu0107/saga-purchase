package result

import (
	"fmt"
	"net/http"

	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/domain/event"
	log "github.com/sirupsen/logrus"
)

// PurchaseResultServiceImpl implements PurchaseResultService interface
type PurchaseResultServiceImpl struct {
	logger *log.Entry
}

// NewPurchaseResultService is the factory of PurchaseResultServiceImpl
func NewPurchaseResultService(config *conf.Config) PurchaseResultService {
	return &PurchaseResultServiceImpl{
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:PurchaseResultService",
		}),
	}
}

// MapPurchaseResult maps protobuf purchase result to a purchase result domain entity
func (svc *PurchaseResultServiceImpl) MapPurchaseResult(purchaseResult *pb.PurchaseResult) *event.PurchaseResult {
	step := getPurchaseStep(purchaseResult.Step)
	status := getPurchaseStatus(purchaseResult.Status)
	svc.logger.WithFields(log.Fields{
		"step":   step,
		"status": status,
	}).Info("new purchase result")
	return &event.PurchaseResult{
		Step:   step,
		Status: status,
	}
}

// GetPurchaseResult retrieves purchase result from http request context
func (svc *PurchaseResultServiceImpl) GetPurchaseResult(req *http.Request) (*event.PurchaseResult, error) {
	purchaseResult, ok := req.Context().Value(conf.MsgKey).(*event.PurchaseResult)
	if !ok {
		return nil, fmt.Errorf("error when casting purchase result")
	}
	return purchaseResult, nil
}

func getPurchaseStep(step pb.PurchaseStep) string {
	switch step {
	case pb.PurchaseStep_STEP_UPDATE_PRODUCT_INVENTORY:
		return "UPDATE_PRODUCT_INVENTORY"
	case pb.PurchaseStep_STEP_CREATE_ORDER:
		return "CREATE_ORDER"
	case pb.PurchaseStep_STEP_CREATE_PAYMENT:
		return "CREATE_PAYMENT"
	}
	return ""
}

func getPurchaseStatus(status pb.PurchaseStatus) string {
	switch status {
	case pb.PurchaseStatus_STATUS_EXUCUTE:
		return "STATUS_EXUCUTE"
	case pb.PurchaseStatus_STATUS_SUCCESS:
		return "STATUS_SUCCESS"
	case pb.PurchaseStatus_STATUS_FAILED:
		return "STATUS_FAILED"
	case pb.PurchaseStatus_STATUS_ROLLBACKED:
		return "STATUS_ROLLBACKED"
	case pb.PurchaseStatus_STATUS_ROLLBACK_FAIL:
		return "STATUS_ROLLBACK_FAIL"
	}
	return ""
}
