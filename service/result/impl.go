package result

import (
	pb "github.com/minghsu0107/saga-pb"
	conf "github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/domain/event"
	"github.com/minghsu0107/saga-purchase/repo"
	log "github.com/sirupsen/logrus"
)

// PurchaseResultServiceImpl implements PurchaseResultService interface
type PurchaseResultServiceImpl struct {
	logger *log.Entry
	repo   repo.PurchaseResultRepository
}

// NewPurchaseResultService is the factory of PurchaseResultServiceImpl
func NewPurchaseResultService(config *conf.Config, repo repo.PurchaseResultRepository) PurchaseResultService {
	return &PurchaseResultServiceImpl{
		repo: repo,
		logger: config.Logger.ContextLogger.WithFields(log.Fields{
			"type": "service:PurchaseResultService",
		}),
	}
}

// SetPurchaseResult save purchase result to local cache
func (svc *PurchaseResultServiceImpl) SetPurchaseResult(purchaseResult *pb.PurchaseResult) error {
	step := getPurchaseStep(purchaseResult.Step)
	status := getPurchaseStatus(purchaseResult.Status)
	svc.logger.WithFields(log.Fields{
		"step":   step,
		"status": status,
	}).Info("new purchase result")
	err := svc.repo.SetPurchaseResult(purchaseResult.CustomerId, &event.PurchaseResult{
		Step:   step,
		Status: status,
	})
	if err != nil {
		svc.logger.Error(err)
		return err
	}
	return nil
}

// GetPurchaseResult retrives a purchase result by customer ID
func (svc *PurchaseResultServiceImpl) GetPurchaseResult(customerID uint64) (*event.PurchaseResult, error) {
	purchaseResult, err := svc.repo.GetPurchaseResult(customerID)
	if err != nil {
		svc.logger.Error(err)
		return nil, err
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
