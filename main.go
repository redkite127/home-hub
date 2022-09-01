package main

import (
	"context"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/oklog/run"
)

func main() {
	initInfluxClient()
	defer influxC.Close()

	var g run.Group

	// launch HTTP server for receiving values from sensors
	{
		s := &http.Server{
			Addr: ":2001", //TODO get this value from config file
		}
		g.Add(
			func() error { return s.ListenAndServe() },
			func(error) { s.Shutdown(context.Background()) },
		)
		defer s.Close()
	}

	// regularly collect and then record electrical data in InfluxDB
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(
			func() error { return scheduler(ctx, collectAndSendElectricalData, time.Minute) },
			func(error) { cancel() },
		)
	}

	// listen interruption signals
	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	log.Printf("application terminated: %v", g.Run())
}
