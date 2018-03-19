package main

import (
	"context"
	"os"
	"time"
)

func main() {

	// First create a logger.
	log, err := New("test", 1, os.Stdout)

	// Check if we're allowed to stdout.
	if err != nil {
		panic(err) // Check for error
	}

	// Create a context this allows to shutdown gracefully.
	ctx := CreateApplicationContext(log)

	// Log we're starting.
	log.Info("Monitor started")

	// Create the channels
	monitorChannel := CreateMonitorChannel()

	// Defer closing of the channels
	defer monitorChannel.CloseChannels()

	MonitorRimoteConnectionStatus(ctx, log, monitorChannel.RimoteMessageChannel)
	NewEthernetMonitor(ctx, monitorChannel.EthernetMessageChannel)

	// Run our message loop blocking ...
	messageloop(ctx, log, monitorChannel)

	log.Info("Monitor is going to shutdown ...")
}

func logGpioOnError(logger *Logger, err error) {

	if err != nil {
		logger.WarningF("%v", err)
	}
}

func messageloop(ctx context.Context, logger *Logger, monitorChannel MonitorChannel) {

	timeout := 1000 * time.Millisecond
	msg := NewMessage()

	msg.GeneralStatus().SetHardwareStatus(true)
	msg.GeneralStatus().SetSoftwareStatus(true)

	for {

		select {

		case <-ctx.Done():
			return
		case rimoteMessage := <-monitorChannel.RimoteMessageChannel:
			msg.RimoteStatus().SetRimoteGUIDPresent(rimoteMessage.HasHardwareID)
			msg.RimoteStatus().SetRimoteConnected(rimoteMessage.IsConnected)

			err := SetRimoteLed(rimoteMessage.IsConnected)
			logGpioOnError(logger, err)
		case ethernetmessage := <-monitorChannel.EthernetMessageChannel:
			msg.ConnectionStatus().SetEth0Status(ethernetmessage.Eth0.Connected)
			msg.ConnectionStatus().SetEth1Status(ethernetmessage.Eth1.Connected)
			msg.ConnectionStatus().SetEthernetConfigurationStatus(ethernetmessage.EthernetConfigured)
			msg.ConnectionStatus().SetWifiEnabled(ethernetmessage.Wifi0.Connected)

			err := SetWifiLed(ethernetmessage.Wifi0.Configured, ethernetmessage.Wifi0.Connected)
			logGpioOnError(logger, err)

			err = SetEth0Led(ethernetmessage.Eth0.Configured, ethernetmessage.Eth0.Connected)
			logGpioOnError(logger, err)

			err = SetEth1Led(ethernetmessage.Eth1.Configured, ethernetmessage.Eth1.Connected)
			logGpioOnError(logger, err)

		default:
			time.Sleep(timeout)

			err := SendMessage("127.0.0.1:9876", msg.Data)
			if err != nil {
				logger.Errorf("could not send status message: %v", err)
			}
		}
	}
}
