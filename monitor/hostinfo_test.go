package main

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestFullWriteRimoteInfo(t *testing.T) {

	hostinfo := &HostInfo{FirmwareVersion: "rm-v1.6-prod-20180118-74",
		ModemEnabled: true,
		SimID:        "8931087616027213997F"}

	err := WriteRimoteInfo("C:\\test\\HostInfo.txt", hostinfo)

	if err != nil {
		t.Errorf("Got unexpected error in write-test: %v", err)
	}

	iobytes, err := ioutil.ReadFile("C:\\test\\HostInfo.txt")

	if err != nil {
		t.Errorf("Got unexpected error in read-after-write-test: %v", err)
	}

	text := string(iobytes)
	if !strings.Contains(text, "firmware-version:") {
		t.Errorf("Missing firmware-version: text in read-after-write-test: %v", err)
	}

	if !strings.Contains(text, "modem-available:") {
		t.Errorf("Missing modem-available: text in read-after-write-test: %v", err)
	}

	if !strings.Contains(text, "sim-number:") {
		t.Errorf("Missing sim-number: text in read-after-write-test: %v", err)
	}
}

/*
func TestWriteRimoteInfoWithModemAvailable(t *testing.T) {

	err := WriteRimoteInfo("C:\\test\\HostInfo.txt", "rm-v1.6-prod-20180118-74", ModemStatusMessage{ModemAvailable: false})

	if err != nil {
		t.Errorf("Got unexpected error in write-test: %v", err)
	}

	iobytes, err := ioutil.ReadFile("C:\\test\\HostInfo.txt")

	if err != nil {
		t.Errorf("Got unexpected error in read-after-write-test: %v", err)
	}

	text := string(iobytes)
	if !strings.Contains(text, "firmware-version:") {
		t.Errorf("Missing firmware-version: text in read-after-write-test: %v", err)
	}

	if !strings.Contains(text, "modem-available:") {
		t.Errorf("Missing modem-available: text in read-after-write-test: %v", err)
	}

	if strings.Contains(text, "sim-number:") {
		t.Errorf("Got sim-number: text in read-after-write-test which does not belong: %v", err)
	}
}
*/
