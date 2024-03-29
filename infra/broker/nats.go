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

var Publisher message.Publisher

// NewNATSPublisher returns a NATS publisher for event streaming
func NewNATSPublisher(config *conf.Config) (message.Publisher, error) {
	var err error
	Publisher, err = nats.NewStreamingPublisher(
		nats.StreamingPublisherConfig{
			ClusterID: config.NATSConfig.ClusterID,
			ClientID:  config.NATSConfig.ClientID + "_publisher",
			StanOptions: []stan.Option{
				stan.NatsURL(config.NATSConfig.URL),
			},
			Marshaler: nats.GobMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	registry, ok := prom.DefaultRegisterer.(*prom.Registry)
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
