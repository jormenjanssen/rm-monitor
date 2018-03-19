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
	// LedWanRed led wan red (eth0)
	LedWanRed ManagerGpio = 83
	// LedLanRed led lan red (eth1)
	LedLanRed ManagerGpio = 88
	// LedWifiRed led wifi red
	LedWifiRed ManagerGpio = 122
	// LedWifiGreen led wifi red
	LedWifiGreen ManagerGpio = 127
	// LedWifiBlue led wifi red
	LedWifiBlue ManagerGpio = 117
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
	// NoSignal error or no signal
	NoSignal SignalStrength = 0
	// WeakSignal or weak signal
	WeakSignal SignalStrength = 1
	// FairSignal strength
	FairSignal SignalStrength = 2
	// GoodSignal strength
	GoodSignal SignalStrength = 3
)

// SetEth0Led set the ethernet led according to the state
func SetEth0Led(configured bool, connected bool) error {
	return setEthernetLed(LedWanRed, configured, connected)
}

// SetEth1Led set the ethernet led according to the state
func SetEth1Led(configured bool, connected bool) error {
	return setEthernetLed(LedLanRed, configured, connected)
}

func setEthernetLed(gpio ManagerGpio, configured bool, connected bool) error {

	return gpioFunc(gpio, func(pin Pin) error {

		// If we are not configured then disable our red led
		if !configured {
			return pin.Low()
		}

		// If we're connected drop the red led
		if connected {
			return pin.Low()
		}

		// If we're configured but don't have a cable detected, the show our red led
		return pin.High()
	})
}

// SetRimoteLed sets the rimote led
func SetRimoteLed(connected bool) error {

	// Set blue led on
	err := gpioFunc(LedPowerBlue, func(pin Pin) error {

		if connected {
			return pin.High()
		}

		return pin.Low()
	})

	if err != nil {
		return err
	}

	// Set led green inverted
	err = gpioFunc(LedPowerGreen, func(pin Pin) error {

		if connected {
			return pin.Low()
		}

		return pin.High()
	})

	return err
}

// SetWifiLed sets the wifi led
func SetWifiLed(strength SignalStrength) error {
	return SignalStrengthToGpio(LedWifiRed, LedWifiGreen, LedWifiBlue, strength)
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

// SignalStrengthToGpio converts signal strength to gpio
func SignalStrengthToGpio(redSignal ManagerGpio, greenSignal ManagerGpio, blueSignal ManagerGpio, strength SignalStrength) error {

	return gpioSignalFunc(redSignal, greenSignal, blueSignal, func(r Pin, g Pin, b Pin) error {

		// Blue is always low.
		b.Low()

		if strength == NoSignal {
			r.Low()
			g.Low()
		}

		if strength == WeakSignal {
			r.High()
			g.Low()
		}

		if strength == FairSignal {
			r.High()
			g.High()
		}

		if strength == GoodSignal {
			r.Low()
			g.High()
		}

		return nil
	})

}

func gpioSignalFunc(gpioLowSignal ManagerGpio, gpioMediumSignal ManagerGpio, gpioHighSignal ManagerGpio, fn func(l Pin, m Pin, h Pin) error) error {

	// First look if we're setup.
	err := setup()
	if err != nil {
		return err
	}

	// Try to get our low pin value.
	pinLow, err := GetPin(gpioLowSignal)

	if err != nil {
		return err
	}

	// Try to get our middle pin value.
	pinMed, err := GetPin(gpioMediumSignal)

	if err != nil {
		return err
	}

	// Try to get our high pin value.
	pinHigh, err := GetPin(gpioHighSignal)

	if err != nil {
		return err
	}

	// Execute our closure.
	return fn(pinLow, pinMed, pinHigh)
}

func setup() error {

	if !IsTargetDevice() {
		return errGpioNotInitialized
	}

	if gpioMapping != nil {
		return nil
	}

	gpioMapping = map[ManagerGpio]Pin{
		LedPowerBlue:  NewOutput(uint(LedPowerBlue), true),
		LedPowerGreen: NewOutput(uint(LedPowerGreen), true),
		LedWanRed:     NewOutput(uint(LedWanRed), true),
		LedLanRed:     NewOutput(uint(LedLanRed), true),
		LedWifiRed:    NewOutput(uint(LedWifiRed), true),
		LedWifiGreen:  NewOutput(uint(LedWifiGreen), true),
		LedWifiBlue:   NewOutput(uint(LedWifiBlue), true),
	}

	return nil
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
