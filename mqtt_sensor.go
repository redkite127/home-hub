package main

import (
	"fmt"

	"github.com/redkite1/home-hub/mqtt"
	log "github.com/sirupsen/logrus"
)

func MQTTregisterRoomSensor(room string, humidity, battery bool) {
	// send the config message if it's the first time it sees this room sensor
	if _, ok := sensors[room]; !ok {
		topic := fmt.Sprintf("homeassistant/sensor/mqtt_%s_temperature/config", room)
		payload := fmt.Sprintf(`{"device_class": "temperature", "name": "mqtt_%s_temperature", "state_topic": "homeassistant/sensor/mqtt_%s/state", "unit_of_measurement": "°C", "value_template": "{{ value_json.temperature}}" }`, room, room)
		if err := mqtt.Publish(topic, false, payload); err != nil {
			log.WithError(err).Errorf("failed to publish sensor config for room '%v'", room)
		}

		if humidity {
			topic := fmt.Sprintf("homeassistant/sensor/mqtt_%s_humidity/config", room)
			payload = fmt.Sprintf(`{"device_class": "humidity", "name": "mqtt_%s_humidity", "state_topic": "homeassistant/sensor/mqtt_%s/state", "unit_of_measurement": "%%", "value_template": "{{ value_json.humidity}}" }`, room, room)
			mqtt.Publish(topic, false, payload)
		}
	}
}

func MQTTsendRoomSensor(room string, sr SensorRecord) {
	topic := fmt.Sprintf("homeassistant/sensor/mqtt_%s/state", room)
	if err := mqtt.Publish(topic, true, sr); err != nil {
		log.WithError(err).Errorf("failed to publish sensor record for room '%v'", room)
	}
}
