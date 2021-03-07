package model

// Order entity
type Order struct {
	CustomerID uint64
	CartItems  *[]CartItem
}

// CartItem entity
type CartItem struct {
	ProductID uint64
	Amount    int64
}
