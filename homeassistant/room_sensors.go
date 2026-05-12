package homeassistant

import (
	"log"
)

type RoomSensorReading struct {
	Temperature *float32
	Humidity    *float32
	Battery     *float32
}

func GetRoomSensors() (map[string]RoomSensorReading, error) {
	readings := map[string]RoomSensorReading{}

	for room, entities := range config.RoomSensors {
		r := RoomSensorReading{}

		if entities.Temperature != "" {
			v, err := GetEntityStateValueFloat64(entities.Temperature)
			if err != nil {
				log.Printf("homeassistant: failed to read temperature for %s (%s): %s", room, entities.Temperature, err)
			} else {
				f := float32(v)
				r.Temperature = &f
			}
		}

		if entities.Humidity != "" {
			v, err := GetEntityStateValueFloat64(entities.Humidity)
			if err != nil {
				log.Printf("homeassistant: failed to read humidity for %s (%s): %s", room, entities.Humidity, err)
			} else {
				f := float32(v)
				r.Humidity = &f
			}
		}

		if entities.Battery != "" {
			v, err := GetEntityStateValueFloat64(entities.Battery)
			if err != nil {
				log.Printf("homeassistant: failed to read battery for %s (%s): %s", room, entities.Battery, err)
			} else {
				f := float32(v)
				r.Battery = &f
			}
		}

		readings[room] = r
	}

	return readings, nil
}
