package main

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
