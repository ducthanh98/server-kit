package consumer

import (
	"context"
	"github.com/ducthanh98/server-kit/kit/app"
	"github.com/ducthanh98/server-kit/kit/config"
	queue "github.com/ducthanh98/server-kit/kit/drivers/rabbitmq-driver"
	"github.com/ducthanh98/server-kit/kit/utils/signal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

const (
	DefaultTaskPreparationTime = 20
)

// CronTask --
type CronTask interface {
	// GetHandlers returns handler(s) with replica option for each one
	GetHandlers() map[string]interface{}
	// SetProducer same with SetDmsClient but for producer (currently just for entity)
	SetProducer(*queue.Producer)
	// Init initialize handler. it will be called by worker internally
	Init(*Worker)
	// Close also be called by worker when closing
	Close(*Worker)
}

// TaskInterface will be deprecated soon. Please use CronTask instead
type TaskInterface interface {
	// GetHandlers returns handler(s) with replica option for each one
	GetHandlers() map[string]HandlerWOption
	// SetProducer same with SetDmsClient but for producer (currently just for entity)
	SetProducer(*queue.Producer)
	// Init initialize handler. it will be called by worker internally
	Init(*Worker)
	// Close also be called by worker when closing
	Close(*Worker)
}

// Bootstrapper --
var Bootstrapper = &Bootstrap{F: make([]TaskInterface, 0)}

// Bootstrap --
type Bootstrap struct {
	F []TaskInterface
	S []CronTask
}

// AddConsumerTask --
func (b *Bootstrap) AddConsumerTask(t TaskInterface) {
	b.F = append(b.F, t)
}

// AddScheduleTask --
func (b *Bootstrap) AddScheduleTask(t CronTask) {
	b.S = append(b.S, t)
}

// Worker --
type Worker struct {
	app.App
	// Common
	// task definition files
	TaskDefinitionURI string
	task              *Task
	//scheduler         *Scheduler
	stopped bool
}

// Callback --
type Callback func(*Worker)

// NewWorker --
func NewWorker(onClose Callback) *Worker {
	w := &Worker{App: app.App{AppName: "worker"}}
	w.initCommon()

	// handle sigterm
	signal.HandleSigterm(func() {
		log.Infof("Stopping...")
		if onClose != nil {
			onClose(w)
		}
		w.Stop()
	})
	return w
}

// CtxKey --
type CtxKey string

func (w *Worker) initCommon() {
	config.LoadConfig()
}

func (w *Worker) loadTaskDef(defFile string) *Task {
	var taskDefinition string
	var err error
	log.Debug("Loading task....")
	if defFile != "" {
		taskDefinition, err = config.ReadRawFile(defFile)
	} else {
		w.TaskDefinitionURI = viper.GetString("task.def")
		taskDefinition, err = config.ReadRawFile(w.TaskDefinitionURI)
	}

	if err != nil || taskDefinition == "" {
		log.Fatal("Cannot read task definition", w.TaskDefinitionURI, err)
	}

	ctx := context.Background()
	taskCtx := context.WithValue(ctx, CtxKey("taskctx"), "task context")
	newCtx, cancelFunc := context.WithCancel(taskCtx)
	taskDefStruct := parseTaskDef(taskDefinition)

	task := Task{
		TaskDefinition:  taskDefStruct,
		consumers:       make(map[string]ConsumerDef),
		consumerHandler: make(map[string]HandlerWOption),
		cancelFunc:      cancelFunc, ctx: newCtx,
	}
	w.task = &task
	log.Debug("Loading task....done")
	return &task
}

// Walk --
func (w *Worker) Walk() {
	log.Debug("Start consumer")
	w.task.start()
	w.Stop()
}

// PrepareTask --
func (w *Worker) PrepareTask(defFile string) *Task {
	w.loadTaskDef(defFile)

	w.task.prepare()

	return w.task
}

// BuildHandler --
func BuildHandler(h MsgHandler, replica int64) *HandlerWOption {
	ignoreMetrics := false
	//if disableMetrics, ok := h.(DisableMetricTask); ok {
	//	ignoreMetrics = disableMetrics.IsDisableMetric()
	//}

	return &HandlerWOption{MsgHandler: h, Replica: replica, ignoreMetrics: ignoreMetrics}
}

// PrepareHandlers --
func (w *Worker) PrepareHandlers(m map[string]HandlerWOption) {
	w.task.InitializeHandler(m)
}

// PrepareHandler --
func (w *Worker) PrepareHandler(h HandlerWOption) {
	m := make(map[string]HandlerWOption)
	m[DefaultConsumerName] = h
	w.task.InitializeHandler(m)
}

// GetProducer --
func (w *Worker) GetProducer() *queue.Producer {
	if w.task.producer != nil {
		return w.task.producer.Producer
	}
	return nil
}

// GetProducerFromConf --
func (w *Worker) GetProducerFromConf() *queue.Producer {
	p := queue.NewRMQProducerFromDefConf()
	if p == nil {
		log.Info("nil producer")
	} else {
		if err := p.Connect(); err != nil {
			log.Fatal("Cannot connect to rabbitmq. Please check configuration file for more information", err)
		}
		p.Start()
		signal.HandleSigterm(func() {
			if p != nil {
				log.Debug("Stop producer")
				p.Close()
			}
		})
	}
	return p
}

// Stop stops worker
func (w *Worker) Stop() {
	if w.stopped {
		log.Info("Call Stop method multiple times")
		return
	}
	w.stopped = true
	if w.task != nil {
		log.Info("Stop consumer")
		defer w.task.stop()
	}
	//if w.scheduler != nil {
	//	log.Info("Stop scheduler")
	//	w.scheduler.stop()
	//}
	time.Sleep(2 * time.Second)
}
