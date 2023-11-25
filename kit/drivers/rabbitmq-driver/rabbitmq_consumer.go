package rabbitmq_driver

import (
	"errors"
	"fmt"
	"github.com/ducthanh98/server-kit/kit/consumer/entity"
	"github.com/ducthanh98/server-kit/kit/utils/string_utils"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"

	"time"
)

const (
	// DefaultRetries --
	DefaultRetries = 10
)

// Consumer holds all infromation
// about the RabbitMQ connection
// This setup does limit a consumer
// to one exchange. This should not be
// an issue. Having to connect to multiple
// exchanges means something else is
// structured improperly.
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan error
	config  *RabbitMqConfiguration
	retries int
	// consumerTag  string // Name that consumer identifies itself to the server with
	// uri          string // uri of the entity server
	// exchange     string // exchange that we will bind to
	// exchangeType string // topic, direct, etc...
	// bindingKey   string // routing key that we are using
}

// NewRMQConsumer returns a Consumer struct
// that has been initialized properly
// essentially don't touch conn, channel, or
// done and you can create Consumer manually
func NewRMQConsumer(
	uri,
	exchange,
	exchangeType,
	queueName,
	queueType,
	bindingKey string) *Consumer {
	defaultConfig := DefaultRmqConfiguration(uri, exchange, queueName)
	defaultConfig.ExchangeConfig.Type = exchangeType
	defaultConfig.QueueConfig.Type = queueType
	defaultConfig.QueueConfig.Others["binding_key"] = bindingKey

	return &Consumer{
		config:  defaultConfig,
		done:    make(chan error),
		retries: DefaultRetries,
	}
}

// returns exchange and queue information
func convertToRmqConf(i *entity.RmqExchQueueInfo) *RmqConfiguration {
	m := make(map[string]interface{})
	if len(i.RoutingKey) > 0 {
		m["binding_key"] = i.RoutingKey
	}

	return &RmqConfiguration{
		Name:       i.Name,
		Type:       i.Type,
		Durable:    i.Durable,
		AutoDelete: i.AutoDelete,
		Exclusive:  i.Exclusive,
		Internal:   i.Internal,
		NoWait:     i.Nowait,
		Others:     m,
	}
}

// NewRMQConsumerFConfig returns new entity consumer with preset configuration
func NewRMQConsumerFConfig(uri string, ci *entity.RmqInputConf) *Consumer {
	conf := &RabbitMqConfiguration{
		ExchangeConfig: *convertToRmqConf(ci.Exch),
		QueueConfig:    *convertToRmqConf(ci.Queue),
		URI:            uri,
	}
	return &Consumer{
		config:  conf,
		done:    make(chan error),
		retries: DefaultRetries,
	}
}

// NewRMQConsumerWConfig returns new entity consumer with preset configuration
func NewRMQConsumerWConfig(conf *RabbitMqConfiguration) *Consumer {
	return &Consumer{
		config:  conf,
		done:    make(chan error),
		retries: DefaultRetries,
	}
}

// reconnect is called in places where NotifyClose() channel is called
// wait 30 seconds before trying to reconnect. Any shorter amount of time
// will  likely destroy the error log while waiting for servers to come
// back online. This requires two parameters which is just to satisfy
// the AccounceQueue call and allows greater flexability
func (c *Consumer) reconnect() (<-chan amqp.Delivery, error) {
	time.Sleep(30 * time.Second)

	if err := c.Connect(); err != nil {
		log.Errorln("Could not connect in reconnect call", "errError()", err.Error())
		return nil, err
	}

	deliveries, err := c.AnnounceQueue()
	if err != nil {
		return deliveries, errors.New("Couldn't connect")
	}

	return deliveries, nil
}

