package output

import (
	"github.com/ducthanh98/server-kit/kit/consumer/entity"
	"github.com/ducthanh98/server-kit/kit/consumer/input"
	queue "github.com/ducthanh98/server-kit/kit/drivers/rabbitmq-driver"
	"github.com/ducthanh98/server-kit/kit/logger"

	"fmt"

	"github.com/spf13/viper"
)

const (
	// RabbitmqMode mode
	RabbitmqMode = "entity"
)

type RmqOutputConf struct {
	Exch       *input.RmqExchQueueInfo `json:"exch,omitempty"`
	Outputdefs []*RmqOutRoutingConf    `json:"outputdefs,omitempty"`
	Mode       string                  `json:"mode,omitempty"`
}

type RmqOutRoutingConf struct {
	RoutingKey string `json:"routing_key,omitempty"`
	Type       string `json:"type,omitempty"`
}

// RabbitMQOutput --
type RabbitMQOutput struct {
	*queue.Producer
}

func preCheckOutputInfo(out *entity.RmqOutputConf) {
	if out.Exch.Name == "" || out.Mode == "" {
		logger.Log.Fatal("Invalid output information: exchange name or mode is empty")
	}
}

// BuildRabbitmqOutput --
func BuildRabbitmqOutput(out *entity.RmqOutputConf) (*RabbitMQOutput, error) {
	preCheckOutputInfo(out)

	uri := viper.GetString("rabbitmq.uri")
	retries := viper.GetInt64("rabbitmq.retries")
	internalQueueSize := viper.GetInt64("rabbitmq.internal_queue_size")
	numThread := viper.GetInt("rabbitmq.num_thread")
	producer := queue.NewRMQProducerFConfig(uri, retries, internalQueueSize, out)

	if numThread <= 0 {
		numThread = queue.MaxThread
	}

	if err := producer.ConnectMulti(numThread); err != nil {
		logger.Log.Errorf("Cannot connect to entity due to error: %v", err)
		return nil, fmt.Errorf("cannot connect to rabbitmq. %v", err)
	}

	producer.Start()

	return &RabbitMQOutput{Producer: producer}, nil
}
