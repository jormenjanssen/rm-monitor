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

func executeWithLogger(logger *Logger, context string, fn func() error) {

	err := fn()

	if err != nil {
		logger.WarningF("%v %v", context, err)
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

			executeWithLogger(logger, "led:rimote", func() error {
				return SetRimoteLed(rimoteMessage.IsConnected)
			})

		case ethernetmessage := <-monitorChannel.EthernetMessageChannel:
			msg.ConnectionStatus().SetEth0Status(ethernetmessage.Eth0.Connected)
			msg.ConnectionStatus().SetEth1Status(ethernetmessage.Eth1.Connected)
			msg.ConnectionStatus().SetEthernetConfigurationStatus(ethernetmessage.EthernetConfigured)
			msg.ConnectionStatus().SetWifiEnabled(ethernetmessage.Wifi0.Connected)

			executeWithLogger(logger, "led:wifi", func() error {
				if ethernetmessage.Wifi0.Connected {
					return SetWifiLed(GoodSignal)
				}

				return SetWifiLed(NoSignal)
			})

			executeWithLogger(logger, "led:eth0", func() error {
				return SetEth0Led(ethernetmessage.Eth0.Configured, ethernetmessage.Eth0.Connected)
			})

			executeWithLogger(logger, "led:eth1", func() error {
				return SetEth1Led(ethernetmessage.Eth1.Configured, ethernetmessage.Eth1.Connected)
			})
		default:
			time.Sleep(timeout)

			err := SendMessage("127.0.0.1:9876", msg.Data)
			if err != nil {
				logger.Errorf("could not send status message: %v", err)
			}
		}
	}
}