// Connect to RabbitMQ server
func (c *Consumer) Connect() error {

	var err error

	log.Infof("Connecting to %q", string_utils.CensorString(c.config.URI))

	c.conn, err = amqp.DialConfig(c.config.URI, amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, DefaultNetworkTimeoutInSec*time.Second)
		},
	})
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}

	go func() {
		// Waits here for the channel to be closed
		log.Debugf("Closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
		// Let Handle know it's not time to reconnect
		c.done <- errors.New("Channel Closed")
	}()

	log.Debugf("Got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	log.Debugf("Got Channel, declaring Exchange (%q)", c.config.ExchangeConfig.Name)
	if err = c.channel.ExchangeDeclare(
		c.config.ExchangeConfig.Name,
		c.config.ExchangeConfig.Type,
		c.config.ExchangeConfig.Durable,
		c.config.ExchangeConfig.AutoDelete,
		c.config.ExchangeConfig.Internal,
		c.config.ExchangeConfig.NoWait, // noWait
		c.config.ExchangeConfig.Others, // arguments
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}
	log.Infof(" Connected to %q", string_utils.CensorString(c.config.URI))

	return nil
}

// AnnounceQueue sets the queue that will be listened to for this connection
func (c *Consumer) AnnounceQueue() (<-chan amqp.Delivery, error) {
	log.Debugf("declared Exchange, declaring Queue %q", c.config.QueueConfig.Name)
	queue, err := c.channel.QueueDeclare(
		c.config.QueueConfig.Name,
		c.config.QueueConfig.Durable,
		c.config.QueueConfig.AutoDelete,
		c.config.QueueConfig.Exclusive,
		c.config.QueueConfig.NoWait,
		c.config.QueueConfig.Others,
	)

	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	// get other configuration
	var qos int
	var bindingKey string
	var autoAck bool

	value, ok := c.config.QueueConfig.Others["qos"]
	if ok {
		qos = value.(int)
	} else {
		qos = 50
	}

	value, ok = c.config.QueueConfig.Others["binding_key"]
	if ok {
		bindingKey = value.(string)
	} else {
		bindingKey = ""
	}

	value, ok = c.config.QueueConfig.Others["auto_ack"]
	if ok {
		autoAck = value.(bool)
	} else {
		autoAck = false
	}

	log.Debugf("Declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, bindingKey)

	// Qos determines the amount of messages that the queue will pass to you before
	// it waits for you to ack them. This will slow down queue consumption but
	// give you more certainty that all messages are being processed. As load increases
	// I would reccomend upping the about of Threads and Processors the go process
	// uses before changing this although you will eventually need to reach some
	// balance between threads, procs, and Qos.
	err = c.channel.Qos(qos, 0, false)
	if err != nil {
		return nil, fmt.Errorf("Error setting qos: %s", err)
	}

	if err = c.channel.QueueBind(
		c.config.QueueConfig.Name,
		bindingKey,
		c.config.ExchangeConfig.Name,
		c.config.QueueConfig.NoWait,
		c.config.QueueConfig.Others,
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	log.Debugf("Queue bound to Exchange, starting Consume")
	deliveries, err := c.channel.Consume(
		c.config.QueueConfig.Name,
		"",
		autoAck,
		c.config.QueueConfig.Exclusive,
		false,
		c.config.QueueConfig.NoWait,
		c.config.QueueConfig.Others,
	)

	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	return deliveries, nil
}

// Handle has all the logic to make sure your program keeps running
// d should be a delievey channel as created when you call AnnounceQueue
// fn should be a function that handles the processing of deliveries
// this should be the last thing called in main as code under it will
// become unreachable unless put int a goroutine. The q and rk params
// are redundent but allow you to have multiple queue listeners in main
// without them you would be tied into only using one queue per connection
func (c *Consumer) Handle(
	d <-chan amqp.Delivery,
	fn func(<-chan amqp.Delivery),
	threads int) {

	var err error

	for {
		for i := 0; i < threads; i++ {
			go fn(d)
		}

		// Go into reconnect loop when
		// c.done is passed non nil values
		if e := <-c.done; e != nil {
			if strings.Contains(e.Error(), "Channel Closed") { // retry
				d, err = c.reconnect()
				retries := 0
				for err != nil {

					// Very likely chance of failing
					// should not cause worker to terminate
					log.Errorln("Reconnecting Error", "err", err, "retries", retries)
					retries++
					if retries > c.retries {
						panic(errors.New("Cannot reconnect to entity"))
					}
					d, err = c.reconnect()
				}
				log.Infof("Reconnected")
			} else { // stop
				return
			}
		}
	}
}

// Close close the consumer
func (c *Consumer) Close() {
	c.done <- errors.New("Stop Consumer")
	c.channel.Close()
	c.conn.Close()
}
