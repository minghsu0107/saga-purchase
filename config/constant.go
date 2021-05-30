package config

type HTTPContextKey string

const (
	// JWTAuthHeader is the auth header containing customer ID
	JWTAuthHeader = "Authorization"
	// CustomerKey is the key name for retrieving jwt-decoded customer id in a http request context
	CustomerKey HTTPContextKey = "customer_key"

	// SpanContextKey is the message metadata key of span context passed accross process boundaries
	SpanContextKey = "span_ctx_key"
	// PurchaseTopic is the topic to which we publish new purchase
	PurchaseTopic = "purchase"
	// PurchaseResultTopic is the subscribed topic for purchase result
	PurchaseResultTopic = "purchase.result"
)
