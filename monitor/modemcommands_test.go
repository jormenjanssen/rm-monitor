package main

import (
	"testing"
)

func TestGetCcidFromCcidLine(t *testing.T) {

	str := getCcidFromCcidLine("+CCID: \"8931087616027213997F\"")

	if str == "" {
		t.Errorf("Test failed got empty string")
	}

	if str != "8931087616027213997F" {
		t.Errorf("Exected output: 8931087616027213997F got: %v", str)
	}
}
