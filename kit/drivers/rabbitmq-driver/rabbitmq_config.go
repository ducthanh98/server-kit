package rabbitmq_driver

import "github.com/spf13/viper"

const (
	DefaultNetworkTimeoutInSec = 25
)

var (
	RabbitProducerTimeout = 10000
	AppId                 = ""
)

// RabbitMqConfiguration configuration for entity
type RabbitMqConfiguration struct {
	// amqp://user:pass@host:port/vhost?heartbeat=10&connection_timeout=10000&channel_max=100
	URI string
	// Exchange configuration
	ExchangeConfig RmqConfiguration
	QueueConfig    RmqConfiguration
}

// RmqConfiguration is configuration for exchange
type RmqConfiguration struct {
	Name       string
	Type       string
	AutoDelete bool
	Durable    bool
	Internal   bool
	Exclusive  bool
	NoWait     bool
	Others     map[string]interface{}
}

// DefaultRmqConfiguration create default conf for entity client
func DefaultRmqConfiguration(uri, exchangeName, queueName string) *RabbitMqConfiguration {
	return &RabbitMqConfiguration{
		uri,
		RmqConfiguration{exchangeName, "direct", false, true, false, false, false, make(map[string]interface{})},
		RmqConfiguration{queueName, "direct", false, true, false, false, false, make(map[string]interface{})},
	}
}

// LoadExchConfig Load exchange configuration from file
func LoadExchConfig() *RmqConfiguration {
	exch := viper.GetString("rabbitmq.exchange")
	exchType := viper.GetString("rabbitmq.exchange_type")
	xchAutoDelete := viper.GetBool("rabbitmq.exchange_autodelete")
	xchDurable := viper.GetBool("rabbitmq.exchange_durable")

	return &RmqConfiguration{exch, exchType, xchAutoDelete, xchDurable, false, false, false, make(map[string]interface{})}
}

// LoadOtherConfigParams -- internal queue size, number of retries
func LoadOtherConfigParams() map[string]interface{} {
	internalQueueSize := viper.GetInt("rabbitmq.internal_queue_size")
	retries := viper.GetInt("rabbitmq.retries")
	uri := viper.GetString("rabbitmq.uri")
	numThread := viper.GetInt("rabbitmq.num_thread")
	m := make(map[string]interface{})
	m["internal_queue_size"] = internalQueueSize
	m["retries"] = retries
	m["uri"] = uri
	m["num_thread"] = numThread
	return m
}
