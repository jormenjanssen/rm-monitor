package main

// MonitorChannel structure
type MonitorChannel struct {
	EthernetMessageChannel chan EthernetMessage
	RimoteMessageChannel   chan RimoteMessage
}

// CreateMonitorChannel creates a monitor channel
func CreateMonitorChannel() MonitorChannel {

	return MonitorChannel{
		EthernetMessageChannel: make(chan EthernetMessage),
		RimoteMessageChannel:   make(chan RimoteMessage)}
}

// CloseChannels closes all channels
func (monitorChannel *MonitorChannel) CloseChannels() {
	close(monitorChannel.EthernetMessageChannel)
	close(monitorChannel.RimoteMessageChannel)
}
