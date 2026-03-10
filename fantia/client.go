package fantia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/carlmjohnson/requests"
)

const (
	DelayMs         = 500
	ChromeUserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36`
)

// ReadCookiesFromFile 从文件中读取 Cookies
func ReadCookiesFromFile(filePath string) ([]Cookie, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cookies []Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cookies: %w", err)
	}

	return cookies, nil
}

// GetCookiesString 获取 Cookies 字符串
func GetCookiesString(cookies []Cookie) string {
	var cookiesString string
	for _, cookie := range cookies {
		cookiesString += cookie.Name + "=" + cookie.Value + "; "
	}
	return cookiesString
}

// GetCookies 从指定路径获取 Cookies 字符串
func GetCookies(cookiePath string) (string, error) {
	cookies, err := ReadCookiesFromFile(cookiePath)
	if err != nil {
		return "", fmt.Errorf("failed to read cookies from file: %w", err)
	}
	return GetCookiesString(cookies), nil
}

// buildFantiaHeaders 构建 Fantia 专用的请求头
func buildFantiaHeaders(host string, cookieString string) http.Header {
	return http.Header{
		"User-Agent": {ChromeUserAgent},
		"Cookie":     {cookieString},
		"Host":       {host},
	}
}

// NewRequestGet 发送 GET 请求并返回响应字节
func NewRequestGet(host string, url string, cookieString string) ([]byte, error) {
	var body bytes.Buffer
	err := requests.
		URL(url).
		Headers(buildFantiaHeaders(host, cookieString)).
		ToBytesBuffer(&body).
		Fetch(context.Background())
	if err != nil {
		return nil, fmt.Errorf("GET %s failed: %w", url, err)
	}
	return body.Bytes(), nil
}
