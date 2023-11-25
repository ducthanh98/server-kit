package redis_driver

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Connection interface {
	BuildRedisClient() (redis.UniversalClient, error)
}

var (
	ErrorMissingRedisAddress = errors.New("missing redis address")
)

const (
	// DefaultPoolSize --
	DefaultPoolSize     = 100
	DefaultReadTimeout  = 1000
	DefaultWriteTimeout = 1000
)

// DefaultRedisConnectionFromConfig -- load connection settings in conf with default key
func DefaultRedisConnectionFromConfig() Connection {
	poolSize := viper.GetInt("redis.pool_size")
	if poolSize <= 0 {
		poolSize = DefaultPoolSize
	}

	return &SingleConnection{
		address:      viper.GetString("redis.address"),
		password:     viper.GetString("redis.password"),
		db:           viper.GetInt("redis.db"),
		readTimeout:  viper.GetInt("redis.read_timeout"),
		writeTimeout: viper.GetInt("redis.write_timeout"),
		poolSize:     poolSize,
	}
}

// NewRedisConfig --
func NewRedisConfig(add string, db int) Connection {
	return &SingleConnection{
		address:      add,
		db:           db,
		poolSize:     DefaultPoolSize,
		readTimeout:  viper.GetInt("redis.read_timeout"),
		writeTimeout: viper.GetInt("redis.write_timeout"),
	}
}

// NewRedisConfigWithPool --
func NewRedisConfigWithPool(add string, db, poolSize int) Connection {
	return &SingleConnection{
		address:      add,
		db:           db,
		poolSize:     poolSize,
		readTimeout:  viper.GetInt("redis.read_timeout"),
		writeTimeout: viper.GetInt("redis.write_timeout"),
	}
}
