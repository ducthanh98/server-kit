package app

import (
	"github.com/ducthanh98/server-kit/kit/drivers"
	rabbitmq_driver "github.com/ducthanh98/server-kit/kit/drivers/rabbitmq-driver"
	redis_driver "github.com/ducthanh98/server-kit/kit/drivers/redis-driver"
	"github.com/ducthanh98/server-kit/kit/logger"
	"github.com/ducthanh98/server-kit/kit/utils/types"
	"gorm.io/gorm"
)

// CreateSqlConnection --
func (app *App) CreateSqlConnection(conn interface{}, driver types.Driver) *gorm.DB {
	factory := drivers.InitConnection(driver)
	db, err := factory.NewConnection(conn)
	if err != nil {
		logger.Log.Panic(err)
	}

	return db
}

// CreateRedisConnection --
func (app *App) CreateRedisConnection(conn redis_driver.Connection) *redis_driver.Redis {
	if conn == nil {
		conn = redis_driver.DefaultRedisConnectionFromConfig()
	}
	db, err := redis_driver.NewConnection(conn)
	if err != nil {
		logger.Log.Panic(err)
	}
	return db
}

// CreateRabbitMQProducerConnection creates new producer connection
func (app *App) CreateRabbitMQProducerConnection(conn *rabbitmq_driver.RabbitMqConfiguration) *rabbitmq_driver.Producer {
	internalQueueSize := 1000
	retries := 10
	numThread := 1
	if conn == nil {
		misc := rabbitmq_driver.LoadOtherConfigParams()
		exch := rabbitmq_driver.LoadExchConfig()
		conn = &rabbitmq_driver.RabbitMqConfiguration{URI: misc["uri"].(string), ExchangeConfig: *exch}
		internalQueueSize = misc["internal_queue_size"].(int)
		retries = misc["retries"].(int)
		numThread, _ = misc["num_thread"].(int)

		if numThread <= 0 {
			numThread = 1
		}
	}

	p := rabbitmq_driver.NewRMQProducerFromConf(conn, internalQueueSize, retries)

	if err := p.ConnectMulti(numThread); err != nil {
		logger.Log.Panic(err)
	}
	p.Start()

	return p
}
