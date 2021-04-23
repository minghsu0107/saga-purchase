package http

import (
	"context"
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
	customerID := r.Context().Value(config.CustomerKey).(uint64)
	if customerID == purchaseResult.CustomerId {
		r = r.WithContext(context.WithValue(r.Context(), config.MsgKey, h.PurchaseResultSvc.MapPurchaseResult(purchaseResult)))
		ok = true
	}
	return
}

// GetResponse is the http handler that generates SSE responses
func (h *PurchaseResultStreamHandler) GetResponse(w http.ResponseWriter, r *http.Request) (response interface{}, ok bool) {
	purchaseResult, err := h.PurchaseResultSvc.GetPurchaseResult(r)
	if err != nil {
		return nil, false
	}
	if purchaseResult == nil {
		return nil, true
	}
	return &presenter.PurchaseResult{
		Step:      purchaseResult.Step,
		Status:    purchaseResult.Status,
		Timestamp: time.Now().Local().Unix(),
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
		response(c, http.StatusUnauthorized, presenter.ErrUnautorized)
		return
	}
	err := h.PurchasingSvc.CreatePurchase(c.Request.Context(), customerID, &curPurchase)
	switch err {
	case purchase.ErrProductNotfound:
		response(c, http.StatusNotFound, purchase.ErrProductNotfound)
	case nil:
		c.JSON(http.StatusCreated, presenter.OkMsg)
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
