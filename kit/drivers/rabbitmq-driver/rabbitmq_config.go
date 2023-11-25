package rabbitmq_driver

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
