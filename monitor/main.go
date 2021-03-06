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
		panic(err)
	}

	/*go func() {
		http.ListenAndServe("localhost:6060", nil)
		log.Debug("Start pprof @ localhost:6060")
	}()

	*/

	// Create a context this allows to shutdown gracefully.
	ctx := CreateApplicationContext(log)

	// Log we're starting.
	if IsDebugMode() {
		log.Debug("Monitor started [Debug]")
	} else {
		log.Info("Monitor started")
	}

	// Create the channels
	monitorChannel := CreateMonitorChannel()

	// Defer closing of the channels
	defer monitorChannel.CloseChannels()

	// Configure all our watches
	MonitorRimoteConnectionStatus(ctx, log, monitorChannel.RimoteMessageChannel)
	NewEthernetMonitor(ctx, monitorChannel.EthernetMessageChannel)
	WatchModem(ctx, log, monitorChannel.ModemStatusMessageChannel)
	HandleHostInfo(ctx, log, monitorChannel.InfoMessageChannel)

	// Run our message loop blocking ...
	messageloop(ctx, log, monitorChannel)

	log.Info("Monitor is going to shutdown in 10 seconds ...")
	time.Sleep(10 * time.Second)
}

func executeWithLogger(logger *Logger, context string, fn func() error) {
	err := fn()
	if err != nil {
		logger.WarningF("%v %v", context, err)
	}
}

func initDefaults(msg *Message) {

	msg.GeneralStatus().SetHardwareStatus(true)
	msg.GeneralStatus().SetSoftwareStatus(true)
	msg.GeneralStatus().SetVccStatus(true)
	msg.HardwareStatus().SetNandStatus(true)
}

func messageloop(ctx context.Context, logger *Logger, monitorChannel MonitorChannel) {

	timeout := 2000 * time.Millisecond
	msg := NewMessage()

	// Set some defaults
	initDefaults(msg)

	// Try to detect the firmware version at startup.
	firmwareVersion, err := GetFirmwareVersion()

	if err != nil {
		logger.WarningF("Cannot detect firmware version")
	}

	if IsDebugMode() {
		logger.DebugF("Got the following firmware version information: %v", firmwareVersion)
	}

	for {

		select {

		case <-ctx.Done():
			return
		case rimoteMessage := <-monitorChannel.RimoteMessageChannel:
			msg.RimoteStatus().SetRimoteGUIDPresent(rimoteMessage.HasHardwareID)
			msg.RimoteStatus().SetRimoteConnected(rimoteMessage.IsConnected)
			// todo: Fix this in the feature
			msg.RimoteStatus().SetRimoteSSLOk(true)
			msg.RimoteStatus().SetRimoteConfOk(true)

			executeWithLogger(logger, "led:rimote", func() error {
				return SetRimoteLed(rimoteMessage.IsConnected)
			})
		case ethernetMessage := <-monitorChannel.EthernetMessageChannel:
			msg.ConnectionStatus().SetEth0Status(ethernetMessage.Eth0.Connected)
			msg.ConnectionStatus().SetEth1Status(ethernetMessage.Eth1.Connected)
			msg.ConnectionStatus().SetEthernetConfigurationStatus(ethernetMessage.EthernetConfigured)
			msg.ConnectionStatus().SetWifiEnabled(ethernetMessage.Wifi0.Connected)
			if ethernetMessage.Wifi0.Connected {
				msg.ConnectionStatus().SetWifiSignal(FairSignal)
			} else {
				msg.ConnectionStatus().SetWifiSignal(NoSignal)
			}
			setConnectionLeds(logger, ethernetMessage)
		case modemMessage := <-monitorChannel.ModemStatusMessageChannel:
			msg.ConnectionStatus().SetMobileInternetEnabled(modemMessage.ModemAvailable)
			msg.ConnectionStatus().SetSimPinOK(modemMessage.SimpinOk)
			msg.ConnectionStatus().SetModemSignal(modemMessage.SignalStrength)
			msg.ConnectionStatus().SetBroadbandConnectionType(modemMessage.BroadbandConnType)

			setModemLed(logger, modemMessage)

			// Report status back
			monitorChannel.InfoMessageChannel <- HostInfo{
				ModemEnabled: modemMessage.ModemAvailable,
				SimID:        modemMessage.SimUccid,
			}

		default:
			time.Sleep(timeout)
			err := SendMessage(logger, msg.Data)

			if err != nil {
				logger.Errorf("could not send status message: %v", err)
			}
		}
	}
}

func setModemLed(logger *Logger, modemStatusMessage ModemStatusMessage) {
	executeWithLogger(logger, "led:broadband", func() error {

		if modemStatusMessage.ConfigAvailable && !modemStatusMessage.ModemAvailable {
			return SetBroadbandLed(ErrorSignal)
		}

		if modemStatusMessage.DataAvailable {
			return SetBroadbandLed(GoodSignal)
		}

		if modemStatusMessage.ModemAvailable {
			return SetBroadbandLed(NoSignal)
		}

		return SetBroadbandLed(NoSignal)
	})
}

func setConnectionLeds(logger *Logger, ethernetmessage EthernetMessage) {

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
}
