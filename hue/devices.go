package hue

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var devices []Device

func InitDevices() error {
	var err error
	devices, err = GetDevices()
	if err != nil {
		return fmt.Errorf("failed to initialize HUE devices: %w", err)
	}

	return nil
}

func GetDevices() ([]Device, error) {
	url, err := url.JoinPath(config.URL, "clip", "v2", "resource", "device")
	if err != nil {
		return []Device{}, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []Device{}, err
	}

	data, _, resp, err := doRequest(req)
	if err != nil {
		return []Device{}, fmt.Errorf("failed to retrieve devices from HUE API: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return []Device{}, fmt.Errorf("unexpected status code received from HUE API: %d", resp.StatusCode)
	}

	var devices []Device
	if err = json.Unmarshal(data, &devices); err != nil {
		return []Device{}, err
	}

	return devices, nil
}

func GetDevice(id string) (d Device, err error) {
	url, err := url.JoinPath(config.URL, "clip", "v2", "resource", "device", id)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	data, _, resp, err := doRequest(req)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return Device{}, fmt.Errorf("HUE device not found")
		}

		return Device{}, fmt.Errorf("unexpected status code received from HUE API: %d", resp.StatusCode)
	}

	var devices []Device
	if err = json.Unmarshal(data, &devices); err != nil {
		return
	}

	if len(devices) != 1 {
		return Device{}, fmt.Errorf("unexpected amount of HUE devices returned: %d", len(devices))
	}

	return devices[0], nil
}
