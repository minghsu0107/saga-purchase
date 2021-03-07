package repo

import (
	"encoding/json"
	"strconv"

	"github.com/allegro/bigcache/v3"
	"github.com/minghsu0107/saga-purchase/domain/event"
)

// PurchaseResultRepository defines interface for caching purchase result per customer
type PurchaseResultRepository interface {
	SetPurchaseResult(customerID uint64, purchaseResult *event.PurchaseResult) error
	GetPurchaseResult(customerID uint64) (*event.PurchaseResult, error)
}

// PurchaseResultRepositoryImpl is the implementation of PurchaseResultRepository
type PurchaseResultRepositoryImpl struct {
	cache *bigcache.BigCache
}

// NewPurchaseResultRepository is the factory of PurchaseResultRepository
func NewPurchaseResultRepository(cache *bigcache.BigCache) (PurchaseResultRepository, error) {
	return &PurchaseResultRepositoryImpl{
		cache: cache,
	}, nil
}

// SetPurchaseResult sets a key-value pair (customerID:purchaseResult) to local cache
func (repo *PurchaseResultRepositoryImpl) SetPurchaseResult(customerID uint64, purchaseResult *event.PurchaseResult) error {
	serializedRes, err := json.Marshal(purchaseResult)
	if err != nil {
		return err
	}
	if err = repo.cache.Set(strconv.FormatUint(customerID, 10), serializedRes); err != nil {
		return err
	}
	return nil
}

// GetPurchaseResult retrieves purchaseResult by customerID
func (repo *PurchaseResultRepositoryImpl) GetPurchaseResult(customerID uint64) (*event.PurchaseResult, error) {
	serializedRes, err := repo.cache.Get(strconv.FormatUint(customerID, 10))
	if err != nil {
		return nil, err
	}
	var purchaseResult event.PurchaseResult
	if err = json.Unmarshal(serializedRes, &purchaseResult); err != nil {
		return nil, err
	}
	return &purchaseResult, nil
}
