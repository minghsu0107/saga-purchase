package config

import (
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	log "github.com/sirupsen/logrus"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

// Config is a type for general configuration
type Config struct {
	App            string          `yaml:"app" envconfig:"APP"`
	GinMode        string          `yaml:"ginMode" envconfig:"GIN_MODE"`
	HTTPPort       string          `yaml:"httpPort" envconfig:"HTTP_PORT"`
	PromPort       string          `yaml:"promPort" envconfig:"PROM_PORT"`
	OcAgentHost    string          `yaml:"ocAgentHost" envconfig:"OC_AGENT_HOST"`
	Resolver       string          `yaml:"resolver" envconfig:"RESOLVER"`
	NATSConfig     *NATSConfig     `yaml:"natsConfig"`
	RedisConfig    *RedisConfig    `yaml:"redisConfig"`
	RPCEndpoints   *RPCEndpoints   `yaml:"rpcEndpoints"`
	ServiceOptions *ServiceOptions `yaml:"serviceOptions"`
	Logger         *Logger
}

// NATSConfig wraps NATS client configurations
type NATSConfig struct {
	ClusterID  string          `yaml:"clusterID" envconfig:"NATS_CLUSTER_ID"`
	ClientID   string          `yaml:"clientID" envconfig:"NATS_CLIENT_ID"`
	URL        string          `yaml:"url" envconfig:"NATS_URL"`
	Subscriber *NATSSubscriber `yaml:"subscriber"`
}

type NATSSubscriber struct {
	QueueGroup  string `yaml:"queueGroup" envconfig:"NATS_SUBSCRIBER_QUEUE_GROUP"`
	DurableName string `yaml:"durableName" envconfig:"NATS_SUBSCRIBER_DURABLE_NAME"`
	Count       int    `yaml:"count" envconfig:"NATS_SUBSCRIBER_COUNT"`
}

// RedisConfig is redis config type
type RedisConfig struct {
	Addrs              string           `yaml:"addrs" envconfig:"REDIS_ADDRS"`
	Password           string           `yaml:"password" envconfig:"REDIS_PASSWORD"`
	DB                 int              `yaml:"db" envconfig:"REDIS_DB"`
	PoolSize           int              `yaml:"poolSize" envconfig:"REDIS_POOL_SIZE"`
	MaxRetries         int              `yaml:"maxRetries" envconfig:"REDIS_MAX_RETRIES"`
	ExpirationSeconds  int64            `yaml:"expirationSeconds" envconfig:"REDIS_EXPIRATION_SECONDS"`
	IdleTimeoutSeconds int64            `yaml:"idleTimeoutSeconds" envconfig:"REDIS_IDLE_TIMEOUT_SECONDS"`
	Subscriber         *RedisSubscriber `yaml:"subscriber"`
}

type RedisSubscriber struct {
	ConsumerID    string `yaml:"consumerID" envconfig:"REDIS_SUBSCRIBER_CONSUMER_ID"`
	ConsumerGroup string `yaml:"consumerGroup" envconfig:"REDIS_SUBSCRIBER_CONSUMER_GROUP"`
}

// RPCEndpoints wraps all rpc server urls
type RPCEndpoints struct {
	AuthSvcHost    string `yaml:"authSvcHost" envconfig:"RPC_AUTH_SVC_HOST"`
	ProductSvcHost string `yaml:"productSvcHost" envconfig:"RPC_PRODUCT_SVC_HOST"`
}

// ServiceOptions defines options for rpc calls between services
type ServiceOptions struct {
	Rps           int `yaml:"rps" envconfig:"SERVICE_OPTIONS_RPS"`
	TimeoutSecond int `yaml:"timeoutSecond" envconfig:"SERVICE_OPTIONS_TIMEOUT_SECOND"`
	Timeout       time.Duration
}

// NewConfig is a factory for Config instance
func NewConfig() (*Config, error) {
	var config Config
	if err := readFile(&config); err != nil {
		return nil, err
	}
	if err := readEnv(&config); err != nil {
		return nil, err
	}
	config.Logger = newLogger(config.App)
	log.SetOutput(config.Logger.Writer)
	if config.NATSConfig.ClientID == "" {
		config.NATSConfig.ClientID = watermill.NewShortUUID()
	}
	if config.RedisConfig.Subscriber.ConsumerID == "" {
		config.RedisConfig.Subscriber.ConsumerID = watermill.NewShortUUID()
	}
	config.ServiceOptions.Timeout = time.Duration(config.ServiceOptions.TimeoutSecond) * time.Second
	return &config, nil
}

func readFile(config *Config) error {
	f, err := os.Open("config.yml")
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)
	if err != nil {
		return err
	}
	return nil
}

func readEnv(config *Config) error {
	err := envconfig.Process("", config)
	if err != nil {
		return err
	}
	return nil
}
