package consumer

import (
	"context"
	"fmt"
	"github.com/ducthanh98/server-kit/kit/consumer/entity"
	"github.com/ducthanh98/server-kit/kit/consumer/input"
	"github.com/ducthanh98/server-kit/kit/consumer/output"
	"github.com/ducthanh98/server-kit/kit/utils/string_utils"
	log "github.com/sirupsen/logrus"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// ConsumerDef --
type ConsumerDef struct {
	Consumer interface{}
	Mode     string
}

// ProducerDef --
type ProducerDef struct {
	Producer interface{}
	Mode     string
}

// StreamMessages for processing from multiple streams on just one handler
type StreamMessages struct {
	StreamName string
	Message    []byte
}

// MsgHandler --
type MsgHandler interface {
	Handle(msg []byte) ProcessStatus
	//HandleMultipleStream(msgs StreamMessages) ProcessStatus
}

// HandlerWOption --
type HandlerWOption struct {
	MsgHandler
	Replica int64

	// allow disable metrics
	ignoreMetrics bool
	IdleSleeptime int
}

// Task --
type Task struct {
	TaskDefinition  *GConsumerDef
	consumerHandler map[string]HandlerWOption
	consumers       map[string]ConsumerDef
	producer        *output.RabbitMQOutput // for convenience, we just need one producer for all outputs
	ctx             context.Context
	cancelFunc      context.CancelFunc
}

func (t *Task) prepare() {
	t.buildIO()
}

// InitializeHandler initialize message handlers for the task
func (t *Task) InitializeHandler(handlers map[string]HandlerWOption) {
	for k, v := range handlers {
		replicaNumber := v.Replica
		if replicaNumber < 0 {
			v.Replica = int64(runtime.NumCPU())
		}
		v.IdleSleeptime = DefaultSleepTimeWhenIdle
		if viper.IsSet("task.idle_sleep_time") {
			v.IdleSleeptime = viper.GetInt("task.idle_sleep_time")
		}
		log.Infof("Setup idle time in %v (ms)", v.IdleSleeptime)
		t.consumerHandler[k] = v
	}
}

func (t *Task) buildInput() {
	for _, i := range t.TaskDefinition.Inputs {
		log.Debug("input: ", string_utils.ToJSONString(t.TaskDefinition.Inputs))
		switch i.Mode {
		case input.RabbitmqMode:
			c, err := input.BuildRabbitmqInput(t.ctx,
				t.TaskDefinition.Group.Name, i.Config.(*entity.RmqInputConf))
			if err != nil {
				log.Fatal("Cannot connect to entity", i)
				continue
			}

			key := fmt.Sprintf("%v-%v", c.XchName, c.QueueName)
			t.consumers[key] = ConsumerDef{Consumer: c, Mode: i.Mode}
		}
	}
}

func (t *Task) buildOutput() {
	for _, i := range t.TaskDefinition.Outputs {
		switch i.Mode {
		case output.RabbitmqMode:
			co := i.Config.(*entity.RmqOutputConf)
			if t.producer == nil {
				p, err := output.BuildRabbitmqOutput(co)
				if err != nil {
					log.Fatal("Cannot initialize entity producer", i, err)
				}
				if err := p.DeclareSpecificExchN(co.Exch); err != nil {
					log.Fatal("Cannot declare exchange due to error", i, err)
				}
				t.producer = p
			} else {
				_ = t.producer.DeclareSpecificExchN(co.Exch)
			}
		}
	}
}

func (t *Task) buildIO() {
	t.buildInput()
	t.buildOutput()
}

func (t *Task) findHandlerForConsumer(consumerName string) (HandlerWOption, error) {
	// find exactly
	log.Debugf("Consumer name: %v and t_consumer", consumerName, t.consumerHandler)
	if v, ok := t.consumerHandler[consumerName]; ok {
		return v, nil
	}
	// find by exchange only
	for k, v := range t.consumerHandler {
		if consumerName == k {
			return v, nil
		} else if strings.Contains(consumerName, k) {
			return v, nil
		}
	}
	if _, ok := t.consumerHandler[DefaultConsumerName]; ok { // get default consumer
		return t.consumerHandler[DefaultConsumerName], nil
	}
	return HandlerWOption{}, fmt.Errorf("cannot find handler for consumer %v", consumerName)
}

func (t *Task) start() {
	wait := &sync.WaitGroup{}
	log.Infoln("t_consumer: ", t.consumers)
	for k, v := range t.consumers {
		if h, err := t.findHandlerForConsumer(k); err == nil {
			log.Infof("Bind %v to %v", h.MsgHandler, k)
			wait.Add(1)
			go run(t.ctx, t, k, v, h, wait)
		} else {
			log.Fatal("Cannot find handler for consumer ", k)
		}
	}
	wait.Wait()
	log.Info("Exiting task...")
}

func (t *Task) stop() {
	t.cancelFunc()

	if t.producer != nil {
		defer t.producer.Close()
	}
	log.Info("Wait for other task to complete...")
	time.Sleep(100 * time.Millisecond)
}
