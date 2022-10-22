package main

import (
	"log"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/redkite127/home-hub/homeassistant"
)

type ElectricalState struct {
	energyConsumedDay   float64
	energyConsumedNight float64

	powerConsumptionL1 float64
	powerConsumptionL2 float64
	powerConsumptionL3 float64

	voltageL1 float64
	voltageL2 float64
	voltageL3 float64

	timestamp time.Time
}

func collectAndSendElectricalData() error {
	es, err := collectElectricalData()
	if err != nil {
		log.Printf("failed to collect electrical data: %s", err)
		return nil // we don't want to interrupt everything else
	}
	sendElectricalData(es)

	return nil
}

func collectElectricalData() (es ElectricalState, err error) {
	// retrive time on the first sample
	if entity, err := homeassistant.GetEntityState("sensor.electricity_meter_energy_consumption_tarif_1"); err != nil {
		return ElectricalState{}, err
	} else {
		es.timestamp = entity.LastUpdated

		v, err := strconv.ParseFloat(entity.State, 64)
		if err != nil {
			return ElectricalState{}, err
		}

		es.energyConsumedDay = v
	}

	if es.energyConsumedNight, err = homeassistant.GetEntityStateValueFloat64("sensor.electricity_meter_energy_consumption_tarif_2"); err != nil {
		return ElectricalState{}, err
	}

	if es.powerConsumptionL1, err = homeassistant.GetEntityStateValueFloat64("sensor.electricity_meter_power_consumption_phase_l1"); err != nil {
		return ElectricalState{}, err
	}

	if es.powerConsumptionL2, err = homeassistant.GetEntityStateValueFloat64("sensor.electricity_meter_power_consumption_phase_l2"); err != nil {
		return ElectricalState{}, err
	}

	if es.powerConsumptionL3, err = homeassistant.GetEntityStateValueFloat64("sensor.electricity_meter_power_consumption_phase_l3"); err != nil {
		return ElectricalState{}, err
	}

	if es.voltageL1, err = homeassistant.GetEntityStateValueFloat64("sensor.electricity_meter_voltage_phase_l1"); err != nil {
		return ElectricalState{}, err
	}

	if es.voltageL2, err = homeassistant.GetEntityStateValueFloat64("sensor.electricity_meter_voltage_phase_l2"); err != nil {
		return ElectricalState{}, err
	}

	if es.voltageL3, err = homeassistant.GetEntityStateValueFloat64("sensor.electricity_meter_voltage_phase_l3"); err != nil {
		return ElectricalState{}, err
	}

	return es, nil
}

func sendElectricalData(es ElectricalState) {
	p1 := influxdb2.NewPoint(
		"energy_meter",
		map[string]string{"phase": "1"},
		map[string]interface{}{
			"power":   es.powerConsumptionL1,
			"voltage": es.voltageL1,
		},
		es.timestamp)
	influxWriter.WritePoint(p1)

	p2 := influxdb2.NewPoint(
		"energy_meter",
		map[string]string{"phase": "2"},
		map[string]interface{}{
			"power":   es.powerConsumptionL2,
			"voltage": es.voltageL2,
		},
		es.timestamp)
	influxWriter.WritePoint(p2)

	p3 := influxdb2.NewPoint(
		"energy_meter",
		map[string]string{"phase": "3"},
		map[string]interface{}{
			"power":   es.powerConsumptionL3,
			"voltage": es.voltageL3,
		},
		es.timestamp)
	influxWriter.WritePoint(p3)

	p4 := influxdb2.NewPoint(
		"energy_consumed",
		map[string]string{},
		map[string]interface{}{
			"day":   es.energyConsumedDay,
			"night": es.energyConsumedNight,
		},
		es.timestamp)
	influxWriter.WritePoint(p4)

	influxWriter.Flush()
}
