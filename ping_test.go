package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGraphDefinition(t *testing.T) {
	var pp PingPlugin
	pp.Hosts = []string{"127.0.0.1"}
	pp.Labels = []string{"localhost"}

	gd := pp.GraphDefinition()

	actual := gd["ping.rtt"].Label
	expected := "Ping Round Trip Times"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = gd["ping.rtt"].Unit
	expected = "float"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual_stacked := gd["ping.rtt"].Metrics[0].Stacked
	expected_stacked := false
	if actual_stacked != expected_stacked {
		t.Errorf("got %v\nwant %v", actual_stacked, expected_stacked)
	}

	actual = fmt.Sprintf("%v", reflect.TypeOf(gd["ping.rtt"].Metrics))
	expected = "[]mackerelplugin.Metrics"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestEscapeHostName(t *testing.T) {
	actual := escapeHostName("127.0.0.1")
	expected := "127_0_0_1"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = escapeHostName("8.8.8.8")
	expected = "8_8_8_8"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = escapeHostName("8_8_8_8")
	expected = "8_8_8_8"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestValidate(t *testing.T) {
	actual := validate("127.0.0.1")
	expected := true
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = validate("8.8.8.8")
	expected = true
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = validate("8.8.8.")
	expected = false
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = validate("localhost")
	expected = false
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestParseHostsString(t *testing.T) {
	actualIPs, actualLabels, err := parseHostsString("127.0.0.1", false)
	expected := []string{"127.0.0.1"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("8.8.8.8,8.8.4.4", false)
	expected = []string{"8.8.8.8", "8.8.4.4"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}
	if actualIPs[1] != expected[1] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[1] != expected[1] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("8.8.8.8:google-public-dns-a", false)
	expected = []string{"8.8.8.8"}
	expected_labels := []string{"google-public-dns-a"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("8.8.8.8:google-public-dns-a,8.8.4.4:google-public-dns-b", false)
	expected = []string{"8.8.8.8", "8.8.4.4"}
	expected_labels = []string{"google-public-dns-a", "google-public-dns-b"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}
	if actualIPs[1] != expected[1] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[1] != expected_labels[1] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	_, _, err = parseHostsString("8.8.8.", false, "1")
	if err == nil {
		t.Errorf("got %v", err)
	}

	actualIPs, actualLabels, err = parseHostsString("m.root-servers.net", false, "1")
	expected = []string{"202.12.27.33"}
	expected_labels = []string{"m.root-servers.net"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("m.root-servers.net:m-root", false)
	expected = []string{"202.12.27.33"}
	expected_labels = []string{"m-root"}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}
}


func TestGraphDefinitionIPv6(t *testing.T) {
	var pp PingPlugin
	pp.Hosts = []string{"::1"}
	pp.Labels = []string{"localhost"}

	gd := pp.GraphDefinition()

	actual := gd["ping.rtt"].Label
	expected := "Ping Round Trip Times"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = gd["ping.rtt"].Unit
	expected = "float"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual_stacked := gd["ping.rtt"].Metrics[0].Stacked
	expected_stacked := false
	if actual_stacked != expected_stacked {
		t.Errorf("got %v\nwant %v", actual_stacked, expected_stacked)
	}

	actual = fmt.Sprintf("%v", reflect.TypeOf(gd["ping.rtt"].Metrics))
	expected = "[]mackerelplugin.Metrics"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestEscapeHostNameIPv6(t *testing.T) {
	actual := escapeHostName("::1")
	expected := "__1"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = escapeHostName("2001:4860:4860::8888")
	expected = "2001_4860_4860__8888"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}

	actual = escapeHostName("2001_4860_4860__8888")
	expected = "2001_4860_4860__8888"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestParseHostsStringIPv6(t *testing.T) {
	actualIPs, actualLabels, err := parseHostsString("[::1]", true)
	expected := []string{"::1"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("[2001:4860:4860::8888],[2001:4860:4860::8844]", true)
	expected = []string{"2001:4860:4860::8888", "2001:4860:4860::8844"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}
	if actualIPs[1] != expected[1] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[1] != expected[1] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("[2001:4860:4860::8888]:google-public-dns-a", true)
	expected = []string{"2001:4860:4860::8888"}
	expected_labels := []string{"google-public-dns-a"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("[2001:4860:4860::8888]:google-public-dns-a,[2001:4860:4860::8844]:google-public-dns-b", true)
	expected = []string{"2001:4860:4860::8888", "2001:4860:4860::8844"}
	expected_labels = []string{"google-public-dns-a", "google-public-dns-b"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}
	if actualIPs[1] != expected[1] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[1] != expected_labels[1] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	_, _, err = parseHostsString("2001:4860:4860:", true, "1")
	if err == nil {
		t.Errorf("got %v", err)
	}

	actualIPs, actualLabels, err = parseHostsString("m.root-servers.net", true, "1")
	expected = []string{"2001:dc3::35"}
	expected_labels = []string{"m.root-servers.net"}
	if err != nil {
		t.Errorf("got %v", err)
	}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}

	actualIPs, actualLabels, err = parseHostsString("m.root-servers.net:m-root", true)
	expected = []string{"2001:dc3::35"}
	expected_labels = []string{"m-root"}
	if actualIPs[0] != expected[0] {
		t.Errorf("got %v\nwant %v", actualIPs, expected)
	}
	if actualLabels[0] != expected_labels[0] {
		t.Errorf("got %v\nwant %v", actualLabels, expected)
	}
}
