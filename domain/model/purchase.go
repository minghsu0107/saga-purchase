package model

// Purchase aggregate
type Purchase struct {
	ID      uint64
	Order   *Order
	Payment *Payment
}
