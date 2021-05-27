package purchase

import "errors"

var (
	// ErrInvalidCartItemAmount is invalid cart item amount error
	ErrInvalidCartItemAmount = errors.New("invalid cart item amount")
	// ErrProductNotfound is product not found error
	ErrProductNotfound = errors.New("product not found")
	// ErrUnkownProductStatus unkown product status error
	ErrUnkownProductStatus = errors.New("unknown product status")
)
