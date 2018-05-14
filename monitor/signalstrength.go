package main

// SignalStrength type
type SignalStrength uint

const (
	// ErrorSignal error or no signal
	ErrorSignal SignalStrength = 0
	// NoSignal no signal
	NoSignal SignalStrength = 1
	// WeakSignal or weak signal
	WeakSignal SignalStrength = 2
	// FairSignal strength
	FairSignal SignalStrength = 3
	// GoodSignal strength
	GoodSignal SignalStrength = 4
)
