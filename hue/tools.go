package hue

import (
	"encoding/json"
	"net/http"
)

func doRequest(req *http.Request) (data json.RawMessage, errs *ApiErrors, resp *http.Response, err error) {
	req.Header.Add("hue-application-key", config.Token)

	resp, err = client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var reply ApiReply
	if err = json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return
	}

	return reply.Data, reply.Errors, resp, nil
}
