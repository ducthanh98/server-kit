package consumer

import (
	"context"
	"github.com/ducthanh98/server-kit/kit/consumer/input"
	"github.com/ducthanh98/server-kit/kit/logger"
	"sync"
	"time"

	"sync/atomic"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ProcessStatus --
type ProcessStatus struct {
	Code    int64
	Message []byte
}

var ( // Status
	// ProcessOK marks message as done
	ProcessOK = ProcessStatus{Code: ProcessOKCode}
	// ProcessFailRetry marks message as fail and retry
	ProcessFailRetry = ProcessStatus{Code: ProcessFailRetryCode}
	// ProcessFailDrop marks message as fail and drop
	ProcessFailDrop = ProcessStatus{Code: ProcessFailDropCode}
	// ProcessFailReproduce --
	ProcessFailReproduce = ProcessStatus{Code: ProcessFailReproduceCode}
)

// NewProcessStatus --
func NewProcessStatus(code int64, message []byte) ProcessStatus {
	return ProcessStatus{Code: code, Message: message}
}

const (
	// DefaultConsumerName --
	DefaultConsumerName = "**_QJVtqiSTGcrFkatIyzja_**"
)

func run(ctx context.Context, task *Task, name string, consumerDef ConsumerDef,
	handler HandlerWOption, wait *sync.WaitGroup) {
	defer wait.Done()
	var success, retry, drop int64

	if !handler.ignoreMetrics {
		tick := time.NewTicker(2 * time.Minute)
		defer tick.Stop()

		go func() {
			for {
				select {
				case <-tick.C:
					logger.Log.Debugf(`{"type": "consumer", "counter": {"success": %v, "fail_retried": %v, "drop": %v}}`,
						success, retry, drop)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	switch consumer := consumerDef.Consumer.(type) {
	case *input.RabbitMQInput:
		consumer.Handle(consumer.Delivery, func(deliveries <-chan amqp.Delivery) {
			for {
				select {
				case d := <-deliveries:
					if len(d.Body) > 0 {
						resp := handler.Handle(d.Body)
						switch resp.Code {
						case ProcessOKCode:
							atomic.AddInt64(&success, 1)
							_ = d.Ack(false)
						case ProcessFailDropCode:
							atomic.AddInt64(&drop, 1)
							_ = d.Ack(false)
						case ProcessFailRetryCode:
							atomic.AddInt64(&retry, 1)
							_ = d.Nack(false, true)
						case ProcessFailReproduceCode:
							var republishMsg = d
							if resp.Message != nil && len(resp.Message) > 0 {
								republishMsg = amqp.Delivery{Body: resp.Message, Exchange: d.Exchange, RoutingKey: d.RoutingKey}
							}
							err := task.producer.PublishMessage(republishMsg)
							if err != nil {
								logger.Log.Warnf("Cannot republish message %v. %v", string(d.Body), err)
								_ = d.Nack(false, true)
							} else {
								_ = d.Ack(false)
							}

						default: // case of using defer in handler
							atomic.AddInt64(&drop, 1)
							_ = d.Ack(false)
						}
					}
				case <-ctx.Done():
					// exit consumer
					consumer.Close()
					logger.Log.Debugf("Exiting %v...", name)
					return

				default:
					time.Sleep(time.Duration(handler.IdleSleeptime) * time.Millisecond)
				}
			}
		}, int(handler.Replica))

	}
	logger.Log.Infof("Exited consumer %v...", name)
}
