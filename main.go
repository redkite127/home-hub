package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
)

// SensorRecordExporter is the metrics exporter object.
// Currently does not contain anything
type SensorRecordExporter struct{}

// SensorRecord structure defining the values of a sensor
type SensorRecord struct {
	Timestamp   time.Time
	Temperature *float64
	Humidity    *float64
	Power       *float64
}

// PowerUsageRecord structure holds the amount of watts consummed in a period
type PowerUsageRecord struct {
	From  time.Time
	To    time.Time
	Watts float64
}

type RoomSensor struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity,omitempty"`
	Battery     float64 `json:"battery,omitempty"`
}

const (
	sensor_namespace  = "sensor"
	house_power_usage = "house_power_usage_watts"
	program           = "sensor_record"
	collect_time      = 5 * time.Minute
)

var sensors = map[string]SensorRecord{}
var lastSensors = map[string]SensorRecord{} //Won't be erased after a collect
var sensors_mutex = &sync.Mutex{}

var housePowerUsage struct {
	L1    float64
	L2    float64
	L3    float64
	From  time.Time
	To    time.Time
	Count int
}
var housePowerUsage_mutex = &sync.Mutex{}

var addr = flag.String("listen-address", ":2001", "The address to listen on for HTTP requests.")

func collect() {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "home_hub",
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Returns sensor records
	sensors_mutex.Lock()
	defer sensors_mutex.Unlock()

	for k, s := range sensors {
		fields := map[string]interface{}{}
		if s.Temperature != nil {
			fields["temperature"] = *s.Temperature
		}
		if s.Humidity != nil {
			fields["humidity"] = *s.Humidity
		}
		if s.Power != nil {
			fields["power"] = *s.Power
		}

		pt, err := client.NewPoint(
			"sensors",
			map[string]string{"room": k},
			fields,
			s.Timestamp)
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)

		delete(sensors, k)
	}

	// Returns power records
	housePowerUsage_mutex.Lock()
	defer housePowerUsage_mutex.Unlock()

	if !housePowerUsage.To.IsZero() {
		pt, err := client.NewPoint(
			"house_power",
			map[string]string{},
			map[string]interface{}{
				"L1": housePowerUsage.L1 / float64(housePowerUsage.Count),
				"L2": housePowerUsage.L2 / float64(housePowerUsage.Count),
				"L3": housePowerUsage.L3 / float64(housePowerUsage.Count),
			},
			//housePowerUsage.From.Add(housePowerUsage.To.Sub(housePowerUsage.From)/2),
			housePowerUsage.To,
		)
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)

		housePowerUsage.From = housePowerUsage.To
		housePowerUsage.To = time.Time{}
		housePowerUsage.L1 = 0
		housePowerUsage.L2 = 0
		housePowerUsage.L3 = 0
		housePowerUsage.Count = 0
	}

	if len(bp.Points()) > 0 {
		// Write the batch
		if err := c.Write(bp); err != nil {
			log.Fatal(err)
		}
	}
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
		str := string(data)

		//TODO send a specific frame which send a reset for saying we restarted the probe
		if t == "L1;L2;L3" {
			var l1, l2, l3 float64
			if _, err = fmt.Sscanf(str, "%f;%f;%f", &l1, &l2, &l3); err != nil {
				log.Errorln("Failed to parse body!")
				return
			}

			housePowerUsage_mutex.Lock()
			if housePowerUsage.From.IsZero() {
				// I received an amount of power consummed, but I don't know since when...
				// ==> trash it, reset accumulator
				housePowerUsage.From = now
				housePowerUsage.To = time.Time{}
				housePowerUsage.L1 = 0
				housePowerUsage.L2 = 0
				housePowerUsage.L3 = 0
				housePowerUsage.Count = 0
				log.Debugln("first house power usage received")
			} else if (housePowerUsage.To.IsZero() && now.Sub(housePowerUsage.From) > 2*time.Minute) || (!housePowerUsage.To.IsZero() && now.Sub(housePowerUsage.To) > 2*time.Minute) {
				// Too much time between from & to, there was probably a problem
				// ==> trash it, reset accumulator
				housePowerUsage.From = now
				housePowerUsage.To = time.Time{}
				housePowerUsage.L1 = 0
				housePowerUsage.L2 = 0
				housePowerUsage.L3 = 0
				housePowerUsage.Count = 0
				log.Debugln("staled house power usage received")
			} else {
				housePowerUsage.L1 += l1
				housePowerUsage.L2 += l2
				housePowerUsage.L3 += l3
				housePowerUsage.To = now
				housePowerUsage.Count++
				log.Debugln("house power usage recorded", housePowerUsage)
			}
			housePowerUsage_mutex.Unlock()
		} else {
			var sr SensorRecord
			sr.Timestamp = now
			if t == "temperature" {
				sr.Temperature = new(float64)
				_, err = fmt.Sscanf(str, "%f", sr.Temperature)
			} else if t == "temperature;humidity" {
				sr.Temperature = new(float64)
				sr.Humidity = new(float64)
				_, err = fmt.Sscanf(str, "%f;%f", sr.Temperature, sr.Humidity)
			} else if t == "temperature;power" {
				sr.Temperature = new(float64)
				sr.Power = new(float64)
				_, err = fmt.Sscanf(str, "%f;%f", sr.Temperature, sr.Power)
			} else if t == "temperature;humidity;power" {
				sr.Temperature = new(float64)
				sr.Humidity = new(float64)
				sr.Power = new(float64)
				_, err = fmt.Sscanf(str, "%f;%f;%f", sr.Temperature, sr.Humidity, sr.Power)
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
			lastSensors[room] = sr
			sensors_mutex.Unlock()

			log.Debugf("stored sensor record for room '%v'", room)
		}
	} else if r.Method == "GET" {

		// if len(lastSensors) == 0 {
		// 	f := new(float64)
		// 	*f = 32
		// 	lastSensors["kitchen"] = SensorRecord{Temperature: f}
		// }

		if room == "" {
			//return all rooms
			rooms := map[string]RoomSensor{}

			for k, s := range lastSensors {
				var room RoomSensor
				if s.Temperature != nil {
					room.Temperature = *s.Temperature
				}
				if s.Humidity != nil {
					room.Humidity = *s.Humidity
				}
				if s.Power != nil {
					room.Battery = *s.Power
				}

				rooms[k] = room
			}

			roomsJSON, err := json.Marshal(rooms)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(roomsJSON)
			return
		} else {
			rs, ok := lastSensors[room]
			if !ok {
				http.NotFound(w, r)
				return
			}

			roomJSON, err := json.Marshal(rs)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(roomJSON)
			return
		}
	}
}

var c client.Client

func init() {
	// Create a new HTTPClient
	var err error
	c, err = client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://10.161.0.130:8086",
		// Username: username,
		// Password: password,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	q := client.NewQuery("CREATE DATABASE home_hub", "", "")
	if response, err := c.Query(q); err == nil && response.Error() != nil {
		log.Fatal(response.Error())
	}

	log.SetLevel(log.DebugLevel)
}

func main() {
	http.HandleFunc("/sensors", sensorsHandler)

	ticker := time.NewTicker(collect_time)
	go func() {
		for range ticker.C {
			collect()
		}
	}()

	log.Fatal(http.ListenAndServe(*addr, nil))
}
