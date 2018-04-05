package main

import (
	"context"
	"time"
)

const deviceModemPort = "/dev/ttyUSB3"
const debugModemPort = "COM10"

// ModemStatusMessage structure
type ModemStatusMessage struct {
	ModemAvailable   bool
	DataAvailable    bool
	SimUccid         string
	SimpinOk         bool
	SimCardAvailable bool
	SignalStrength   SignalStrength
}

// WatchModem func
func WatchModem(ctx context.Context, logger *Logger, modemStatusMessageChannel chan ModemStatusMessage) {

	timeout := 30 * time.Second

	// Run the watcher in a new go routine.
	go func() {
		for {
			select {
			// Check if we're closed
			case <-ctx.Done():
				return
			// Handle modem logic
			default:
				// Modem handling
				err := handleModem(ctx, logger, modemStatusMessageChannel)

				if err != nil {
					logger.Errorf("Modem error: %v", err)
					modemStatusMessageChannel <- ModemStatusMessage{ModemAvailable: false}

					if IsDebugMode() {
						logger.Debugf("Waiting: %v before retrying to connect", timeout)
					}
					// Sleep to prevent a mad reconnect loop.
					time.Sleep(timeout)
				} else {
					// We have no errors so, gracefull shutdown ...
					return
				}
			}
		}
	}()
}

func handleModem(ctx context.Context, logger *Logger, modemStatusMessageChannel chan ModemStatusMessage) error {

	// Get the modem string
	modemport := getModemPort()
	commandTimeout := 5 * time.Second

	// Build the config
	config := &Config{
		Name: modemport,
		Baud: 115200,
	}

	port, err := OpenPort(config)

	if err != nil {
		return err
	}

	// Some debug logging
	if IsDebugMode() {
		logger.Debugf("Succesfully opened modem port: %v with baudrate: %v and default timeout of: %v", config.Name, config.Baud, commandTimeout)
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

	err = handleModemData(ctx, port, commandTimeout, logger, modemStatusMessageChannel)
	return err
}

func handleModemData(ctx context.Context, port *Port, commandTimeout time.Duration, logger *Logger, modemStatusMessageChannel chan ModemStatusMessage) error {

	// Report we have modem
	modemStatusMessageChannel <- ModemStatusMessage{ModemAvailable: true}

	for {
		select {
		// Check if we're closed
		case <-ctx.Done():
			return nil

		case <-time.After(commandTimeout):
			err := handleAT(ctx, port, commandTimeout, logger, modemStatusMessageChannel)
			if err != nil {
				return err
			}

		}
	}
}

const atSeperator = '\r'

// RunModemCommands runs one or more modem commands with retry support
func RunModemCommands(commands []func() error, retryCount int) error {

	for _, command := range commands {
		for i := 0; i < retryCount; i++ {

			err := command()
			if err != nil && i == retryCount {
				return err
			}

			if err == nil {
				break
			}
		}
	}
	return nil
}

func getModemPort() string {

	if IsTargetDevice() {
		return deviceModemPort
	}

	return debugModemPort
}
