package main

import (
	"errors"
	"net"
)

var udpConnection = CreateUDPConnection()

// UDPConnection struct
type UDPConnection struct {
	Address    *net.UDPAddr
	connection net.Conn
}

// CreateUDPConnection creates udp connection
func CreateUDPConnection() *UDPConnection {

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9876")
	return &UDPConnection{Address: addr}
}

// Execute UDP call
func (udpConnection *UDPConnection) Execute(logger *Logger, f func(net.Conn) error) error {

	if udpConnection.connection == nil {
		udp, err := net.DialUDP("udp", nil, udpConnection.Address)
		if err == nil {
			udpConnection.connection = udp
			err = f(udp)
			if err != nil {
				udpConnection.connection.Close()
				udpConnection.connection = nil
			}

			return err
		}

		return err
	}

	udp := udpConnection.connection

	err := f(udp)
	if err != nil {
		udpConnection.connection.Close()
		udpConnection.connection = nil

		logger.DebugF("Cannot send distributed status message: %v", err)

	} else {
		return nil
	}

	udpConnection.connection = nil
	return errors.New("Invallid cast")
}

// SendMessage send an udp message
func SendMessage(logger *Logger, message [8]byte) error {

	return udpConnection.Execute(logger, func(udp net.Conn) error {

		n, err := udp.Write(message[:])
		if n != 8 && err == nil {
			return errors.New("Invallid data length")
		}

		return err
	})
}
