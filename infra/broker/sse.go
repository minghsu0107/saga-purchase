package broker

import (
	watermillHTTP "github.com/ThreeDotsLabs/watermill-http/pkg/http"
	"github.com/ThreeDotsLabs/watermill/message"
)

// NewSSERouter returns a server-sent-events router
func NewSSERouter(subsciber message.Subscriber) (*watermillHTTP.SSERouter, error) {
	sseRouter, err := watermillHTTP.NewSSERouter(
		watermillHTTP.SSERouterConfig{
			UpstreamSubscriber: subsciber,
			ErrorHandler:       watermillHTTP.DefaultErrorHandler,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &sseRouter, nil
}
