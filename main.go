package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SensorRecord struct {
	Timestamp   time.Time
	Temperature float64
	Humidity    float64
}

var sensors sync.Map

var addr = flag.String("listen-address", ":1234", "The address to listen on for HTTP requests.")

func sensorsHandler(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	if r.Method == "POST" {
		//name := r.URL.Query().Get("name")
		t := r.URL.Query().Get("type")
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Failed to read body!")
			return
		}

		if t == "temperature;humidity" {
			var sr SensorRecord
			sr.Timestamp = time.Now()
			if _, err := fmt.Sscanf(string(data), "%f;%f", &sr.Temperature, &sr.Humidity); err != nil {
				log.Println("Failed to parse body!")
				return
			}

			sensors.Store(room, sr)
		}
	} else if r.Method == "GET" {
		var sr SensorRecord
		if v, ok := sensors.Load(room); !ok {
			log.Println("No temperature value found for ", room)
		} else {
			sr = v.(SensorRecord)
		}

		fmt.Fprintf(w, "%.2f\n%.2f\n%s", sr.Temperature, sr.Humidity, sr.Timestamp.UTC().Format(time.RFC3339))
	}
}

func main() {
	// Fake data ==> http://localhost:1234/sensors?room=kitchen
	sr := SensorRecord{Timestamp: time.Now(), Temperature: 21.64, Humidity: 54.98}
	sensors.Store("kitchen", sr)

	opts := prometheus.GaugeOpts{
		Name: "kitchen_temperature_celcius",
		Help: "The kitchen temperature in degree celcius.",
	}
	gf := prometheus.NewGaugeFunc(opts, func() float64 {
		if v, ok := sensors.Load("kitchen"); !ok {
			return 0.0
		} else {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			return v.(SensorRecord).Temperature + float64(r.Intn(200)-100)/100
		}
	})
	prometheus.Register(gf)

	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/sensors", sensorsHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
