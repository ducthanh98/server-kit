package config

import (
	"github.com/ducthanh98/server-kit/kit/utils/io"
	"github.com/ducthanh98/server-kit/kit/logger"
	"github.com/spf13/viper"
	"os"
)

func LoadConfig() {
	viper.SetConfigType("toml")
	viper.SetConfigName("conf")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Log.Warnf("Config file not found")
		} else {
			logger.Log.Warnf("Error when loading config file:%v", err)
		}
	}
}

// ReadRawFile --
func ReadRawFile(file string) (string, error) {
	return io.ReadRawFile(file)
}

func GetPodName() string {
	return os.Getenv("HOSTNAME")
}
