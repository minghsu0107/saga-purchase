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
	NATS           *NATS           `yaml:"nats"`
	RPCEndpoints   *RPCEndpoints   `yaml:"rpcEndpoints"`
	ServiceOptions *ServiceOptions `yaml:"serviceOptions"`
	Logger         *Logger
}

// NATS wraps NATS client configurations
type NATS struct {
	ClusterID       string `yaml:"clusterID" envconfig:"NATS_CLUSTER_ID"`
	URL             string `yaml:"url" envconfig:"NATS_URL"`
	ClientID        string `yaml:"clientID" envconfig:"NATS_CLIENT_ID"`
	QueueGroup      string `yaml:"queueGroup" envconfig:"NATS_QUEUE_GROUP"`
	DurableName     string `yaml:"durableName" envconfig:"NATS_DURABLE_NAME"`
	SubscriberCount int    `yaml:"subscriberCount" envconfig:"NATS_SUBSCRIBER_COUNT"`
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
	if config.NATS.ClientID == "" {
		config.NATS.ClientID = watermill.NewShortUUID()
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
