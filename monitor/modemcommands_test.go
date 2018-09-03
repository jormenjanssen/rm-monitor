package main

import (
	"strconv"
	"strings"
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

func TestCnsMod(t *testing.T) {

	str := "+CNSMOD: 0,5"
	ct := ConnTypeNoNetwork

	items := strings.Split(str, ",")
	n, err := strconv.Atoi(items[1])
	if err == nil && n > 0 {
		if n <= 3 {
			ct = ConnType2G
		} else {
			ct = ConnType3G
		}
	}

	if ct != ConnType3G {
		t.Errorf("Invallid connection type expected: %v but got: %v", ConnType3G, ct)
	}
}
