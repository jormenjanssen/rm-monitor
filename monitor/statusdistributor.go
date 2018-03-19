package main

import (
	"errors"
	"net"
)

// SendMessage send an udp message
func SendMessage(address string, message [8]byte) error {

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	c, err := net.DialUDP("udp", nil, addr)

	if err != nil {
		return err
	}

	n, err := c.Write(message[:])

	if n != 8 && err == nil {
		return errors.New("Invallid data length")
	}

	return err
}
