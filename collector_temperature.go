package main

import "fmt"

func collectAndSendTemperatureData() error {
	ts, err := collectTemperatureData()
	if err != nil {
		return fmt.Errorf("failed to collect electrical data: %w", err)
	}
	//log.Println(es)  // TODO log it only in DEBUG mode
	sendTemperatureData(ts)

	return nil
}

func collectTemperatureData() (string, error) {
	return "", nil
}

func sendTemperatureData(string) {

}
