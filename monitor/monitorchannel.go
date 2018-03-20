package main

// MonitorChannel structure
type MonitorChannel struct {
	EthernetMessageChannel    chan EthernetMessage
	RimoteMessageChannel      chan RimoteMessage
	ModemStatusMessageChannel chan ModemStatusMessage
}

// CreateMonitorChannel creates a monitor channel
func CreateMonitorChannel() MonitorChannel {

	return MonitorChannel{
		EthernetMessageChannel:    make(chan EthernetMessage),
		RimoteMessageChannel:      make(chan RimoteMessage),
		ModemStatusMessageChannel: make(chan ModemStatusMessage)}
}

// CloseChannels closes all channels
func (monitorChannel *MonitorChannel) CloseChannels() {
	close(monitorChannel.EthernetMessageChannel)
	close(monitorChannel.RimoteMessageChannel)
	close(monitorChannel.ModemStatusMessageChannel)
}
