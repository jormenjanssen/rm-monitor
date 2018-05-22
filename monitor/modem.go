package main

import (
	"context"
	"os"
	"time"
)

const deviceModemPort = "/dev/ttyUSB3"
const debugModemPort = "COM10"

const modemConfigFile = "/etc/wvdial.conf"

// ModemStatusMessage structure
type ModemStatusMessage struct {
	ConfigAvailable  bool
	ModemAvailable   bool
	DataAvailable    bool
	SimUccid         string
	SimpinOk         bool
	SimCardAvailable bool
	SignalStrength   SignalStrength
}

// TranslateModemDBM translates dbm, ber into a rawvalue
func TranslateModemDBM(rawValue int, berValue int) SignalStrength {

	if berValue == 99 {
		return NoSignal
	}

	if berValue > 7 {
		return WeakSignal
	}

	if rawValue == 0 {
		return NoSignal
	}

	if rawValue == 1 {
		return WeakSignal
	}

	// 109 - 53 DBM range
	// 2 - 30 = 28 steps
	// 109 DBM - 2 DBM/step

	// Good/Excellent treshold: -70DBM (19.5s)
	// Weak treshold: -100 (5s)
	// Fair treshold: -86 (11.5s)

	if rawValue > 1 && rawValue < 31 {

		if rawValue <= 5 {
			return WeakSignal
		}

		if rawValue <= 12 {
			return FairSignal
		}

		if rawValue > 12 {
			return GoodSignal
		}
	}

	if rawValue >= 31 && rawValue < 99 {
		return GoodSignal
	}

	return NoSignal
}

// WatchModem func
func WatchModem(ctx context.Context, logger *Logger, modemStatusMessageChannel chan ModemStatusMessage) {

	timeout := 30 * time.Second
	modemConfigAvailable := false

	// Run the watcher in a new go routine.
	go func() {
		for {
			select {
			// Check if we're closed
			case <-ctx.Done():
				return
			// Handle modem logic
			default:

				modemConfigAvailable = CheckModemConfigAvailable()

				if modemConfigAvailable {

					// Run some checks before trying to connect
					preFlightModemCheck(ctx, logger)

					// Modem handling
					err := handleModem(ctx, logger, modemStatusMessageChannel)

					if err != nil {
						logger.Errorf("Modem error: %v", err)
						modemStatusMessageChannel <- ModemStatusMessage{ConfigAvailable: true, ModemAvailable: false}

						if IsDebugMode() {
							logger.Debugf("Waiting: %v before retrying to connect", timeout)
						}
						// Sleep to prevent a mad reconnect loop.
						time.Sleep(timeout)
					} else {
						// We have no errors so, gracefull shutdown ...
						return
					}
				} else {
					// Report we don't have a config of a modem
					modemStatusMessageChannel <- ModemStatusMessage{ConfigAvailable: false, ModemAvailable: CheckModemConfigAvailable()}

					// Sleep a while before retrying.
					time.Sleep(timeout)
				}
			}
		}
	}()
}

// CheckModemConfigAvailable checks if a modem config file is available
func CheckModemConfigAvailable() bool {

	if !IsTargetDevice() {
		return true
	}

	if _, err := os.Stat(modemConfigFile); err == nil {
		return true
	}

	return false
}

func preFlightModemCheck(ctx context.Context, logger *Logger) {

	// Skip when running for testing
	if !IsTargetDevice() {
		logger.DebugF("Skipped pre-flight modem check because we are not running on target device")
		return
	}

	// Defaults for maximum
	// We need atleast 45 (9 attempts * 5s) seconds to prevent reporting the status to early
	maxAttempts := 10
	sleepDuration := 5 * time.Second

	// Try to do some upfront checks to prevent errors from connecting to early
	for i := 0; i < maxAttempts; i++ {

		select {
		case <-ctx.Done():
			return
		default:
			// Return early when we are able to stat the modem.
			_, err := os.Stat(deviceModemPort)
			if err == nil {
				return
			}

			// Sleep a while
			time.Sleep(sleepDuration)
		}

	}

	logger.Warningf("Modem pre-flight check failed after %v attempts", maxAttempts)
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

	initialConnected, err := handleModemData(ctx, port, commandTimeout, logger, modemStatusMessageChannel)

	// Report that we have a modem atleast.
	if initialConnected {
		modemStatusMessageChannel <- ModemStatusMessage{ModemAvailable: true}
	}

	return err
}

func handleModemData(ctx context.Context, port *Port, commandTimeout time.Duration, logger *Logger, modemStatusMessageChannel chan ModemStatusMessage) (bool, error) {

	for {
		select {
		// Check if we're closed
		case <-ctx.Done():
			return false, nil

		case <-time.After(commandTimeout):
			initialConnected, err := handleAT(ctx, port, commandTimeout, logger, modemStatusMessageChannel)
			if err != nil {
				return initialConnected, err
			}

		}
	}
}

const atSeperator = '\r'

func getModemPort() string {

	if IsTargetDevice() {
		return deviceModemPort
	}

	return debugModemPort
}
