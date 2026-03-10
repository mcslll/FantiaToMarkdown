package fantia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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
func buildFantiaHeaders(cookieString string, extraHeaders map[string]string) http.Header {
	h := http.Header{
		"User-Agent": {ChromeUserAgent},
		"Cookie":     {cookieString},
		"Accept":     {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	}
	for k, v := range extraHeaders {
		h.Set(k, v)
	}
	return h
}

// NewRequestGet 发送 GET 请求并返回响应字节
func NewRequestGet(host string, url string, cookieString string, extraHeaders ...map[string]string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var extra map[string]string
	if len(extraHeaders) > 0 {
		extra = extraHeaders[0]
	}

	var body bytes.Buffer
	err := requests.
		URL(url).
		Headers(buildFantiaHeaders(cookieString, extra)).
		ToBytesBuffer(&body).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("GET %s failed: %w", url, err)
	}
	return body.Bytes(), nil
}
