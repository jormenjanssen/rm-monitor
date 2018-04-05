package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// DeviceFirmwareFilePath The path of the file containing the firmware version info
const DeviceFirmwareFilePath string = "/etc/mender/artifact_info"

// DeviceRimoteInfoFilePath The path of the file containing the firmware version info
const DeviceRimoteInfoFilePath string = "/usr/share/Riwo/Rimote/HostInfo.txt"

// SystemConfigurationFilePath The path of the file containing the system parameters
const SystemConfigurationFilePath string = "/data/system/configuration.xml"

// FactorySystemConfigurationFilePath The path of the file containing the system parameters (factory-default)
const FactorySystemConfigurationFilePath string = "/usr/local/rimote/riwo.rimote-management/app/factory.xml"

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

	// Update our sim-id, if we got a vallid one which is not equal to our current
	if newInfo.SimID != "" && hostInfo.SimID != newInfo.SimID {

		// Only update our sim-id, if we're not using the system factory config.
		if !DeviceIsUsingFactoryConfig() {
			hostInfo.SimID = newInfo.SimID
			updated = true
		}
	}

	return updated
}

// HandleHostInfo handles host info writes
func HandleHostInfo(ctx context.Context, logger *Logger, hostInfoInputChannel <-chan HostInfo) {

	// Handle inside go routine
	go func() {

		// Run some tests
		runUpfronConfigurationChecks(logger)

		emptyInfo := HostInfo{}
		firmwareVersion, _ := GetFirmwareVersion()
		current := &HostInfo{FirmwareVersion: firmwareVersion}

		// Initial select case where we want to wait max 45 seconds before giving up and writing the result anyway without sim-info
		select {
		// Context exit requested
		case <-ctx.Done():
			return
		// We got hostinfo before the timeout
		case hostmsg := <-hostInfoInputChannel:
			writeInternal(logger, current, hostmsg, false)
		// We got no hostinfo before the timeout force the current one with only firmware information
		case <-time.After(45 * time.Second):
			writeInternal(logger, current, emptyInfo, true)
		}

		// Further case where we only want to write, if we got a new sim-id
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

func runUpfronConfigurationChecks(logger *Logger) {

	if err := FactoryFilePresent(); err != nil {
		logger.WarningF("The factory system configuration file is not present error: %v", err)
	}

	if err := SystemConfigFilePresent(); err != nil {
		logger.WarningF("The system configuration file is not present error: %v", err)
	}

	factorySha1 := "undected"
	configFileSha1 := "undected"

	var err error

	if factorySha1, err = GetSha1SumFromFile(FactorySystemConfigurationFilePath); err != nil {
		logger.WarningF("Cannot determina sha1 for factory system config file: %v", err)
		return
	}

	if configFileSha1, err = GetSha1SumFromFile(SystemConfigurationFilePath); err != nil {
		logger.WarningF("Cannot determina sha1 for system config file: %v", err)
		return
	}

	matching := factorySha1 == configFileSha1

	if IsDebugMode() {
		logger.DebugF("Running SHA1 comparison between Factory[%v] and SystemConfig[%v] Matching: %v", factorySha1, configFileSha1, matching)
	}

	if matching {
		logger.InfoF("The system is running with factory default configuration. (Ignoring sim-id)")
	}
}

// GetFirmwareVersion return the firmware version from the Firmware
func GetFirmwareVersion() (string, error) {

	if !IsTargetDevice() {
		return "[DEVELOPMENT]", nil
	}

	databytes, err := ioutil.ReadFile(DeviceFirmwareFilePath)
	version := strings.TrimSpace(strings.Split(string(databytes), "=")[1])

	if err != nil {
		return "[UNDECTABLE]", err
	}

	return version, nil
}

func writeInternal(logger *Logger, currentInfo *HostInfo, newInfo HostInfo, forced bool) {

	// Update properties where applicable
	infoUpdated := currentInfo.UpdateModemInfo(newInfo)

	// Write debug/verbose logging.
	if infoUpdated && IsDebugMode() {
		writeHostInfoToLogger(logger, currentInfo)
	}

	// Check if we need to write our new host info to the system.
	if forced || (infoUpdated && checkWrite(currentInfo)) {

		if IsDebugMode() {
			logger.DebugF("Writing rimote connection info [forced: %v] [factory-config: %v]", forced, DeviceIsUsingFactoryConfig())
		}

		// Perform the actual write
		err := WriteRimoteInfo(DeviceRimoteInfoFilePath, currentInfo)

		if err != nil {
			logger.Errorf("Could not write Rimote info file @ path: %v error: %v", DeviceRimoteInfoFilePath, err)
		}
	}
}

func writeHostInfoToLogger(logger *Logger, hostInfo *HostInfo) {
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
		fmt.Fprintln(buffer, fmt.Sprintf("sim-number: %v", hostInfo.SimID))
	}

	data := buffer.Bytes()
	dataBytes := []byte(data)
	return ioutil.WriteFile(path, dataBytes, 0666)
}

// DeviceIsUsingFactoryConfig return true if this device is using the factory config
func DeviceIsUsingFactoryConfig() bool {

	// Return false in debug mode.
	if IsDebugMode() {
		return false
	}

	// Without a factory file we can't do comparison
	if FactoryFilePresent() != nil {
		return false
	}

	// Without a system configuration consider our self valid
	if SystemConfigFilePresent() != nil {
		return true
	}

	factorySha1, _ := GetSha1SumFromFile(FactorySystemConfigurationFilePath)
	configFileSha1, _ := GetSha1SumFromFile(SystemConfigurationFilePath)

	// Run sha1 comparison
	if factorySha1 == configFileSha1 {
		return true
	}

	return false
}

// FactoryFilePresent check if the factory file is present and return nil when everyting works like expected
func FactoryFilePresent() error {
	_, err := os.Stat(FactorySystemConfigurationFilePath)
	return err
}

// SystemConfigFilePresent check if the factory file is present and return nil when everyting works like expected
func SystemConfigFilePresent() error {
	_, err := os.Stat(SystemConfigurationFilePath)
	return err
}

// GetSha1SumFromFile return sha1 from path
func GetSha1SumFromFile(path string) (string, error) {

	f, err := os.Open(path)

	if err != nil {
		return "", err
	}

	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
