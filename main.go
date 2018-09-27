package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
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

// PowerUsageRecord structure holds the amount of watts consummed in a period
type PowerUsageRecord struct {
	From  time.Time
	To    time.Time
	Watts float64
}

const (
	sensor_namespace  = "sensor"
	house_power_usage = "house_power_usage_watts"
	program           = "sensor_record"
)

var sensors = map[string]SensorRecord{}
var sensors_mutex = &sync.Mutex{}

var housePowerUsage struct {
	L1   float64
	L2   float64
	L3   float64
	From time.Time
	To   time.Time
}
var housePowerUsage_mutex = &sync.Mutex{}

var addr = flag.String("listen-address", ":2001", "The address to listen on for HTTP requests.")
var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(sensor_namespace, "scrape", "collector_duration_seconds"),
		"sensor_exporter: Duration of a collector scrape.",
		nil,
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(sensor_namespace, "scrape", "collector_success"),
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
	now := time.Now().UTC()

	// Returns sensor records
	sensors_mutex.Lock()
	defer sensors_mutex.Unlock()

	for k, s := range sensors {
		ch <- prometheus.NewMetricWithTimestamp(s.Timestamp,
			prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					sensor_namespace+"_temperature",
					"The temperature of the room",
					[]string{"room"},
					nil),
				prometheus.GaugeValue,
				s.Temperature,
				k,
			),
		)
		if s.Humidity != 0 {
			ch <- prometheus.NewMetricWithTimestamp(s.Timestamp,
				prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						sensor_namespace+"_humidity",
						"The humidity of the room",
						[]string{"room"},
						nil),
					prometheus.GaugeValue,
					s.Humidity,
					k,
				),
			)
		}
		if s.Power != 0 {
			ch <- prometheus.NewMetricWithTimestamp(s.Timestamp,
				prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						sensor_namespace+"_power",
						"The power of the room",
						[]string{"room"},
						nil),
					prometheus.GaugeValue,
					s.Power,
					k,
				),
			)
		}

		delete(sensors, k)
	}

	// Returns power records
	housePowerUsage_mutex.Lock()
	defer housePowerUsage_mutex.Unlock()

	meanTime := housePowerUsage.From.Add(housePowerUsage.To.Sub(housePowerUsage.From) / 2)
	description := prometheus.NewDesc(
		house_power_usage,
		"The usage of electrical lines",
		[]string{"line"},
		nil)

	if !housePowerUsage.To.IsZero() {
		ch <- prometheus.NewMetricWithTimestamp(meanTime,
			prometheus.MustNewConstMetric(
				description,
				prometheus.GaugeValue,
				housePowerUsage.L1,
				"L1",
			),
		)
		ch <- prometheus.NewMetricWithTimestamp(meanTime,
			prometheus.MustNewConstMetric(
				description,
				prometheus.GaugeValue,
				housePowerUsage.L2,
				"L2",
			),
		)
		ch <- prometheus.NewMetricWithTimestamp(meanTime,
			prometheus.MustNewConstMetric(
				description,
				prometheus.GaugeValue,
				housePowerUsage.L3,
				"L3",
			),
		)

		housePowerUsage.From = housePowerUsage.To
		housePowerUsage.To = time.Time{}
		housePowerUsage.L1 = 0
		housePowerUsage.L2 = 0
		housePowerUsage.L3 = 0
	}

	duration := time.Since(now)

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
	now := time.Now().UTC()
	room := r.URL.Query().Get("room")
	if r.Method == "POST" {
		t := r.URL.Query().Get("type")
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorln("Failed to read body!")
			return
		}

		if t == "L1;L2;L3" {
			var l1, l2, l3 float64
			if _, err = fmt.Sscanf(string(data), "%f;%f;%f", &l1, &l2, &l3); err != nil {
				log.Errorln("Failed to parse body!")
				return
			}

			housePowerUsage_mutex.Lock()
			defer housePowerUsage_mutex.Unlock()

			if housePowerUsage.From.IsZero() {
				// I received an amount of power consummed, but I don't know since when...
				// ==> trash it, but init From
				housePowerUsage.From = now
				return
			} else if now.Sub(housePowerUsage.From) > 2*time.Minute {
				// Too much time between from & to, there was probably a problem
				// ==> trash it, but init From
				housePowerUsage.From = now
				return
			} else {
				housePowerUsage.L1 += l1
				housePowerUsage.L2 += l2
				housePowerUsage.L3 += l3
				housePowerUsage.To = now
			}
		} else {
			var sr SensorRecord
			sr.Timestamp = now
			if t == "temperature" {
				_, err = fmt.Sscanf(string(data), "%f", &sr.Temperature)
			} else if t == "temperature;humidity" {
				_, err = fmt.Sscanf(string(data), "%f;%f", &sr.Temperature, &sr.Humidity)
			} else if t == "temperature;power" {
				_, err = fmt.Sscanf(string(data), "%f;%f", &sr.Temperature, &sr.Power)
			} else if t == "temperature;humidity;power" {
				_, err = fmt.Sscanf(string(data), "%f;%f;%f", &sr.Temperature, &sr.Humidity, &sr.Power)
			} else {
				log.Errorln("Unkown type!")
				return
			}

			if err != nil {
				log.Errorln("Failed to parse body!")
				return
			}

			sensors_mutex.Lock()
			sensors[room] = sr
			sensors_mutex.Unlock()
		}

		log.Debugln("Stored record for room: ", room)
	} else if r.Method == "GET" {
		var sr SensorRecord
		var ok bool
		sensors_mutex.Lock()
		if sr, ok = sensors[room]; !ok {
			log.Infoln("No temperature value found for room: ", room)
			return
		}
		sensors_mutex.Unlock()

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
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
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
