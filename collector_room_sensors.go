package main

import (
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
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
		return fmt.Errorf("failed to collect room data: %w", err)
	}
	//log.Println(es)  // TODO log it only in DEBUG mode
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
		if state.battery != nil {
			p.AddField("battery", *state.battery)
		}
		p.SetTime(state.timestamp)
		influxWriter.WritePoint(p)
	}

	influxWriter.Flush()
}
