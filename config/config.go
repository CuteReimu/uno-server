package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log/slog"
)

var GlobalConfig *viper.Viper

func init() {
	GlobalConfig = viper.New()
	GlobalConfig.SetConfigName("config")
	GlobalConfig.SetConfigType("yaml")
	GlobalConfig.AddConfigPath(".")
	err := GlobalConfig.ReadInConfig()
	if err != nil {
		slog.Error("unable to write logs", "error", err)
		panic(fmt.Sprintf("unable to write logs, error: %+v", err))
	}
}
