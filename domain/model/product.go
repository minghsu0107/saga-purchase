package model

// ProductDetail value object
type ProductDetail struct {
	ProductName string
	Description string
	BrandName   string
	Inventory   int64
}

// ProductStatus enumeration
type ProductStatus int

const (
	// ProductOk is ok status
	ProductOk ProductStatus = iota
	// ProductNotFound is not found status
	ProductNotFound
)
