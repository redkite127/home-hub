package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("default")
	viper.AddConfigPath("./configs/")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func main() {
	initInfluxClient()
	defer influxC.Close()

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
				err := scheduler(ctx, collectAndSendElectricalData, time.Minute)
				log.Println("stopped collecting electrical data")

				return err
			},
			func(error) { cancel() },
		)
	}

	// listen interruption signals
	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	log.Printf("application terminated: %v", g.Run())
}
