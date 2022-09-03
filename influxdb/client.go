package influxdb

import (
	"sync"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var client influxdb2.Client
var once sync.Once

func initClient() {
	client = influxdb2.NewClient(config.URL, config.Token)
}

func GetClient() influxdb2.Client {
	once.Do(initClient)

	return client
}
