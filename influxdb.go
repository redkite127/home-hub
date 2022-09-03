package main

import (
	"log"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/viper"
)

var influxC influxdb2.Client

func initInfluxClient() {
	config := viper.Sub("influxdb")

	influxC = influxdb2.NewClient(config.GetString("url"), config.GetString("token"))
	influxEnergyW = influxC.WriteAPI(config.GetString("organization"), "home_hub_test")

	log.Println("initialized InfluxDB client:", influxC.ServerURL())
}
