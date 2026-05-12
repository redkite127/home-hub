package homeassistant

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type RoomSensorEntities struct {
	Temperature string `mapstructure:"temperature"`
	Humidity    string `mapstructure:"humidity"`
	Battery     string `mapstructure:"battery"`
}

var config struct {
	URL         string                        `mapstructure:"url"`
	Token       string                        `mapstructure:"token"`
	RoomSensors map[string]RoomSensorEntities `mapstructure:"room_sensors"`
}

var client = &http.Client{Timeout: 10 * time.Second}

func InitConfig() {
	if err := viper.UnmarshalKey("homeassistant", &config); err != nil {
		panic(fmt.Errorf("fatal error initializing Home Assistant config: %w", err))
	}
}
