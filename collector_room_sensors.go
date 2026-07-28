package main

import (
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/redkite127/home-hub/homeassistant"
	"github.com/redkite127/home-hub/hue"
)

type roomKey struct {
	room       string
	sensorType string
}

type roomState struct {
	temperature *float64
	humidity    *float64
	battery     *float64

	timestamp time.Time
}

func collectAndSendRoomData() error {
	ts, err := collectRoomData()
	if err != nil {
		log.Printf("failed to collect room data: %s", err)
		return nil // we don't want to interrupt everything else
	}
	sendRoomData(ts)

	return nil
}

func collectRoomData() (rs map[roomKey]roomState, err error) {
	now := time.Now().UTC()
	rs = map[roomKey]roomState{}

	// collect data from Philips HUE motion sensors
	{
		temperatures, err := hue.GetTemperatures()
		if err != nil {
			return map[roomKey]roomState{}, err
		}
		for room, t := range temperatures {
			key := roomKey{room: room, sensorType: "hue"}
			rstate := rs[key]
			rt := t
			rstate.temperature = &rt
			rstate.timestamp = now
			rs[key] = rstate
		}

		batteries, err := hue.GetBatteries()
		if err != nil {
			return map[roomKey]roomState{}, err
		}
		for room, b := range batteries {
			key := roomKey{room: room, sensorType: "hue"}
			rstate := rs[key]
			rb := b
			rstate.battery = &rb
			rstate.timestamp = now
			rs[key] = rstate
		}
	}

	// collect data from Home Assistant room sensors (e.g. IKEA Matter)
	{
		readings, err := homeassistant.GetRoomSensors()
		if err != nil {
			return map[roomKey]roomState{}, err
		}
		for room, r := range readings {
			key := roomKey{room: room, sensorType: "homeassistant"}
			rstate := rs[key]
			if r.Temperature != nil {
				rstate.temperature = r.Temperature
			}
			if r.Humidity != nil {
				rstate.humidity = r.Humidity
			}
			if r.Battery != nil {
				rstate.battery = r.Battery
			}
			rstate.timestamp = now
			rs[key] = rstate
		}
	}

	return
}

func sendRoomData(rs map[roomKey]roomState) {
	for key, state := range rs {
		p := influxdb2.NewPointWithMeasurement("room_sensors")
		p.AddTag("room", key.room)
		p.AddTag("type", key.sensorType)
		addFieldRounded(p, "temperature", state.temperature, 1)
		addFieldRounded(p, "humidity", state.humidity, 1)
		addFieldRounded(p, "battery", state.battery, 0)
		if len(p.FieldList()) == 0 {
			continue // every sensor of this room failed; InfluxDB rejects points without fields
		}
		p.SetTime(state.timestamp)
		writePoint(p)
	}

	flushPoints()
}
