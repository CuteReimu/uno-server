package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var GlobalConfig *viper.Viper

func init() {
	GlobalConfig = viper.New()
	GlobalConfig.SetConfigName("config")
	GlobalConfig.SetConfigType("yaml")
	GlobalConfig.AddConfigPath(".")
	err := GlobalConfig.ReadInConfig()
	if err != nil {
		logrus.WithError(err).Fatalln("unable to write logs")
	}
}
