package homeassistant

import (
	"fmt"

	"github.com/spf13/viper"
)

var config struct {
	URL   string `mapstructure:"url"`
	Token string `mapstructure:"token"`
}

func InitConfig() {
	if err := viper.UnmarshalKey("homeassistant", &config); err != nil {
		panic(fmt.Errorf("fatal error initializing Home Assistant config: %w", err))
	}
}
