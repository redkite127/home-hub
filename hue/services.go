package hue

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var unexpectedTemperatureServices map[string]bool = map[string]bool{}
var unexpectedBatteryServices map[string]bool = map[string]bool{}

func GetTemperatures() (map[string]float32, error) {
	url, err := url.JoinPath(config.URL, "clip", "v2", "resource", "temperature")
	if err != nil {
		return map[string]float32{}, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return map[string]float32{}, err
	}

	data, _, resp, err := doRequest(req)
	if err != nil {
		return map[string]float32{}, fmt.Errorf("failed to retrieve temperatures from HUE API: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return map[string]float32{}, fmt.Errorf("unexpected status code received from HUE API: %d", resp.StatusCode)
	}

	var ts []TemperatureService
	if err = json.Unmarshal(data, &ts); err != nil {
		return map[string]float32{}, err
	}

	temps := map[string]float32{}
	for i := range ts {
		if !ts[i].Temperature.TemperatureValid {
			continue
		}

		deviceName, ok := config.Devices[ts[i].Owner.RessourceID]
		if !ok {
			if !unexpectedTemperatureServices[ts[i].ID] {
				fmt.Println("NOTICE - unexpected temperature service found")
				fmt.Println("       - id:", ts[i].ID)
				fmt.Println("       - owner.id:", ts[i].Owner.RessourceID)
				unexpectedTemperatureServices[ts[i].ID] = true
			}
			continue
		}

		temps[deviceName] = ts[i].Temperature.Temperature
	}

	return temps, nil
}

func GetBatteries() (map[string]float32, error) {
	url, err := url.JoinPath(config.URL, "clip", "v2", "resource", "device_power")
	if err != nil {
		return map[string]float32{}, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return map[string]float32{}, err
	}

	data, _, resp, err := doRequest(req)
	if err != nil {
		return map[string]float32{}, fmt.Errorf("failed to retrieve batteries from HUE API: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return map[string]float32{}, fmt.Errorf("unexpected status code received from HUE API: %d", resp.StatusCode)
	}

	var bs []BatteryService
	if err = json.Unmarshal(data, &bs); err != nil {
		return map[string]float32{}, err
	}

	batteries := map[string]float32{}
	for i := range bs {
		deviceName, ok := config.Devices[bs[i].Owner.RessourceID]
		if !ok {
			if !unexpectedBatteryServices[bs[i].ID] {
				fmt.Println("NOTICE - unexpected battery service found")
				fmt.Println("       - id:", bs[i].ID)
				fmt.Println("       - owner.id:", bs[i].Owner.RessourceID)
				unexpectedBatteryServices[bs[i].ID] = true
			}
			continue
		}

		batteries[deviceName] = float32(bs[i].PowerState.BatteryLevel)
	}

	return batteries, nil
}
