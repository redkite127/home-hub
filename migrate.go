package main

import (
	"encoding/json"
	"log"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

// At the end of init:

// start := time.Date(2018, 01, 01, 00, 00, 00, 00, time.UTC)
// //start := time.Date(2018, 9, 25, 0, 0, 0, 0, time.UTC)
// for start.Before(time.Now()) {
// 	end := start.AddDate(0, 1, 0)
// 	migrate1(start.Format("2006-01-02 15:04:05.999999999"), end.Format("2006-01-02 15:04:05.999999999"))
// 	migrate2(start.Format("2006-01-02 15:04:05.999999999"), end.Format("2006-01-02 15:04:05.999999999"), "sensor_temperature")
// 	migrate2(start.Format("2006-01-02 15:04:05.999999999"), end.Format("2006-01-02 15:04:05.999999999"), "sensor_humidity")
// 	migrate2(start.Format("2006-01-02 15:04:05.999999999"), end.Format("2006-01-02 15:04:05.999999999"), "sensor_power")
// 	start = end
// }

func migrate1(from, to string) {
	log.Println("SELECT time, __name__, room, f64  FROM _ WHERE room != '' AND time >= '" + from + "' AND time < '" + to + "'")
	q := client.Query{
		Command:  "SELECT time, __name__, room, f64  FROM _ WHERE room != '' AND time >= '" + from + "' AND time < '" + to + "'",
		Database: "prometheus",
	}
	if response, err := c.Query(q); err == nil {
		if response.Error() != nil {
			log.Fatal(response.Error())
		}
		//log.Print(response.Results[0])

		oldMeasures := map[string]map[time.Time]SensorRecord{}
		if len(response.Results) > 0 && len(response.Results[0].Series) > 0 {
			for _, row := range response.Results[0].Series[0].Values {
				timeStr := row[0].(string)
				name := row[1].(string)
				room := row[2].(string)
				value, _ := row[3].(json.Number).Float64()

				t, _ := time.Parse(time.RFC3339Nano, timeStr)
				//log.Println(i, row)

				if _, ok := oldMeasures[room]; !ok {
					oldMeasures[room] = make(map[time.Time]SensorRecord)
				}

				if name == "sensor_temperature" {
					sr := oldMeasures[room][t]
					sr.Temperature = &value
					oldMeasures[room][t] = sr
				} else if name == "sensor_humidity" {
					sr := oldMeasures[room][t]
					sr.Humidity = &value
					oldMeasures[room][t] = sr
				} else if name == "sensor_power" {
					sr := oldMeasures[room][t]
					sr.Power = &value
					oldMeasures[room][t] = sr
				}
			}
		}

		//Insert them in the new measurements
		// Create a new point batch
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  "home_hub",
			Precision: "s",
		})
		if err != nil {
			log.Fatal(err)
		}

		for room, measures := range oldMeasures {
			for t, s := range measures {
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
					map[string]string{"room": room},
					fields,
					t)
				if err != nil {
					log.Fatal(err)
				}
				bp.AddPoint(pt)
			}
		}

		if len(bp.Points()) > 0 {
			// Write the batch
			if err := c.Write(bp); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		log.Fatal(response.Error())
	}
	log.Printf("migrated from %s to %s", from, to)
}

func migrate2(from, to, table string) {
	q := client.Query{
		Command:  "SELECT time, __name__, room, value FROM " + table + " WHERE time >= '" + from + "' AND time < '" + to + "'",
		Database: "prometheus",
	}
	if response, err := c.Query(q); err == nil {
		if response.Error() != nil {
			log.Fatal(response.Error())
		}
		//log.Print(response.Results[0])

		oldMeasures := map[string]map[time.Time]SensorRecord{}
		if len(response.Results) > 0 && len(response.Results[0].Series) > 0 {
			for _, row := range response.Results[0].Series[0].Values {
				timeStr := row[0].(string)
				name := row[1].(string)
				room := row[2].(string)
				value, _ := row[3].(json.Number).Float64()

				t, _ := time.Parse(time.RFC3339Nano, timeStr)
				//log.Println(i, row)

				if _, ok := oldMeasures[room]; !ok {
					oldMeasures[room] = make(map[time.Time]SensorRecord)
				}

				if name == "sensor_temperature" {
					sr := oldMeasures[room][t]
					sr.Temperature = &value
					oldMeasures[room][t] = sr
				} else if name == "sensor_humidity" {
					sr := oldMeasures[room][t]
					sr.Humidity = &value
					oldMeasures[room][t] = sr
				} else if name == "sensor_power" {
					sr := oldMeasures[room][t]
					sr.Power = &value
					oldMeasures[room][t] = sr
				}
			}
		}

		//Insert them in the new measurements
		// Create a new point batch
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  "home_hub",
			Precision: "s",
		})
		if err != nil {
			log.Fatal(err)
		}

		for room, measures := range oldMeasures {
			for t, s := range measures {
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
					map[string]string{"room": room},
					fields,
					t)
				if err != nil {
					log.Fatal(err)
				}
				bp.AddPoint(pt)
			}
		}

		if len(bp.Points()) > 0 {
			// Write the batch
			if err := c.Write(bp); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		log.Fatal(response.Error())
	}
	log.Printf("migrated from %s to %s", from, to)
}
