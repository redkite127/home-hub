package hue

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func GetBridge() (Bridge, error) {
	url, err := url.JoinPath(config.URL, "clip", "v2", "resource", "bridge")
	if err != nil {
		return Bridge{}, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Bridge{}, err
	}

	data, _, resp, err := doRequest(req)
	if resp.StatusCode != http.StatusOK {
		return Bridge{}, fmt.Errorf("unexpected status code received from HUE API: %d", resp.StatusCode)
	}

	var b []Bridge
	if err = json.Unmarshal(data, &b); err != nil {
		return Bridge{}, err
	}

	if len(b) != 1 {
		return Bridge{}, fmt.Errorf("unexpected amount of HUE bridge returned: %d", len(b))
	}

	return b[0], nil
}
