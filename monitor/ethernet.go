package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// NewEthernetMonitor create new ethernet monitor
func NewEthernetMonitor(ctx context.Context, ethernetMessageChannel chan EthernetMessage) *Monitor {
	monitor := &Monitor{}

	// Run our logic inside a new goroutine.
	go monitor.run(ctx, ethernetMessageChannel)

	return monitor
}

// EthernetMessage type
type EthernetMessage struct {
	Eth0               Adapter
	Eth1               Adapter
	Wifi0              Adapter
	Ppp0               Adapter
	EthernetConfigured bool
}

// Monitor type
type Monitor struct {
}

func (*Monitor) run(ctx context.Context, ethernetMessageChannel chan EthernetMessage) {

	eth0 := &Adapter{Name: "eth0"}
	eth1 := &Adapter{Name: "eth1"}
	wifi0 := &Adapter{Name: "wifi0"}
	ppp0 := &Adapter{Name: "ppp0"}

	defaultTimeout := 15 * time.Second

	for {

		// Refresh our adapter states
		eth0.update()
		eth1.update()
		wifi0.update()
		ppp0.update()

		ethernetMessageChannel <- EthernetMessage{
			Eth0:               Adapter{Name: eth0.Name, Connected: eth0.Connected},
			Eth1:               Adapter{Name: eth1.Name, Connected: eth1.Connected},
			Wifi0:              Adapter{Name: wifi0.Name, Connected: wifi0.Connected},
			Ppp0:               Adapter{Name: ppp0.Name, Connected: ppp0.Connected},
			EthernetConfigured: eth0.Configured || eth1.Configured}

		time.Sleep(defaultTimeout)
	}
}

// Adapter structure
type Adapter struct {
	Name       string
	Connected  bool
	Configured bool
}

func (adapter *Adapter) update() {

	// Check if the adapter exsists on the OS.
	intf, err := net.InterfaceByName(adapter.Name)

	if err != nil {
		adapter.Configured = false
		adapter.Connected = false
		return
	}

	// No errors mean we are configured.
	adapter.Configured = true

	// Check if the connection is up.
	if strings.Contains(intf.Flags.String(), "up") {
		adapter.Connected = adapter.checkCarrierState()
	} else {
		adapter.Connected = false
	}
}

func (adapter *Adapter) checkCarrierState() bool {
	path := fmt.Sprintf("/sys/class/net/%v/carrier", adapter.Name)
	file, err := os.OpenFile(path, os.O_RDONLY, 0600)

	if err != nil {
		return false
	}

	// Defer file closing
	defer file.Close()

	buf := make([]byte, 1)
	_, err = file.Read(buf)

	if err != nil {
		return false
	}

	c := buf[0]

	switch c {
	case '0':
		return false
	case '1':
		return true
	default:
		return false
	}
}
