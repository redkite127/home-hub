package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"syscall"

	influxdb2_api "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/oklog/run"
	"github.com/redkite127/home-hub/homeassistant"
	"github.com/redkite127/home-hub/hue"
	"github.com/redkite127/home-hub/influxdb"
	"github.com/spf13/viper"
)

var influxWriter influxdb2_api.WriteAPI
var influxErrors <-chan error

func init() {
	viper.SetConfigName("default")
	viper.AddConfigPath("/usr/local/etc/")
	viper.AddConfigPath("./configs/")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	homeassistant.InitConfig()
	hue.InitConfig()
	hue.InitDevices()

	influxdb.InitConfig()
	influxWriter = influxdb.GetWriter()
}

func main() {
	defer influxdb.GetClient().Close()

	var g run.Group

	// launch HTTP server for receiving values from sensors
	{
		s := &http.Server{
			Addr: ":" + viper.GetString("port"),
		}
		g.Add(
			func() error {
				log.Println("started HTTP listening on port", viper.GetString("port"))
				err := s.ListenAndServe()
				log.Println("stopped HTTP listening")

				return err
			},
			func(error) { s.Shutdown(context.Background()) },
		)
		defer s.Close()
	}

	// regularly collect and then record electrical data in InfluxDB
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(
			func() error {
				log.Println("started collecting electrical data every", viper.GetDuration("frequencies.electrical_data"))
				err := scheduler(ctx, collectAndSendElectricalData, viper.GetDuration("frequencies.electrical_data"))
				log.Println("stopped collecting electrical data")

				return err
			},
			func(error) { cancel() },
		)
	}

	// regularly collect and then record room data in InfluxDB (HUE)
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(
			func() error {
				log.Println("started collecting room data every", viper.GetDuration("frequencies.room_data"))
				err := scheduler(ctx, collectAndSendRoomData, viper.GetDuration("frequencies.room_data"))
				log.Println("stopped collecting room data")

				return err
			},
			func(error) { cancel() },
		)
	}

	// listen interruption signals
	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	log.Printf("application terminated: %v", g.Run())
}
