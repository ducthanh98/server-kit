package rabbitmq_driver

import (
	"errors"
	"fmt"
	"github.com/ducthanh98/server-kit/kit/config"
	"github.com/ducthanh98/server-kit/kit/consumer/entity"
	"github.com/ducthanh98/server-kit/kit/logger"
	"github.com/ducthanh98/server-kit/kit/utils/string_utils"
	"math"
	"net"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spf13/viper"

	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// BufferingData data for queue
type BufferingData struct {
	data       []byte
	exchName   string
	routingKey string
	mandatory  bool
	immediate  bool
}

type messageInfo struct {
	msg        []byte
	routingKey string
	xchName    string
	mandatory  bool
	immediate  bool
	retries    int64
}

type ProducerInterface interface {
	PublishSimple(exchName string, data []byte) (err error)
	PublishRouting(exchName, routingKey string, data []byte) (err error)
}

// Producer holds all infromation
// about the RabbitMQ connection
type Producer struct {
	conn          *amqp.Connection
	done          chan error
	config        *RabbitMqConfiguration
	status        bool
	retries       int
	messages      chan *messageInfo
	closed        int32
	inputCounter  int64
	outputCounter int64
	errorCounter  int64
	ticker        *time.Ticker

	maxThread int
	appId     string
}

var (
	// ErrSendToClosedProducer --
	ErrSendToClosedProducer = errors.New("send to closed producer...exiting")
)

// MaxThread --
var MaxThread = runtime.NumCPU()

func init() {
	AppId = config.GetPodName()
}

// NewRMQProducerFromConf creates Producer from configuration
func NewRMQProducerFromConf(conf *RabbitMqConfiguration,
	internalQueueSize int,
	retries int) *Producer {
	if len(conf.URI) == 0 {
		return nil
	}
	return &Producer{
		config:   conf,
		done:     make(chan error),
		status:   false,
		retries:  retries,
		messages: make(chan *messageInfo, internalQueueSize),
		ticker:   time.NewTicker(2 * time.Minute),
	}
}

// NewRMQProducerFromDefConf --
func NewRMQProducerFromDefConf() *Producer {
	uri := viper.GetString("rabbitmq.uri")
	retries := viper.GetInt64("rabbitmq.retries")
	internalQueueSize := viper.GetInt64("rabbitmq.internal_queue_size")
	if len(uri) == 0 {
		return nil
	}
	conf := &RabbitMqConfiguration{
		URI: uri,
	}

	return &Producer{
		config:   conf,
		done:     make(chan error),
		status:   false,
		retries:  int(retries),
		messages: make(chan *messageInfo, internalQueueSize),
		ticker:   time.NewTicker(2 * time.Minute),
	}

}

// NewRMQProducerFConfig returns new entity consumer with preset configuration
func NewRMQProducerFConfig(uri string, retries, internalQueueSize int64,
	co *entity.RmqOutputConf) *Producer {
	if len(uri) == 0 {
		return nil
	}
	conf := &RabbitMqConfiguration{
		ExchangeConfig: *convertToRmqConf(co.Exch),
		URI:            uri,
	}

	return &Producer{
		config:   conf,
		done:     make(chan error),
		status:   false,
		retries:  int(retries),
		messages: make(chan *messageInfo, internalQueueSize),
		ticker:   time.NewTicker(2 * time.Minute),
	}
}

// NewRMQProducer returns a Producer struct
// that has been initialized properly
// essentially don't touch conn, channel, or
// done and you can create Producer manually
func NewRMQProducer(
	uri,
	exchange,
	exchangeType string,
	internalQueueSize,
	retries int) *Producer {
	if len(uri) == 0 {
		return nil
	}
	defaultConfig := DefaultRmqConfiguration(uri, exchange, "")
	defaultConfig.ExchangeConfig.Type = exchangeType

	return &Producer{
		config:   defaultConfig,
		done:     make(chan error),
		status:   false,
		retries:  retries,
		messages: make(chan *messageInfo, internalQueueSize),
		ticker:   time.NewTicker(2 * time.Minute),
	}
}

// reconnect is called in places where NotifyClose() channel is called
// wait 30 seconds before trying to reconnect. Any shorter amount of time
// will  likely destroy the error log while waiting for servers to come
// back online. This requires two parameters which is just to satisfy
// the AccounceQueue call and allows greater flexability
func (p *Producer) reconnect() error {
	p.status = false
	time.Sleep(20 * time.Second)
	if err := p.connectQueue(); err != nil {
		logger.Log.Errorw("Could not connect in reconnect call", "errError()", err.Error())
		return err
	}

	return nil
}

func (p *Producer) Connect() error {
	p.maxThread = 1
	return p.connectQueue()
}

func (p *Producer) ConnectMulti(numThread int) error {
	p.maxThread = numThread
	return p.connectQueue()
}

// Connect to RabbitMQ server
func (p *Producer) connectQueue() error {
	var err error

	if p.config == nil {
		err = errors.New("missing entity configuration")

		logger.Log.Errorw(err.Error())
		return err
	}

	logger.Log.Infof("Connecting to %q", string_utils.CensorString(p.config.URI))
	p.conn, err = amqp.DialConfig(p.config.URI, amqp.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, DefaultNetworkTimeoutInSec*time.Second)
		},
	})

	if err != nil {
		return fmt.Errorf("Dial error: %s", err)
	}

	go func() {
		// Waits here for the channel to be closed
		closed := <-p.conn.NotifyClose(make(chan *amqp.Error))
		logger.Log.Debugf("Closing: %s", closed)

		if closed != nil {
			logger.Log.Info("Closing channel: ", closed.Reason, closed.Recover, closed.Code, closed.Server, closed.Error())
			// Let Handle know it's not time to reconnect
			// ensure goroutine go to end in every case
			select {
			case p.done <- errors.New("channel closed"):
			case <-time.After(10 * time.Second):
				return
			}
		}
	}()

	logger.Log.Info("Connected")

	p.status = true

	return nil
}

