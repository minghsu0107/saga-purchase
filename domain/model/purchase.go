package model

// Purchase aggregate
type Purchase struct {
	Order   *Order
	Payment *Payment
}
