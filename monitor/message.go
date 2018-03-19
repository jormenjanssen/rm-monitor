package main

// Message structure
type Message struct {
	Data [8]byte
}

// ConnectionStatus structure
type ConnectionStatus struct {
	State1 *byte
	State2 *byte
}

// GeneralStatus structure
type GeneralStatus struct {
	State1 *byte
}

// RimoteStatus structure
type RimoteStatus struct {
	State1 *byte
	State2 *byte
}

// GeneralStatus returns the general status message
func (message *Message) GeneralStatus() *GeneralStatus {
	return &GeneralStatus{State1: &message.Data[0]}
}

// ConnectionStatus returns the general status message
func (message *Message) ConnectionStatus() *ConnectionStatus {
	return &ConnectionStatus{State1: &message.Data[2], State2: &message.Data[3]}
}

// RimoteStatus returns the rimote status message
func (message *Message) RimoteStatus() *RimoteStatus {
	return &RimoteStatus{State1: &message.Data[4], State2: &message.Data[5]}
}

// SetHardwareStatus sets the hardware status
func (generalStatus *GeneralStatus) SetHardwareStatus(value bool) {
	setBit(generalStatus.State1, 0, value)
}

// GetHardwareStatus get the hardware status
func (generalStatus *GeneralStatus) GetHardwareStatus() bool {
	return getBit(generalStatus.State1, 0)
}

// SetSoftwareStatus sets the software status
func (generalStatus *GeneralStatus) SetSoftwareStatus(value bool) {
	setBit(generalStatus.State1, 1, value)
}

// GetSoftwareStatus get the software status
func (generalStatus *GeneralStatus) GetSoftwareStatus() bool {
	return getBit(generalStatus.State1, 1)
}

// GetEth0Status return the connection status of eth1
func (connectionStatus *ConnectionStatus) GetEth0Status() bool {
	return getBit(connectionStatus.State2, 6)
}

// GetEth1Status return the connection status of eth1
func (connectionStatus *ConnectionStatus) GetEth1Status() bool {
	return getBit(connectionStatus.State2, 7)
}

// SetEth0Status sets the connection status of eth0
func (connectionStatus *ConnectionStatus) SetEth0Status(value bool) {
	setBit(connectionStatus.State2, 6, value)
}

// SetEth1Status sets the connection status of eth0
func (connectionStatus *ConnectionStatus) SetEth1Status(value bool) {
	setBit(connectionStatus.State2, 7, value)
}

// GetEthernetConfigurationStatus return the ethernet connection status
func (connectionStatus *ConnectionStatus) GetEthernetConfigurationStatus() bool {
	return getBit(connectionStatus.State2, 5)
}

// SetEthernetConfigurationStatus sets the ethernet connection status
func (connectionStatus *ConnectionStatus) SetEthernetConfigurationStatus(value bool) {
	setBit(connectionStatus.State2, 5, value)
}

// SetWifiEnabled sets the ethernet connection status
func (connectionStatus *ConnectionStatus) SetWifiEnabled(value bool) {
	setBit(connectionStatus.State1, 0, value)
}

// GetWifiEnabled returns the wifi enabled status
func (connectionStatus *ConnectionStatus) GetWifiEnabled() bool {
	return getBit(connectionStatus.State1, 0)
}

// GetRimoteConnected returns the rimote connected status
func (rimotestatus *RimoteStatus) GetRimoteConnected() bool {
	return getBit(rimotestatus.State1, 0)
}

// GetRimoteGUIDPresent returns the rimote connected status
func (rimotestatus *RimoteStatus) GetRimoteGUIDPresent() bool {
	return getBit(rimotestatus.State1, 1)
}

// SetRimoteConnected sets the rimote connection status
func (rimotestatus *RimoteStatus) SetRimoteConnected(value bool) {
	setBit(rimotestatus.State1, 0, value)
}

// SetRimoteGUIDPresent sets the rimote guid present
func (rimotestatus *RimoteStatus) SetRimoteGUIDPresent(value bool) {
	setBit(rimotestatus.State1, 1, value)
}

func setBit(b *byte, bit uint, value bool) {
	x := *b
	if value {
		x |= (1 << bit)
	} else {
		x &= ^(1 << bit)
	}
	*b = x
}

func getBit(b *byte, bit uint) bool {
	x := *b
	value := x & (1 << bit)
	return value != 0
}

// NewMessage creates Empty structure
func NewMessage() (message *Message) {
	return &Message{Data: [8]byte{}}
}