// DeclareExchWithDefaultConfig --
func (p *Producer) DeclareExchWithDefaultConfig(xchName, xchType string) error {
	xch := RmqConfiguration{Name: xchName, Type: xchType, Durable: true}
	return p.DeclareSpecificExch(&xch)
}

// DeclareSpecificExchN declares exchange - use after start connect
func (p *Producer) DeclareSpecificExchN(xch *entity.RmqExchQueueInfo) error {
	nxch := convertToRmqConf(xch)
	return p.DeclareSpecificExch(nxch)
}

// DeclareSpecificExch declares exchange - use after start connect
func (p *Producer) DeclareSpecificExch(xch *RmqConfiguration) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	logger.Log.Debugf("Got Channel, declaring Exchange (%q)", xch.Name)
	if err = ch.ExchangeDeclare(
		xch.Name,
		xch.Type,
		xch.Durable,
		xch.AutoDelete,
		xch.Internal,
		xch.NoWait, // noWait
		xch.Others, // arguments
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	return nil
}

// DeclareExch declares exchange - use after start connect
func (p *Producer) DeclareExch() error {
	return p.DeclareSpecificExch(&p.config.ExchangeConfig)
}

// publish publishes data to entity with full option
func (p *Producer) publish(c *amqp.Channel, mi *messageInfo) error {
	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        mi.msg,
		Timestamp:   time.Now(),
		AppId:       AppId,
	}
	err := c.Publish(mi.xchName, mi.routingKey, mi.mandatory, mi.immediate, msg)
	if err != nil {
		atomic.AddInt64(&(p.errorCounter), 1)
		err = fmt.Errorf("Cannot publish message to exchange, %v - %v - %v - %v - %v. %v",
			mi.xchName, mi.routingKey, mi.mandatory, mi.immediate, string(mi.msg), err)
		logger.Log.Error(err.Error())
		return err
	} else if mi.retries > 0 {
		atomic.AddInt64(&p.errorCounter, -mi.retries)
	}
	atomic.AddInt64(&(p.outputCounter), 1)

	return nil
}

// PublishMessage --
func (p *Producer) PublishMessage(m amqp.Delivery) (err error) {
	if !p.IsClosed() {
		err = p.publishWithTimeout(&messageInfo{msg: m.Body, xchName: m.Exchange, routingKey: m.RoutingKey})
	} else {
		err = ErrSendToClosedProducer
	}
	atomic.AddInt64(&(p.inputCounter), 1)
	return
}

// PublishSimple publishes message to entity with simplest options
func (p *Producer) PublishSimple(exchName string, data []byte) (err error) {
	if data == nil {
		return
	}
	if !p.IsClosed() {
		err = p.publishWithTimeout(&messageInfo{msg: data, xchName: exchName, routingKey: ""})
	} else {
		err = ErrSendToClosedProducer
	}
	atomic.AddInt64(&(p.inputCounter), 1)
	return
}

func (p *Producer) publishWithTimeout(mess *messageInfo) error {
	select {
	case p.messages <- mess:
	case <-time.After(time.Duration(RabbitProducerTimeout) * time.Millisecond):
		logger.Log.Warn("Publish message timeout :", mess.xchName, ",", RabbitProducerTimeout, ", ", string(mess.msg))
		return errors.New("publish message to rabbbit timeout")
	}

	return nil
}

// PublishRouting publishes message entity with a specific routing key
func (p *Producer) PublishRouting(exchName, routingKey string, data []byte) (err error) {
	if data == nil {
		return
	}
	if !p.IsClosed() {
		err = p.publishWithTimeout(&messageInfo{msg: data, xchName: exchName, routingKey: routingKey})
	} else {
		err = ErrSendToClosedProducer
	}
	atomic.AddInt64(&(p.inputCounter), 1)
	return
}

