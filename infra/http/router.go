package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	pb "github.com/minghsu0107/saga-pb"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/minghsu0107/saga-purchase/config"
	"github.com/minghsu0107/saga-purchase/infra/http/presenter"
	"github.com/minghsu0107/saga-purchase/service/purchase"
	"github.com/minghsu0107/saga-purchase/service/result"
)

// Router wraps http handlers
type Router struct {
	PurchaseResultStreamHandler *PurchaseResultStreamHandler
	PurchasingHandler           *PurchasingHandler
}

// NewRouter is a factory for router instance
func NewRouter(purchaseResultStreamHandler *PurchaseResultStreamHandler, purchasingHandler *PurchasingHandler) *Router {
	return &Router{
		PurchaseResultStreamHandler: purchaseResultStreamHandler,
		PurchasingHandler:           purchasingHandler,
	}
}

// PurchaseResultStreamHandler handles SSE stream
type PurchaseResultStreamHandler struct {
	PurchaseResultSvc result.PurchaseResultService
}

// NewPurchaseResultStreamHandler is the factory of PurchaseResultStreamHandler
func NewPurchaseResultStreamHandler(purchaseResultSvc result.PurchaseResultService) *PurchaseResultStreamHandler {
	return &PurchaseResultStreamHandler{
		PurchaseResultSvc: purchaseResultSvc,
	}
}

// Validate determine whether we should process the incoming message for the current http request
func (h *PurchaseResultStreamHandler) Validate(r *http.Request, msg *message.Message) (ok bool) {
	ok = false
	purchaseResult := &pb.PurchaseResult{}
	err := json.Unmarshal(msg.Payload, purchaseResult)
	if err != nil {
		return
	}
	customerID, valid := r.Context().Value(config.CustomerKey).(uint64)
	if !valid {
		return
	}

	if customerID == purchaseResult.CustomerId {
		ok = true
	}
	return
}

// GetResponse is the http handler that generates SSE responses
func (h *PurchaseResultStreamHandler) GetResponse(w http.ResponseWriter, r *http.Request, msg *message.Message) (response interface{}, ok bool) {
	if msg == nil {
		return nil, true
	}

	pbPurchaseResult := &pb.PurchaseResult{}
	err := json.Unmarshal(msg.Payload, pbPurchaseResult)
	if err != nil {
		return nil, false
	}
	purchaseResult := h.PurchaseResultSvc.MapPurchaseResult(pbPurchaseResult)

	if purchaseResult == nil {
		return nil, true
	}

	return &presenter.PurchaseResult{
		PurchaseID: purchaseResult.PurchaseID,
		Step:       purchaseResult.Step,
		Status:     purchaseResult.Status,
		Timestamp:  time.Now().Local().Unix(),
	}, true
}

// PurchasingHandler handles purchasing http endpoints
type PurchasingHandler struct {
	PurchasingSvc purchase.PurchasingService
}

// NewPurchasingHandler is the factory of PurchasingHandler
func NewPurchasingHandler(purchasingSvc purchase.PurchasingService) *PurchasingHandler {
	return &PurchasingHandler{
		PurchasingSvc: purchasingSvc,
	}
}

// CreatePurchase is the http handler that creates a purchase
func (h *PurchasingHandler) CreatePurchase(c *gin.Context) {
	var curPurchase presenter.Purchase
	if err := c.ShouldBindJSON(&curPurchase); err != nil {
		response(c, http.StatusBadRequest, presenter.ErrInvalidParam)
		return
	}
	customerID, ok := c.Request.Context().Value(config.CustomerKey).(uint64)
	if !ok {
		response(c, http.StatusUnauthorized, presenter.ErrUnauthorized)
		return
	}
	purchaseID, err := h.PurchasingSvc.CreatePurchase(c.Request.Context(), customerID, &curPurchase)
	switch err {
	case purchase.ErrInvalidCartItemAmount, purchase.ErrUnkownProductStatus:
		response(c, http.StatusBadRequest, presenter.ErrInvalidParam)
		return
	case purchase.ErrProductNotfound:
		response(c, http.StatusNotFound, purchase.ErrProductNotfound)
	case nil:
		c.JSON(http.StatusCreated, &presenter.PurchaseCreation{
			PurchaseID: purchaseID,
		})
		return
	default:
		response(c, http.StatusInternalServerError, presenter.ErrServer)
		return
	}
}

func response(c *gin.Context, httpCode int, err error) {
	message := err.Error()
	c.JSON(httpCode, presenter.ErrResponse{
		Message: message,
	})
}
