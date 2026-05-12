package main

import (
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/redkite127/home-hub/homeassistant"
	"github.com/redkite127/home-hub/hue"
)

type roomState struct {
	temperature *float32
	humidity    *float32
	battery     *float32

	sensorType string
	timestamp  time.Time
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

func collectRoomData() (rs map[string]roomState, err error) {
	now := time.Now().UTC()
	rs = map[string]roomState{}

	// collect data from Philips HUE motion sensors
	{
		temperatures, err := hue.GetTemperatures()
		if err != nil {
			return map[string]roomState{}, err
		}
		for room, t := range temperatures {
			rstate := rs[room]
			rt := t
			rstate.temperature = &rt
			rstate.sensorType = "hue"
			rstate.timestamp = now
			rs[room] = rstate
		}

		batteries, err := hue.GetBatteries()
		if err != nil {
			return map[string]roomState{}, err
		}
		for room, b := range batteries {
			rstate := rs[room]
			rb := b
			rstate.battery = &rb
			rstate.sensorType = "hue"
			rstate.timestamp = now
			rs[room] = rstate
		}
	}

	// collect data from Home Assistant room sensors (e.g. IKEA Matter)
	{
		readings, err := homeassistant.GetRoomSensors()
		if err != nil {
			return map[string]roomState{}, err
		}
		for room, r := range readings {
			rstate := rs[room]
			if r.Temperature != nil {
				rstate.temperature = r.Temperature
			}
			if r.Humidity != nil {
				rstate.humidity = r.Humidity
			}
			if r.Battery != nil {
				rstate.battery = r.Battery
			}
			rstate.sensorType = "homeassistant"
			rstate.timestamp = now
			rs[room] = rstate
		}
	}

	return
}

func sendRoomData(rs map[string]roomState) {
	for room, state := range rs {
		p := influxdb2.NewPointWithMeasurement("room_sensors")
		p.AddTag("room", room)
		p.AddTag("type", state.sensorType)
		if state.temperature != nil {
			p.AddField("temperature", *state.temperature)
		}
		if state.humidity != nil {
			p.AddField("humidity", *state.humidity)
		}
		if state.battery != nil {
			p.AddField("battery", *state.battery)
		}
		p.SetTime(state.timestamp)
		influxWriter.WritePoint(p)
	}

	influxWriter.Flush()
}
