package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	host := "test.jp"
	dataDir := "./data"
	cookiePath := "./cookies.json"

	cfg := NewConfig(host, dataDir, cookiePath)

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
