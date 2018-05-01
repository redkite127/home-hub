package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

// SensorRecordExporter is the metrics exporter object.
// Currently does not contain anything
type SensorRecordExporter struct{}

// SensorRecord structure defining the values of a sensor
type SensorRecord struct {
	Timestamp   time.Time
	Temperature float64
	Humidity    float64
	Power       float64
}

const (
	namespace = "sensor"
	program   = "sensor_record"
)

var sensors sync.Map

var addr = flag.String("listen-address", ":2001", "The address to listen on for HTTP requests.")
var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"sensor_exporter: Duration of a collector scrape.",
		nil,
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"sensor_exporter: Whether the collector succeeded.",
		nil,
		nil,
	)
)

// Describe implements the prometheus.Collector interface
func (e *SensorRecordExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface
func (e *SensorRecordExporter) Collect(ch chan<- prometheus.Metric) {
	begin := time.Now()

	sensors.Range(func(k, v interface{}) bool {
		// Don't send this metric if it is too old (probe not comunicating anymore)
		if time.Now().Sub(v.(SensorRecord).Timestamp) > 15*time.Minute {
			return true
		}

		t := v.(SensorRecord).Temperature
		h := v.(SensorRecord).Humidity
		p := v.(SensorRecord).Power

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "temperature"),
				"The temperature of the room",
				[]string{"room"},
				nil),
			prometheus.GaugeValue,
			t, k.(string))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "humidity"),
				"The humidity of the room",
				[]string{"room"},
				nil),
			prometheus.GaugeValue,
			h, k.(string))
		if p != 0 {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "power"),
					"The power of the room",
					[]string{"room"},
					nil),
				prometheus.GaugeValue,
				p, k.(string))
		}
		return true
	})

	duration := time.Since(begin)

	var err error
	err = nil

	var success float64
	if err != nil {
		log.Errorf("ERROR: collector failed after %fs: %s", duration.Seconds(), err)
		success = 0
	} else {
		log.Debugf("OK: collector succeeded after %fs", duration.Seconds())
		success = 1
	}

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds())
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success)
}

func sensorsHandler(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	if r.Method == "POST" {
		t := r.URL.Query().Get("type")
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorln("Failed to read body!")
			return
		}

		if t == "temperature;humidity" {
			var sr SensorRecord
			sr.Timestamp = time.Now()
			if _, err := fmt.Sscanf(string(data), "%f;%f", &sr.Temperature, &sr.Humidity); err != nil {
				log.Errorln("Failed to parse body!")
				return
			}

			sensors.Store(room, sr)
			log.Debugln("Stored record for room: ", room)
		} else if t == "temperature;humidity;power" {
			var sr SensorRecord
			sr.Timestamp = time.Now()
			if _, err := fmt.Sscanf(string(data), "%f;%f;%f", &sr.Temperature, &sr.Humidity, &sr.Power); err != nil {
				log.Errorln("Failed to parse body!")
				return
			}

			sensors.Store(room, sr)
			log.Debugln("Stored record for room: ", room)
		}
	} else if r.Method == "GET" {
		var sr SensorRecord
		if v, ok := sensors.Load(room); !ok {
			log.Infoln("No temperature value found for room: ", room)
		} else {
			sr = v.(SensorRecord)
		}

		if sr.Power == 0 {
			fmt.Fprintf(w, "%.2f\n%.2f\n%s", sr.Temperature, sr.Humidity, sr.Timestamp.String())
		} else {
			fmt.Fprintf(w, "%.2f\n%.2f\n%.2f\n%s", sr.Temperature, sr.Humidity, sr.Power, sr.Timestamp.String())
		}
	}
}

func init() {
	// TODO - add this if you want to create a versioning of your program
	// prometheus.MustRegister(version.NewCollector(program))
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	// PUSH EXAMPLES
	// sensors.Store("kitchen", SensorRecord{Timestamp: time.Now(), Temperature: 21.64, Humidity: 54.98})
	// sensors.Store("hall", SensorRecord{Timestamp: time.Now(), Temperature: 20.01, Humidity: 54.12})
	// sensors.Store("bedroom", SensorRecord{Timestamp: time.Now(), Temperature: 16.85, Humidity: 45.82})
	// sensors.Store("laundry", SensorRecord{Timestamp: time.Now(), Temperature: 17.15, Humidity: 45.82})
}

func main() {
	log.Infoln("Starting sensor reader/exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	exporter := &SensorRecordExporter{}
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", prometheus.UninstrumentedHandler())
	http.HandleFunc("/sensors", sensorsHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
                        <head><title>SensorRecord Exporter</title></head>
                        <body>
                        <h1>SensorRecord Exporter</h1>
                        <p><a href='/metrics'>Metrics</a></p>
                        <p><a href='/sensors'>Sensors</a></p>
                        </body>
                        </html>`))
	})
	log.Fatal(http.ListenAndServe(*addr, nil))
}
