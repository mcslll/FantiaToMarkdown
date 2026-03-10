package config

import "fmt"

type Config struct {
	Host       string
	HostUrl    string
	DataDir    string
	CookiePath string
}

func NewConfig(host string, dataDir string, cookiePath string) *Config {
	if host == "" {
		host = "fantia.jp"
	}
	return &Config{
		Host:       host,
		HostUrl:    fmt.Sprintf("https://%s", host),
		DataDir:    dataDir,
		CookiePath: cookiePath,
	}
}
