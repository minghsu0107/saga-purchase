package event

import (
	"time"
)

// PurchaseResult event
type PurchaseResult struct {
	Step      string
	Status    string
	Timestamp time.Time
}
