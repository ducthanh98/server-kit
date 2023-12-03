package handlers

import (
	"github.com/ducthanh98/server-kit/kit/consumer"
	rabbitmq_driver "github.com/ducthanh98/server-kit/kit/drivers/rabbitmq-driver"
	redis_driver "github.com/ducthanh98/server-kit/kit/drivers/redis-driver"
	"github.com/ducthanh98/server-kit/kit/logger"
)

// SampleHandler --
type SampleHandler struct {
	// defines exchange name to consume
	Key string
	// entity producer
	Producer *rabbitmq_driver.Producer
	Redis    *redis_driver.Redis
}

// Handle --
func (h *SampleHandler) Handle(msg []byte) consumer.ProcessStatus {
	logger.Log.Info("MSG: ", string(msg))

	return consumer.ProcessOK
}

// GetHandlers --
func (h *SampleHandler) GetHandlers() map[string]consumer.HandlerWOption {
	m := make(map[string]consumer.HandlerWOption)
	var replica int64 = 10
	hwo := consumer.BuildHandler(h, replica)
	m[h.Key] = *hwo
	return m
}

// SetProducer --
func (h *SampleHandler) SetProducer(p *rabbitmq_driver.Producer) {
	h.Producer = p
}

// Init --
func (h *SampleHandler) Init(w *consumer.Worker) {
	// do initialize works
	//h.Redis = w.App.CreateRedisConnection(nil)
	logger.Log.Info("dummy worker")
}

// Close --
func (h *SampleHandler) Close(w *consumer.Worker) {
	logger.Log.Info("Closing...")
}
