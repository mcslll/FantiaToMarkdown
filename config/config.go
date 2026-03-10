package config

import "fmt"

// Config 统一配置结构体，消除全局可变状态
type Config struct {
	Host       string // 主站域名，如 "fantia.jp"
	HostUrl    string // 完整 URL，如 "https://fantia.jp"
	DataDir    string // 数据存储目录
	CookiePath string // cookie 文件路径
	ProxyUrl   string // 代理服务器地址，如 "http://127.0.0.1:7890"
}

// NewConfig 创建配置，自动根据 host 生成 HostUrl
func NewConfig(host string, dataDir string, cookiePath string, proxyUrl string) *Config {
	if host == "" {
		host = "fantia.jp"
	}
	return &Config{
		Host:       host,
		HostUrl:    fmt.Sprintf("https://%s", host),
		DataDir:    dataDir,
		CookiePath: cookiePath,
		ProxyUrl:   proxyUrl,
	}
}
