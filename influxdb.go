package main

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var influxC influxdb2.Client

func initInfluxClient() {
	//TODO get this value from config file
	influxC = influxdb2.NewClient("http://localhost:8086", "apnP6gCZ3E0XaE_nIqjaGcznm5h0yaK8kS0Y3hSxa4hPcNjrbxrfmLW81yufCm1Dp-VGcscj3c780vDP4hwPXQ==")
	influxEnergyW = influxC.WriteAPI("baclain_28", "home_hub")
}
