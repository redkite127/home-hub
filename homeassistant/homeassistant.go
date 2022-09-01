package homeassistant

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

func GetEntityState(id string) (EntityState, error) {
	//TODO get those values from config file
	ha_url := "http://10.28.3.149:8123"
	ha_token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiI2OTI4MmZlNTJiMDA0YmJjOWNmOTM1N2JkZTkzYjRhNyIsImlhdCI6MTY2MDY3OTIxMywiZXhwIjoxOTc2MDM5MjEzfQ.UyD2x8XZqOCsrvdROajq2uz9efvoHMg45mapgZDN9tQ"

	url, err := url.JoinPath(ha_url, "api", "states", id)
	if err != nil {
		return EntityState{}, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return EntityState{}, err
	}

	req.Header.Add("Authorization", "Bearer "+ha_token)
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
