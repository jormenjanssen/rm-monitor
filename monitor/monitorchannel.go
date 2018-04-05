package main

// MonitorChannel structure
type MonitorChannel struct {
	EthernetMessageChannel    chan EthernetMessage
	RimoteMessageChannel      chan RimoteMessage
	ModemStatusMessageChannel chan ModemStatusMessage
	InfoMessageChannel        chan HostInfo
}

// CreateMonitorChannel creates a monitor channel
func CreateMonitorChannel() MonitorChannel {

	return MonitorChannel{
		EthernetMessageChannel:    make(chan EthernetMessage),
		RimoteMessageChannel:      make(chan RimoteMessage),
		ModemStatusMessageChannel: make(chan ModemStatusMessage),
		InfoMessageChannel:        make(chan HostInfo)}
}

// CloseChannels closes all channels
func (monitorChannel *MonitorChannel) CloseChannels() {
	close(monitorChannel.EthernetMessageChannel)
	close(monitorChannel.RimoteMessageChannel)
	close(monitorChannel.ModemStatusMessageChannel)
	close(monitorChannel.InfoMessageChannel)
}
