package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func monitorContext(logger *Logger, cancel func(), signalChannel chan os.Signal) {
	s := <-signalChannel

	if IsDebugMode() {
		logger.DebugF("Got signal: %v invoking cancellation of context", s)
	}

	cancel()
}

// CreateApplicationContext creates context which respects application shutdown
func CreateApplicationContext(logger *Logger) context.Context {

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	signalChannel := make(chan os.Signal, 1)

	// Passing no signals to Notify means that
	// all signals will be sent to the channel.
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	if IsDebugMode() {
		logger.Debugf("Signal handler installed listening for SIGINT | SIGTERM")
	}

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	go monitorContext(logger, cancel, signalChannel)

	return ctx
}
