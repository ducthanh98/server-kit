package input

import (
	"context"
	"fmt"
	"github.com/ducthanh98/server-kit/kit/consumer/entity"
	queue "github.com/ducthanh98/server-kit/kit/drivers/rabbitmq-driver"
	"github.com/ducthanh98/server-kit/kit/utils/string_utils"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// RabbitmqMode mode
	RabbitmqMode = "rabbitmq"
)

type RmqExchQueueInfo struct {
	Name       string `json:"name,omitempty"`
	Type       string `json:"type,omitempty"`
	AutoDelete bool   `json:"auto_delete,omitempty"`
	Durable    bool   `json:"durable,omitempty"`
	Internal   bool   `json:"internal,omitempty"`
	Exclusive  bool   `json:"exclusive,omitempty"`
	Nowait     bool   `json:"nowait,omitempty"`
	RoutingKey string `json:"routing_key,omitempty"`
	BatchAck   bool   `json:"batch_ack,omitempty"`
}

type RmqInputConf struct {
	Exch     *RmqExchQueueInfo `json:"exch,omitempty"`
	Queue    *RmqExchQueueInfo `json:"queue,omitempty"`
	BatchAck bool              `json:"batch_ack,omitempty"`
	Mode     string            `json:"mode,omitempty"`
}

// RabbitMQInput --
type RabbitMQInput struct {
	*queue.Consumer
	Delivery  <-chan amqp.Delivery
	XchName   string
	QueueName string
}

func preCheckInputInfo(inp *entity.RmqInputConf) {
	if inp.Exch.Name == "" || inp.Mode == "" {
		log.Fatal("Invalid input information: exchange name or mode is empty")
	}
}

// BuildRabbitmqInput --
func BuildRabbitmqInput(ctx context.Context, groupName string,
	inp *entity.RmqInputConf) (*RabbitMQInput, error) {
	preCheckInputInfo(inp)

	// create queue name if needed
	if inp.Queue.Name == "" {
		var queueName string
		var gName = groupName
		if groupName == "" {
			gName, _ = string_utils.GenerateRandomString(10)
			// auto delete must be set to true if group name is empty
			inp.Queue.AutoDelete = true
		}
		queueName = fmt.Sprintf("%v_%v", inp.Exch.Name, gName)
		inp.Queue.Name = queueName
	}

	uri := viper.GetString("rabbitmq.uri")
	consumer := queue.NewRMQConsumerFConfig(uri, inp)

	if err := consumer.Connect(); err != nil {
		log.Errorln("Error when connect consumer", "details", err)
		return nil, fmt.Errorf("cannot connect to rabbitmq. %v", err)
	}

	deliveries, err := consumer.AnnounceQueue()
	if err != nil {
		log.Errorln("Error when calling AnnounceQueue()", "errError()", err.Error())
		return nil, fmt.Errorf("cannot create or consume from exch/queue. %v. %v", inp.Exch, inp.Queue)
	}

	return &RabbitMQInput{Consumer: consumer, Delivery: deliveries, XchName: inp.Exch.Name,
		QueueName: inp.Queue.Name}, nil
}
