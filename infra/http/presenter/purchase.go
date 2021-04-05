package presenter

// PurchaseResult is the HTTP JSON response of purchase result
type PurchaseResult struct {
	Step      string `json:"step"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

// CartItem is the JSON request that represents an order
type CartItem struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	Amount    int64  `json:"amount" binding:"required,number,min=1"`
}

// Payment is the JSON request that represents a payment
type Payment struct {
	CurrencyCode string `json:"currency_code" binding:"required,oneof=NT US"`
}

// Purchase is the HTTP JSON request of creating new purchase
type Purchase struct {
	CartItems *[]CartItem `json:"purchase_items" binding:"min=1"`
	Payment   *Payment    `json:"payment"`
}
