package influxdb

import (
	"fmt"

	"github.com/spf13/viper"
)

var config struct {
	URL          string `mapstructure:"url"`
	Token        string `mapstructure:"token"`
	Organization string `mapstructure:"organization"`
	Bucket       string `mapstructure:"bucket"`
}

func InitConfig() {
	if err := viper.UnmarshalKey("influxdb", &config); err != nil {
		panic(fmt.Errorf("fatal error initializing InfluxDB config: %w", err))
	}
}
