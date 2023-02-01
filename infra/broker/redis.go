package broker

import (
	"context"
	"fmt"
	"strings"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	conf "github.com/minghsu0107/saga-purchase/config"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

var (
	Subscriber  message.Subscriber
	RedisClient redis.UniversalClient
)

// NewRedisSubscriber returns a redis subscriber for event streaming
func NewRedisSubscriber(config *conf.Config) (message.Subscriber, error) {
	ctx := context.Background()
	RedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:         getServerAddrs(config.RedisConfig.Addrs),
		Password:      config.RedisConfig.Password,
		PoolSize:      config.RedisConfig.PoolSize,
		MaxRetries:    config.RedisConfig.MaxRetries,
		ReadOnly:      true,
		RouteRandomly: true,
	})
	pong, err := RedisClient.Ping(ctx).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	}
	redisotel.InstrumentTracing(RedisClient)
	config.Logger.ContextLogger.WithField("type", "setup:redis").Info("successful redis connection: " + pong)

	Subscriber, err = redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:       RedisClient,
			Unmarshaller: &redisstream.DefaultMarshallerUnmarshaller{},
			Consumer:     config.RedisConfig.Subscriber.ConsumerID,
			// use fan-out mode if leaved empty
			ConsumerGroup: config.RedisConfig.Subscriber.ConsumerGroup,

			// messages idling longer than this period will be claimed by the current subscriber
			// newly joined subscriber will try to claim pending messages immediately and then claim every 5 seconds
			// configure this only when ConsumerGroup is not empty
			// MaxIdleTime: time.Second * 60,
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

func getServerAddrs(addrs string) []string {
	return strings.Split(addrs, ",")
}
