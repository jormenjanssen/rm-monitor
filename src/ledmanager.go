package main

import (
	"errors"
)

// LedState Type
type LedState uint

const (
	// Off represents nothing
	Off LedState = 0
	// RedColor color
	RedColor LedState = 1
	// OrangeColor color
	OrangeColor LedState = 2
	// GreenColor color
	GreenColor LedState = 3
	// BlueColor color
	BlueColor LedState = 4
)

var gpioMapping map[ManagerGpio]Pin
var errGpioNotInitialized = errors.New("gpio not inialized")

// ManagerGpio type
type ManagerGpio uint

const (
	// LedPowerRed led power led
	LedPowerRed ManagerGpio = 135
	// LedPowerGreen led power green
	LedPowerGreen ManagerGpio = 112
	// LedPowerBlue led power blue
	LedPowerBlue ManagerGpio = 120
)

// SystemLed type
type SystemLed uint

const (
	// Power led
	Power SystemLed = 0
	// Ups led
	Ups SystemLed = 1
	// Eth0 led
	Eth0 SystemLed = 2
	// Eth1 led
	Eth1 SystemLed = 3
	// Gps led
	Gps SystemLed = 4
	// Broadband led
	Broadband SystemLed = 5
	// Wifi led
	Wifi SystemLed = 5
)

// SignalStrength type
type SignalStrength uint

const (
	// Weak low or weak signal
	Weak SignalStrength = 1
	// Fair signal
	Fair SignalStrength = 2
	// Good signal strength
	Good SignalStrength = 3
)

// SetRimoteLed sets the rimote led
func SetRimoteLed(connected bool) error {

	return gpioFunc(LedPowerBlue, func(pin Pin) error {

		if connected {
			return pin.High()
		}

		return pin.Low()
	})
}

// SetWifiLed sets the wifi led
func SetWifiLed(configured bool, connected bool) error {

	return nil
}

func gpioFunc(gpio ManagerGpio, fn func(Pin) error) error {

	// First look if we're setup.
	err := setup()
	if err != nil {
		return err
	}

	// Try to get our pin value.
	pin, err := GetPin(gpio)

	if err != nil {
		return err
	}

	// Execute our closure.
	return fn(pin)
}

func setup() error {

	if !IsTargetDevice() {
		return errGpioNotInitialized
	}

	if gpioMapping != nil {
		return nil
	}

	gpioMapping = map[ManagerGpio]Pin{
		LedPowerBlue: NewOutput(uint(LedPowerRed), false),
	}

	NewOutput(uint(LedPowerBlue), false)

	return errGpioNotInitialized
}

// GetPin returns the pin if exported
func GetPin(managerGpio ManagerGpio) (Pin, error) {

	if gpioMapping == nil {
		return Pin{}, errors.New("gpio mapping not initialized")
	}

	elem, ok := gpioMapping[managerGpio]

	if !ok {
		return elem, errors.New("gpio does not exsist")
	}

	return elem, nil
}
