package homeassistant

import "time"

type EntityState struct {
	EntityID   string `json:"entity_id"`
	State      string `json:"state"`
	Attributes struct {
		StateClass        string `json:"state_class"`
		UnitOfMeasurement string `json:"unit_of_measurement"`
		DeviceClass       string `json:"device_class"`
		Icon              string `json:"icon"`
		FriendlyName      string `json:"friendly_name"`
	} `json:"attributes"`
	LastChanged time.Time `json:"last_changed"`
	LastUpdated time.Time `json:"last_updated"`
}
