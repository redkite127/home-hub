package hue

import "encoding/json"

type ApiErrors []struct {
	Description string `json:"description"`
}

type ApiReply struct {
	Errors *ApiErrors      `json:"errors"`
	Data   json.RawMessage `json:"data"`
}

type Bridge struct {
	ID    string `json:"id"`
	IDv1  string `json:"id_v1"`
	Owner struct {
		RessourceID   string `json:"rid"`
		RessourceType string `json:"rtype"`
	} `json:"owner"`
	BridgeID string `json:"bridge_id"`
	TimeZone struct {
		TimeZone string `json:"time_zone"`
	} `json:"time_zone"`
	Type string `json:"type"`
}

type Device struct {
	ID       string `json:"id"`
	IDv1     string `json:"id_v1"`
	Metadata struct {
		Archetype string `json:"archetype"`
		Name      string `json:"name"`
	} `json:"metadata"`
	ProductData struct {
		Certified        bool   `json:"certified"`
		ManufacturerName string `json:"manufacturer_name"`
		ModelID          string `json:"model_id"`
		ProductArchetype string `json:"product_archetype"`
		ProductName      string `json:"product_name"`
		SoftwareVersion  string `json:"software_version"`
	} `json:"product_data"`
	Services []struct {
		RessourceID   string `json:"rid"`
		RessourceType string `json:"rtype"`
	} `json:"services,omitempty"`
	Type string `json:"type"`
}

type TemperatureService struct {
	ID    string `json:"id"`
	IDv1  string `json:"id_v1"`
	Owner struct {
		RessourceID   string `json:"rid"`
		RessourceType string `json:"rtype"`
	} `json:"owner"`
	Enabled     bool `json:"enabled"`
	Temperature struct {
		Temperature      float32 `json:"temperature"`
		TemperatureValid bool    `json:"temperature_valid"`
	} `json:"temperature"`
	Type string `json:"type"`
}

type BatteryService struct {
	ID    string `json:"id"`
	IDv1  string `json:"id_v1"`
	Owner struct {
		RessourceID   string `json:"rid"`
		RessourceType string `json:"rtype"`
	} `json:"owner"`
	PowerState struct {
		BatteryState string `json:"battery_state"`
		BatteryLevel int    `json:"battery_level"`
	} `json:"power_state"`
	Type string `json:"type"`
}
