package broker

import (
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/ThreeDotsLabs/watermill-nats/pkg/nats"
	conf "github.com/minghsu0107/saga-purchase/config"
	stan "github.com/nats-io/stan.go"
)

// NewNATSPublisher returns a NATS publisher for event streaming
func NewNATSPublisher(config *conf.Config) (message.Publisher, error) {
	publisher, err := nats.NewStreamingPublisher(
		nats.StreamingPublisherConfig{
			ClusterID: config.NATS.ClusterID,
			ClientID:  config.NATS.ClientID,
			StanOptions: []stan.Option{
				stan.NatsURL(config.NATS.URL),
			},
			Marshaler: nats.GobMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

// NewNATSSubscriber returns a NATS subscriber for event streaming
func NewNATSSubscriber(config *conf.Config) (message.Subscriber, error) {
	subscriber, err := nats.NewStreamingSubscriber(
		nats.StreamingSubscriberConfig{
			ClusterID: config.NATS.ClusterID,
			ClientID:  config.NATS.ClientID,

			QueueGroup:       config.NATS.QueueGroup,
			DurableName:      config.NATS.DurableName,
			SubscribersCount: config.NATS.SubscriberCount,
			StanOptions: []stan.Option{
				stan.NatsURL(config.NATS.URL),
			},
			Unmarshaler: nats.GobMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return subscriber, nil
}
