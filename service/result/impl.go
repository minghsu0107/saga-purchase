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
	purchaseID := purchaseResult.PurchaseId
	step := getPurchaseStep(purchaseResult.Step)
	status := getPurchaseStatus(purchaseResult.Status)
	svc.logger.WithFields(log.Fields{
		"purchase_id": purchaseID,
		"step":        step,
		"status":      status,
	}).Info("new purchase result")
	return &event.PurchaseResult{
		PurchaseID: purchaseID,
		Step:       step,
		Status:     status,
	}
}

// GetPurchaseResult retrieves purchase result from http request context
func (svc *PurchaseResultServiceImpl) GetPurchaseResult(req *http.Request) (*event.PurchaseResult, error) {
	val := req.Context().Value(conf.MsgKey)
	if val == nil {
		return nil, nil
	}
	purchaseResult, ok := val.(*event.PurchaseResult)
	if !ok {
		return nil, fmt.Errorf("error when casting purchase result")
	}
	return purchaseResult, nil
}

func getPurchaseStep(step pb.PurchaseStep) string {
	switch step {
	case pb.PurchaseStep_STEP_UPDATE_PRODUCT_INVENTORY:
		return event.StepUpdateProductInventory
	case pb.PurchaseStep_STEP_CREATE_ORDER:
		return event.StepCreateOrder
	case pb.PurchaseStep_STEP_CREATE_PAYMENT:
		return event.StepCreatePayment
	}
	return ""
}

func getPurchaseStatus(status pb.PurchaseStatus) string {
	switch status {
	case pb.PurchaseStatus_STATUS_EXUCUTE:
		return event.StatusExecute
	case pb.PurchaseStatus_STATUS_SUCCESS:
		return event.StatusSucess
	case pb.PurchaseStatus_STATUS_FAILED:
		return event.StatusFailed
	case pb.PurchaseStatus_STATUS_ROLLBACKED:
		return event.StatusRollbacked
	case pb.PurchaseStatus_STATUS_ROLLBACK_FAIL:
		return event.StatusRollbackFailed
	}
	return ""
}
