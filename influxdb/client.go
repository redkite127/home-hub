package influxdb

import (
	"log"
	"sync"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2_api "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/spf13/viper"
)

var once sync.Once
var client influxdb2.Client
var writer influxdb2_api.WriteAPI
var errors <-chan error

func initClient() {
	client = influxdb2.NewClient(config.URL, config.Token)
	writer = client.WriteAPI(viper.GetString("influxdb.organization"), viper.GetString("influxdb.bucket"))
	errors = writer.Errors()

	go func() {
		for err := range errors {
			log.Printf("write error: %s\n", err.Error())
		}
	}()
}

func GetClient() influxdb2.Client {
	once.Do(initClient)

	return client
}

func GetWriter() influxdb2_api.WriteAPI {
	once.Do(initClient)

	return writer
}

func GetErrors() <-chan error {
	once.Do(initClient)

	return errors
}