// PublishWithOption publishes message entity with a specific routing key
func (p *Producer) PublishWithOption(exchName, key string,
	mandatory, immediate bool, data []byte) (err error) {
	if data == nil {
		return
	}
	if !p.IsClosed() {
		err = p.publishWithTimeout(&messageInfo{msg: data, xchName: exchName, routingKey: key, mandatory: mandatory, immediate: immediate})
	} else {
		err = ErrSendToClosedProducer
	}
	atomic.AddInt64(&(p.inputCounter), 1)
	return
}

type logFunc func(args ...interface{})
type logFmtFunc func(string, ...interface{})

// Start has all the logic to make sure your program keeps running
// d should be a delievey channel as created when you call AnnounceQueue
// fn should be a function that handles the processing of deliveries
// this should be the last thing called in main as code under it will
// become unreachable unless put int a goroutine. The q and rk params
// are redundent but allow you to have multiple queue listeners in main
// without them you would be tied into only using one queue per connection
func (p *Producer) Start() {
	var logFmt logFmtFunc
	level := viper.GetString("producer.loglevel")
	isDebug := viper.GetBool("producer.debug")

	switch level {
	case "info":
		logFmt = logger.Log.Infof
	default:
		logFmt = logger.Log.Debugf
	}

	stopTicker := make(chan struct{})
	var countClose int64 = 0
	go func() {
		for {
			select {
			case <-p.ticker.C:
				logFmt(`{"type": "producer", "counter": {"input": %v, "output_success": %v, "error": %v}}`,
					p.InputCount(), p.OutputCount(), p.ErrorCount())
			case <-stopTicker:
				n := atomic.AddInt64(&countClose, 1)
				if n >= int64(p.maxThread) {
					return
				}
			}
		}
	}()

	configTimeout := viper.GetInt("rabbitmq.producer.timeout")
	if configTimeout > 0 {
		RabbitProducerTimeout = configTimeout
	}

	logFmt("Running producer with %v goroutines", p.maxThread)
	m := &sync.Mutex{}
	for i := 0; i < p.maxThread; i++ {
		go func(id int) {
			m.Lock()
			c, err := p.conn.Channel()
			for err != nil {
				time.Sleep(10 * time.Millisecond)
				c, err = p.conn.Channel()
			}
			m.Unlock()
			logFmt("Got channel for %v", id)
			for {
				select {
				case msg := <-p.messages:
					if msg == nil {
						logFmt("Got nil on %v. Breaking...", id)
						stopTicker <- struct{}{}
						defer c.Close()
						return
					}
					err := p.publish(c, msg)
					if err != nil {
						// maybe channel is dead, get new one
						logger.Log.Warnf("Maybe channel is dead, get new one. %v", id)
						c.Close()
						m.Lock()

						if isDebug {
							logger.Log.Warnf("Create chanel %v: %v", id)
						}

						var err = errors.New("start")
						for err != nil {
							time.Sleep(100 * time.Millisecond)
							if p.conn == nil {
								continue
							}

							if isDebug {
								logger.Log.Warnf("Create chanel is fail %v: %v", id, err)
							}

							c, err = p.conn.Channel()
						}
						m.Unlock()
						logger.Log.Infof("Got new channel! %v", id)
						msg.retries++
						go func() {
							p.publishWithTimeout(msg)
						}()
					}
				}
			}
		}(i)
	}

	go func() {
		var err error
		for {
			// Go into reconnect loop when
			// c.done is passed non nil values
			if err = <-p.done; err != nil {
				if strings.Contains(err.Error(), "channel closed") && !p.IsClosed() { // reconnect case
					logger.Log.Errorw("Disconnected", "err", err)
					p.status = false
					err = p.reconnect()
					retry := 0
					var base = 100
					step := 10
					exp := 2
					for err != nil {
						time.Sleep(time.Duration(base+int(math.Pow(float64(step), float64(exp)))) * time.Millisecond)
						// Very likely chance of failing
						// should not cause worker to terminate
						logger.Log.Errorw("Reconnecting Error. Try again", "err", err)
						retry++
						if retry > p.retries {
							panic(fmt.Errorf("Cannot retry connection after %v times", p.retries))
						}
						err = p.reconnect()
					}
					logger.Log.Infof("Reconnected")
				} else { // stop case
					p.conn.Close()
					logger.Log.Infof("Stopped")
					return
				}
			}
		}
	}()
}

// Close close the producer
func (p *Producer) Close() {
	atomic.StoreInt32(&(p.closed), 1)
	time.Sleep(1 * time.Second)
	close(p.messages)
	p.done <- errors.New("stop entity producer")
	p.conn.Close()
	p.ticker.Stop()
}

func (p *Producer) IsClosed() bool {
	return atomic.LoadInt32(&(p.closed)) == 1
}

// Expose metrics
func (p *Producer) OutputCount() int64 {
	return atomic.LoadInt64(&(p.outputCounter))
}
func (p *Producer) InputCount() int64 {
	return atomic.LoadInt64(&(p.inputCounter))
}

func (p *Producer) ErrorCount() int64 {
	return atomic.LoadInt64(&(p.errorCounter))
}
