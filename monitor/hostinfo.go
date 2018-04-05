package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

// DeviceFirmwareFilePath The path of the file containing the firmware version info
const DeviceFirmwareFilePath string = "/etc/mender/artifact_info"

// DeviceRimoteInfoFilePath The path of the file containing the firmware version info
const DeviceRimoteInfoFilePath string = "/usr/share/Riwo/Rimote/HostInfo.txt"

// HostInfo structure
type HostInfo struct {
	FirmwareVersion string
	ModemEnabled    bool
	SimID           string
}

// ModemInfoPresent checks if modem info is present
func (hostInfo *HostInfo) ModemInfoPresent() bool {
	return hostInfo.ModemEnabled || hostInfo.SimID != ""
}

// UpdateModemInfo updates the modem info and returns true if the modem info is updated
func (hostInfo *HostInfo) UpdateModemInfo(newInfo HostInfo) bool {

	updated := false

	// Update our modem info only if we where previous disabled.
	// The modem can be temp offline.
	if !hostInfo.ModemEnabled && newInfo.ModemEnabled {
		hostInfo.ModemEnabled = newInfo.ModemEnabled
		updated = true
	}

	// Update our sim id only if we got a vallid one which is not equal to our current
	if newInfo.SimID != "" && hostInfo.SimID != newInfo.SimID {
		hostInfo.SimID = newInfo.SimID
		updated = true
	}

	return updated
}

// HandleHostInfo handles host info writes
func HandleHostInfo(ctx context.Context, logger *Logger, hostInfoInputChannel <-chan HostInfo) {

	// Handle inside go routine
	go func() {

		emptyInfo := HostInfo{}
		firmwareVersion, _ := GetFirmwareVersion()
		current := &HostInfo{FirmwareVersion: firmwareVersion}

		// Initial case where we want to 45 seconds before giving up and writing the result anyway without sim-info
		select {
		case <-ctx.Done():
			return
		case <-time.After(45 * time.Second):
			writeInternal(logger, current, emptyInfo, true)
		}

		// Further case where we only want to write if we got a new sim-id
		for {
			select {
			case <-ctx.Done():
				return
			case hostmsg := <-hostInfoInputChannel:
				writeInternal(logger, current, hostmsg, false)
			}
		}
	}()
}

// GetFirmwareVersion return the firmware version from the Firmware
func GetFirmwareVersion() (string, error) {

	if !IsTargetDevice() {
		return "[DEVELOPMENT]", nil
	}

	databytes, err := ioutil.ReadFile(DeviceFirmwareFilePath)
	version := strings.Split(string(databytes), "=")[1]

	if err != nil {
		return "[UNDECTABLE]", err
	}

	return version, nil
}

func writeInternal(logger *Logger, currentInfo *HostInfo, newInfo HostInfo, forced bool) {

	// Update properties where applicable
	infoUpdated := currentInfo.UpdateModemInfo(newInfo)

	// Write debug/verbose logging.
	if IsDebugMode() {
		writeHostInfoToLogger(logger, currentInfo)

		if infoUpdated {
			logger.Debug("The hostinfo is updated")
		}
	}

	// Check if we need to write our new host info to the system.
	if forced || checkWrite(currentInfo) {

		if IsDebugMode() {
			logger.DebugF("Writing rimote connection info [forced: %v]", forced)
		}

		// Perform the actual write
		err := WriteRimoteInfo(DeviceRimoteInfoFilePath, currentInfo)

		if err != nil {
			logger.Errorf("Could not write Rimote info file @ path: %v error: %v", DeviceRimoteInfoFilePath, err)
		}
	}
}

func writeHostInfoToLogger(logger *Logger, hostInfo *HostInfo) {

	logger.Debug("Writing HostmodemInfo")
	logger.DebugF("HostmodemInfo[FirmwareVersion]: %v", hostInfo.FirmwareVersion)
	logger.DebugF("HostmodemInfo[ModemEnabled]: %v", hostInfo.ModemEnabled)
	logger.DebugF("HostmodemInfo[SimID]: %v", hostInfo.SimID)
}

func checkWrite(hostInfo *HostInfo) bool {
	return hostInfo.ModemInfoPresent()
}

// WriteRimoteInfo writes HostDeviceInfo file
func WriteRimoteInfo(path string, hostInfo *HostInfo) error {

	db := make([]byte, 0)
	buffer := bytes.NewBuffer(db)
	fmt.Fprintln(buffer, fmt.Sprintf("firmware-version: %v", hostInfo.FirmwareVersion))
	fmt.Fprintln(buffer, fmt.Sprintf("modem-available: %v", hostInfo.ModemEnabled))

	if hostInfo.SimID != "" {
		fmt.Fprint(buffer, fmt.Sprintf("sim-number: %v", hostInfo.SimID))
	}

	data := buffer.Bytes()
	dataBytes := []byte(data)
	return ioutil.WriteFile(path, dataBytes, 0666)
}
