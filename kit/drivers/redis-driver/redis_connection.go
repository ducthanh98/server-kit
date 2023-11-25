package redis_driver

import (
	"encoding/json"
	"github.com/ducthanh98/server-kit/kit/config"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

const (
	DefaultTickerInterval       = 6 * 1e3 // 6 seconds in millisecond
	DefaultSlowQueriesThreshold = 10      // 10 query
	DefaultWarnThreshold        = 50      // 50 millisecond
)

type Redis struct {
	redis.UniversalClient
	IsRedisSlow bool
}

type DetectSlowRedisQueryConf struct {
	Enabled              bool
	SlowQueriesCount     int64
	SlowQueriesThreshold int64
	TickerInterval       int64
	WarnThreshold        int64
	SlackChannelUrl      string
	EnableLog            bool
}

// NewConnection -- open connection to db
func NewConnection(conn Connection) (*Redis, error) {
	var err error

	c, err := conn.BuildRedisClient()
	if err != nil {
		log.Errorln("Could not build redis client", "err", err)
		return nil, err
	}

	pong, err := c.Ping().Result()
	if err != nil {
		log.Errorln("Could not ping to redis", "err", err)
		return nil, err
	}
	log.Info("Ping to redis: ", pong)

	return getRedis(c), nil
}

func getRedis(c redis.UniversalClient) *Redis {
	rd := &Redis{c, false}
	conf := getSlowQueriesDetectConf()
	log.Info("Start redis check slow query ", conf)
	podName := config.GetPodName()
	if conf.Enabled {
		go func() {
			ticker := time.NewTicker(time.Duration(conf.TickerInterval) * time.Millisecond)
			for range ticker.C {
				start := time.Now()
				warnThreshold := time.Duration(conf.WarnThreshold) * time.Millisecond
				_, err := c.Ping().Result()
				elapsed := time.Since(start)

				if conf.EnableLog {
					payload, _ := json.Marshal(conf)
					log.Infof("Is slow: %v .Check redis took %v. conf: %v", rd.IsRedisSlow, elapsed, string(payload))
				}

				if err != nil {
					if !rd.IsRedisSlow {
						log.Errorln("Could not ping to redis", "err", err)
					}
					rd.IsRedisSlow = true
					continue
				}

				if elapsed > warnThreshold {
					conf.SlowQueriesCount++
				} else {
					if rd.IsRedisSlow {
					}
					rd.IsRedisSlow = false
					conf.SlowQueriesCount = 0
				}

				if !rd.IsRedisSlow && conf.SlowQueriesCount >= conf.SlowQueriesThreshold {
					// slow query or redis down
					rd.IsRedisSlow = true
					log.Info("[RD_DBG] Redis slow ", podName)
				}
			}
		}()
	}
	return rd
}
func getSlowQueriesDetectConf() *DetectSlowRedisQueryConf {
	conf := &DetectSlowRedisQueryConf{
		Enabled:              viper.GetBool("redis.enable_slow_detect"),
		SlowQueriesThreshold: viper.GetInt64("redis.slow_query_count_threshold"), // number of slow query for alert
		TickerInterval:       viper.GetInt64("redis.ticker_interval"),
		WarnThreshold:        viper.GetInt64("redis.slow_query_time_threshold"), // threshold for detect a query is slow or not
		SlackChannelUrl:      viper.GetString("redis.slow_alert_slack_channel"),
		EnableLog:            viper.GetBool("redis.enable_slow_detect_log"),
	}

	if conf.TickerInterval == 0 {
		conf.TickerInterval = DefaultTickerInterval
	}

	if conf.WarnThreshold == 0 {
		conf.WarnThreshold = DefaultWarnThreshold
	}

	if conf.SlowQueriesThreshold == 0 {
		conf.SlowQueriesThreshold = DefaultSlowQueriesThreshold
	}
	return conf
}

// NewConnectionFromExistedClient --
func NewConnectionFromExistedClient(c redis.UniversalClient) *Redis {
	return &Redis{c, false}
}

// Close -- close connection
func (br *Redis) Close() error {
	if br != nil {
		return br.UniversalClient.Close()
	}

	return nil
}
