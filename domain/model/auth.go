package model

// AuthResult value object
type AuthResult struct {
	CustomerID uint64
	Active     bool
	Expired    bool
}
