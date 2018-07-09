package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Monitor type
type rimoteMonitor struct {
}

// RimoteMessage structure
type RimoteMessage struct {
	IsConnected   bool
	HasHardwareID bool
}

// RimoteAPIMesssage structure
type RimoteAPIMesssage struct {
	IsConnected bool
	HardwareID  string
}

// MonitorRimoteConnectionStatus handles monitoring of the connection status
func MonitorRimoteConnectionStatus(ctx context.Context, logger *Logger, rimoteMessageChannel chan RimoteMessage) {

	monitor := new(rimoteMonitor)

	// Run our logic inside a go routine.
	go monitor.run(ctx, logger, rimoteMessageChannel)
}

func (*rimoteMonitor) run(ctx context.Context, logger *Logger, rimoteMessageChannel chan RimoteMessage) {

	apiEndpoint := "http://localhost:9000/api/rimote/info"
	defaultTimeout := 30 * time.Second
	rimoteAPIClient := &http.Client{Timeout: defaultTimeout}

	// Allocate message for the api once.
	apiMessage := new(RimoteAPIMesssage)

	for {

		// Get JSON data from the api.
		err := getJSON(rimoteAPIClient, apiEndpoint, apiMessage)

		// log any errors occured
		if err != nil {
			logger.Warningf("Got error while trying to fetch rimote status: %v", err)

			// Send empty response to channel
			rimoteMessageChannel <- RimoteMessage{IsConnected: false, HasHardwareID: false}
		}

		// Only post to our channel when we don't have any api errors.
		if err == nil {
			// Report our status
			rimoteMessageChannel <- RimoteMessage{IsConnected: apiMessage.IsConnected, HasHardwareID: true}
		}

		// Wait a while before retrying ...
		time.Sleep(defaultTimeout)
	}
}

func getJSON(httpClient *http.Client, url string, target interface{}) error {

	r, err := httpClient.Get(url)

	if err != nil {
		return err
	}

	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
