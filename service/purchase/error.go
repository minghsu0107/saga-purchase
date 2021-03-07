package purchase

import "errors"

var (
	// ErrProductNotfound is product not found error
	ErrProductNotfound = errors.New("product not found")
	// ErrUnkownProductStatus unkown product status error
	ErrUnkownProductStatus = errors.New("unknown product status")
)
