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

// HardwareStatus structure
type HardwareStatus struct {
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

// HardwareStatus returns the hardware status message
func (message *Message) HardwareStatus() *HardwareStatus {
	return &HardwareStatus{State1: &message.Data[1]}
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

// SetVccStatus sets the vcc status
func (generalStatus *GeneralStatus) SetVccStatus(value bool) {
	setBit(generalStatus.State1, 2, value)
}

// GetVccStatus get the software status
func (generalStatus *GeneralStatus) GetVccStatus() bool {
	return getBit(generalStatus.State1, 2)
}

// SetSoftwareStatus sets the software status
func (generalStatus *GeneralStatus) SetSoftwareStatus(value bool) {
	setBit(generalStatus.State1, 1, value)
}

// GetSoftwareStatus get the software status
func (generalStatus *GeneralStatus) GetSoftwareStatus() bool {
	return getBit(generalStatus.State1, 1)
}

// SetNandStatus sets the nand status
func (hardwareStatus *HardwareStatus) SetNandStatus(value bool) {
	setBit(hardwareStatus.State1, 3, value)
}

// GetNandStatus get the software status
func (hardwareStatus *HardwareStatus) GetNandStatus() bool {
	return getBit(hardwareStatus.State1, 3)
}

// SetEth0Status sets the connection status of eth0
func (connectionStatus *ConnectionStatus) SetEth0Status(value bool) {
	setBit(connectionStatus.State2, 6, value)
}

// GetEth0Status return the connection status of eth1
func (connectionStatus *ConnectionStatus) GetEth0Status() bool {
	return getBit(connectionStatus.State2, 6)
}

// SetEth1Status sets the connection status of eth1
func (connectionStatus *ConnectionStatus) SetEth1Status(value bool) {
	setBit(connectionStatus.State2, 7, value)
}

// GetEth1Status return the connection status of eth1
func (connectionStatus *ConnectionStatus) GetEth1Status() bool {
	return getBit(connectionStatus.State2, 7)
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

// SetMobileInternetEnabled sets the ethernet connection status
func (connectionStatus *ConnectionStatus) SetMobileInternetEnabled(value bool) {
	setBit(connectionStatus.State1, 4, value)
}

// GetMobileInternetEnabled returns the wifi enabled status
func (connectionStatus *ConnectionStatus) GetMobileInternetEnabled() bool {
	return getBit(connectionStatus.State1, 4)
}

// GetSimpinOk returns the sim status
func (connectionStatus *ConnectionStatus) GetSimpinOk() bool {
	return getBit(connectionStatus.State1, 5)
}

// SetSimPinOK sets the sim status
func (connectionStatus *ConnectionStatus) SetSimPinOK(value bool) {
	setBit(connectionStatus.State1, 5, value)
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

// GetRimoteSSLStatus returns the rimote SSL status
func (rimotestatus *RimoteStatus) GetRimoteSSLStatus() bool {
	return getBit(rimotestatus.State1, 2)
}

// SetRimoteSSLOk sets the rimote SSL status
func (rimotestatus *RimoteStatus) SetRimoteSSLOk(value bool) {
	setBit(rimotestatus.State1, 2, value)
}

// GetRimoteConfOk returns the rimote configuration status
func (rimotestatus *RimoteStatus) GetRimoteConfOk() bool {
	return getBit(rimotestatus.State1, 3)
}

// SetRimoteConfOk sets the rimote configuration status
func (rimotestatus *RimoteStatus) SetRimoteConfOk(value bool) {
	setBit(rimotestatus.State1, 3, value)
}

// SetModemSignal sets the modem signal
func (connectionStatus *ConnectionStatus) SetModemSignal(signalStrength SignalStrength) {

	if signalStrength == NoSignal {
		setBit(connectionStatus.State2, 1, true)
		setBit(connectionStatus.State2, 2, true)
		setBit(connectionStatus.State2, 3, true)
	}

	if signalStrength == WeakSignal {
		setBit(connectionStatus.State2, 1, true)
		setBit(connectionStatus.State2, 2, false)
		setBit(connectionStatus.State2, 3, false)
	}

	if signalStrength == FairSignal {
		setBit(connectionStatus.State2, 1, false)
		setBit(connectionStatus.State2, 2, true)
		setBit(connectionStatus.State2, 3, false)
	}

	if signalStrength == GoodSignal {
		setBit(connectionStatus.State2, 1, false)
		setBit(connectionStatus.State2, 2, false)
		setBit(connectionStatus.State2, 3, true)
	}
}

// SetWifiSignal sets the wifi signal
func (connectionStatus *ConnectionStatus) SetWifiSignal(signalStrength SignalStrength) {

	if signalStrength == NoSignal {
		setBit(connectionStatus.State1, 1, true)
		setBit(connectionStatus.State1, 2, true)
		setBit(connectionStatus.State1, 3, true)
	}

	if signalStrength == WeakSignal {
		setBit(connectionStatus.State1, 1, true)
		setBit(connectionStatus.State1, 2, false)
		setBit(connectionStatus.State1, 3, false)
	}

	if signalStrength == FairSignal {
		setBit(connectionStatus.State1, 1, false)
		setBit(connectionStatus.State1, 2, true)
		setBit(connectionStatus.State1, 3, false)
	}

	if signalStrength == GoodSignal {
		setBit(connectionStatus.State1, 1, false)
		setBit(connectionStatus.State1, 2, false)
		setBit(connectionStatus.State1, 3, true)
	}
}

// SetBroadbandConnectionType sets the wifi signal
func (connectionStatus *ConnectionStatus) SetBroadbandConnectionType(broadbandConnType BroadbandConnType) {

	if broadbandConnType == ConnTypeNoNetwork {
		setBit(connectionStatus.State1, 6, false)
		setBit(connectionStatus.State1, 7, false)
	}

	if broadbandConnType == ConnType2G {
		setBit(connectionStatus.State1, 6, true)
		setBit(connectionStatus.State1, 7, false)
	}

	if broadbandConnType == ConnType3G || broadbandConnType == ConnType4G {
		setBit(connectionStatus.State1, 6, true)
		setBit(connectionStatus.State1, 7, true)
	}
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
