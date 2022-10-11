package hue

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/spf13/viper"
)

var client *http.Client

var config struct {
	URL     string            `mapstructure:"url"`
	Token   string            `mapstructure:"token"`
	Devices map[string]string `mapstructure:"devices"`
}

func InitConfig() {
	if err := viper.UnmarshalKey("philips_hue", &config); err != nil {
		panic(fmt.Errorf("fatal error initializing Philips HUE config: %w", err))
	}

	// deactivate certificate checks for internal HUE bridge
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
}
