package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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
var (
	nodeCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_status",
			Help: "Soa manager service status.",
		},
		[]string{"venture", "service", "status", "resource"},
	)
)

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
    
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())
	prometheus.MustRegister(nodeCounter)

	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/sensors", sensorsHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
