package broker

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill-nats/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	conf "github.com/minghsu0107/saga-purchase/config"
	stan "github.com/nats-io/stan.go"
	prom "github.com/prometheus/client_golang/prometheus"
)

// Publisher singleton
var Publisher message.Publisher

// Subscriber singleton
var Subscriber message.Subscriber

// NewNATSPublisher returns a NATS publisher for event streaming
func NewNATSPublisher(config *conf.Config) (message.Publisher, error) {
	var err error
	Publisher, err = nats.NewStreamingPublisher(
		nats.StreamingPublisherConfig{
			ClusterID: config.NATS.ClusterID,
			ClientID:  config.NATS.ClientID + "_publisher",
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

	registry, ok := prom.DefaultGatherer.(*prom.Registry)
	if !ok {
		return nil, fmt.Errorf("prometheus type casting error")
	}
	metricsBuilder := metrics.NewPrometheusMetricsBuilder(registry, config.App, "pubsub")
	Publisher, err = metricsBuilder.DecoratePublisher(Publisher)
	if err != nil {
		return nil, err
	}
	return Publisher, nil
}

// NewNATSSubscriber returns a NATS subscriber for event streaming
func NewNATSSubscriber(config *conf.Config) (message.Subscriber, error) {
	var err error
	Subscriber, err = nats.NewStreamingSubscriber(
		nats.StreamingSubscriberConfig{
			ClusterID: config.NATS.ClusterID,
			ClientID:  config.NATS.ClientID + "_subscriber",

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

	registry, ok := prom.DefaultGatherer.(*prom.Registry)
	if !ok {
		return nil, fmt.Errorf("prometheus type casting error")
	}
	metricsBuilder := metrics.NewPrometheusMetricsBuilder(registry, config.App, "pubsub")
	Subscriber, err = metricsBuilder.DecorateSubscriber(Subscriber)
	if err != nil {
		return nil, err
	}
	return Subscriber, nil
}
