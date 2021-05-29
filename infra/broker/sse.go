package broker

import (
	pkg "github.com/minghsu0107/saga-purchase/pkg"

	"github.com/ThreeDotsLabs/watermill/message"
)

// NewSSERouter returns a server-sent-events router
func NewSSERouter(subsciber message.Subscriber) (*pkg.SSERouter, error) {
	sseRouter, err := pkg.NewSSERouter(
		pkg.SSERouterConfig{
			UpstreamSubscriber: subsciber,
			ErrorHandler:       pkg.DefaultErrorHandler,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &sseRouter, nil
}
