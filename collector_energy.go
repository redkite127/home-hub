package main

import (
	"log"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/redkite127/home-hub/homeassistant"
)

type ElectricalState struct {
	energyConsumedDay   *float64
	energyConsumedNight *float64

	powerConsumptionL1 *float64
	powerConsumptionL2 *float64
	powerConsumptionL3 *float64

	voltageL1 *float64
	voltageL2 *float64
	voltageL3 *float64

	currentL1 *float64
	currentL2 *float64
	currentL3 *float64

	timestamp time.Time
}

func collectAndSendElectricalData() error {
	sendElectricalData(collectElectricalData())

	return nil
}

// electricalReader reads entities without failing the cycle: a single sensor
// that is unavailable or non-numeric only yields nil. DSMR entities do go
// unavailable when the P1 reader misses a message, and at a 1m collection
// frequency logging each one separately floods the log during an outage, so
// failures are accumulated and reported once per cycle instead.
type electricalReader struct {
	failed []string
}

func (r *electricalReader) read(entity string) *float64 {
	v, err := homeassistant.GetEntityStateValueFloat64(entity)
	if err != nil {
		r.failed = append(r.failed, entity)
		return nil
	}

	return &v
}

func collectElectricalData() ElectricalState {
	r := &electricalReader{}

	es := ElectricalState{
		energyConsumedDay:   r.read("sensor.electricity_meter_energy_consumption_tarif_1"),
		energyConsumedNight: r.read("sensor.electricity_meter_energy_consumption_tarif_2"),

		powerConsumptionL1: r.read("sensor.electricity_meter_power_consumption_phase_l1"),
		powerConsumptionL2: r.read("sensor.electricity_meter_power_consumption_phase_l2"),
		powerConsumptionL3: r.read("sensor.electricity_meter_power_consumption_phase_l3"),

		voltageL1: r.read("sensor.electricity_meter_voltage_phase_l1"),
		voltageL2: r.read("sensor.electricity_meter_voltage_phase_l2"),
		voltageL3: r.read("sensor.electricity_meter_voltage_phase_l3"),

		currentL1: r.read("sensor.electricity_meter_current_phase_l1"),
		currentL2: r.read("sensor.electricity_meter_current_phase_l2"),
		currentL3: r.read("sensor.electricity_meter_current_phase_l3"),

		timestamp: time.Now().UTC(),
	}

	if len(r.failed) > 0 {
		log.Printf("failed to read %d electrical entities: %s", len(r.failed), strings.Join(r.failed, ", "))
	}

	return es
}

func sendElectricalData(es ElectricalState) {
	phases := []struct {
		phase   string
		power   *float64
		voltage *float64
		current *float64
	}{
		{"1", es.powerConsumptionL1, es.voltageL1, es.currentL1},
		{"2", es.powerConsumptionL2, es.voltageL2, es.currentL2},
		{"3", es.powerConsumptionL3, es.voltageL3, es.currentL3},
	}

	for _, ph := range phases {
		p := influxdb2.NewPointWithMeasurement("energy_meter")
		p.AddTag("phase", ph.phase)
		addFieldRounded(p, "power", ph.power, 3)     // kW
		addFieldRounded(p, "voltage", ph.voltage, 1) // V
		addFieldRounded(p, "current", ph.current, 2) // A
		if len(p.FieldList()) == 0 {
			continue // InfluxDB rejects points without fields
		}
		p.SetTime(es.timestamp)
		writePoint(p)
	}

	{
		p := influxdb2.NewPointWithMeasurement("energy_consumed")
		addFieldRounded(p, "day", es.energyConsumedDay, 3)     // kWh
		addFieldRounded(p, "night", es.energyConsumedNight, 3) // kWh
		if len(p.FieldList()) > 0 {
			p.SetTime(es.timestamp)
			writePoint(p)
		}
	}

	flushPoints()
}
