package homeassistant

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

func GetEntityState(id string) (EntityState, error) {
	url, err := url.JoinPath(config.URL, "api", "states", id)
	if err != nil {
		return EntityState{}, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return EntityState{}, err
	}

	req.Header.Add("Authorization", "Bearer "+config.Token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return EntityState{}, err
	}
	defer resp.Body.Close()

	var es EntityState
	if err := json.NewDecoder(resp.Body).Decode(&es); err != nil {
		return EntityState{}, err
	}

	return es, nil
}

func GetEntityStateValueFloat64(id string) (float64, error) {
	es, err := GetEntityState(id)
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseFloat(es.State, 64)
	if err != nil {
		return 0, err
	}

	return v, nil
}
