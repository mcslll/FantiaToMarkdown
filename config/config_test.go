package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	host := "test.jp"
	dataDir := "./data"
	cookiePath := "./cookies.json"
	proxy := ""

	cfg := NewConfig(host, dataDir, cookiePath, proxy)

	if cfg.Host != host {
		t.Errorf("Expected Host %s, got %s", host, cfg.Host)
	}
	if cfg.HostUrl != "https://"+host {
		t.Errorf("Expected HostUrl https://%s, got %s", host, cfg.HostUrl)
	}
	if cfg.DataDir != dataDir {
		t.Errorf("Expected DataDir %s, got %s", dataDir, cfg.DataDir)
	}
}

func TestProxyUrlAutoFix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"127.0.0.1:7890", "http://127.0.0.1:7890"},
		{"http://127.0.0.1:7890", "http://127.0.0.1:7890"},
		{"socks5://127.0.0.1:1080", "socks5://127.0.0.1:1080"},
		{"", ""},
	}

	for _, tc := range tests {
		cfg := NewConfig("", "", "", tc.input)
		if cfg.ProxyUrl != tc.expected {
			t.Errorf("Input '%s': Expected proxy '%s', got '%s'", tc.input, tc.expected, cfg.ProxyUrl)
		}
	}
}
