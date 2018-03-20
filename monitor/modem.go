package main

import (
	"context"
	"time"
)

const deviceModemPort = "/dev/ttyUSB3"
const debugModemPort = "COM3"

// ModemStatusMessage structure
type ModemStatusMessage struct {
	ModemAvailable   bool
	SimUccid         string
	SimpinOk         bool
	SimCardAvailable bool
	SignalStrength   SignalStrength
}

// WatchModem func
func WatchModem(ctx context.Context, logger *Logger, modemStatusMessageChannel chan ModemStatusMessage) {

	// Get the modem string
	modemport := getModemPort()

	// Build the config
	config := &Config{
		Name:        modemport,
		Baud:        115200,
		ReadTimeout: 5 * time.Second}

	port, err := OpenPort(config)

	if err != nil {
		logger.WarningF("Could not open modem reason: %v", err)
	}

	if err == nil {

		// Some debug logging
		if IsDebugMode() {
			logger.Debugf("Succesfully opened modem port: %v with baudrate: %v and default timeout of: %v", config.Name, config.Baud, config.ReadTimeout)
		}

		// Cleanup code
		defer func() {
			err := port.Close()
			if err != nil {
				logger.Warningf("Could not close serial port reason: %v", err)
			} else if IsDebugMode() {
				logger.Debugf("Succesfully closed modem port: %v", config.Name)
			}

		}()
	}
}

func getModemPort() string {

	if IsTargetDevice() {
		return deviceModemPort
	}

	return debugModemPort
}
