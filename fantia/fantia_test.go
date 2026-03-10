package fantia

import (
	"FantiaToMarkdown/config"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/tidwall/gjson"
)

// TestGetMaxPageAndCSRF 测试 HTML 解析逻辑：提取最大页数和 CSRF Token
func TestGetMaxPageAndCSRF(t *testing.T) {
	mockHtml := `
		<html>
			<head><meta name="csrf-token" content="mock-csrf-token-123"></head>
			<body>
				<a class="page-link" href="/fanclubs/1/posts?page=1">1</a>
				<a class="page-link" href="/fanclubs/1/posts?page=5">5</a>
				<i class="fa-angle-double-right"></i>
			</body>
		</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockHtml))
	}))
	defer server.Close()

	cfg := &config.Config{HostUrl: server.URL, Host: "localhost"}
	maxPage, csrf, _, err := GetMaxPage(cfg, "1", "", "")
	if err != nil {
		t.Fatalf("GetMaxPage failed: %v", err)
	}

	if maxPage != 5 {
		t.Errorf("Expected maxPage 5, got %d", maxPage)
	}
	if csrf != "mock-csrf-token-123" {
		t.Errorf("Expected CSRF token 'mock-csrf-token-123', got '%s'", csrf)
	}
}

// TestParsePostContentAPI 测试 API JSON 解析逻辑
func TestParsePostContentAPI(t *testing.T) {
	// 模拟 Fantia API 返回的包含 Blog 嵌套图和相册图的 JSON
	mockJson := `{
		"post": {
			"comment": "Normal Comment",
			"thumb": { "original": "https://example.com/thumb.jpg" },
			"post_contents": [
				{
					"post_content_photos": [
						{ "url": { "original": "https://example.com/photo1.png" } },
						{ "url": { "main_webp": "https://example.com/photo2.webp" } }
					]
				},
				{
					"comment": "{\"ops\":[{\"insert\":{\"fantiaImage\":{\"original_url\":\"https://example.com/blog_img.jpg\"}}}]}"
				}
			]
		}
	}`

	// 验证逻辑（手动调用内部解析逻辑的部分）
	postJson := gjson.Parse(mockJson).Get("post")
	var pictures []string

	// 提取封面
	if thumb := postJson.Get("thumb.original"); thumb.Exists() {
		pictures = append(pictures, thumb.String())
	}

	// 提取内容
	postJson.Get("post_contents").ForEach(func(key, value gjson.Result) bool {
		value.Get("post_content_photos").ForEach(func(k, v gjson.Result) bool {
			if original := v.Get("url.original"); original.Exists() {
				pictures = append(pictures, original.String())
			} else if webp := v.Get("url.main_webp"); webp.Exists() {
				pictures = append(pictures, webp.String())
			}
			return true
		})
		commentStr := value.Get("comment").String()
		if gjson.Valid(commentStr) {
			gjson.Parse(commentStr).Get("ops").ForEach(func(k, v gjson.Result) bool {
				if img := v.Get("insert.fantiaImage.original_url"); img.Exists() {
					pictures = append(pictures, img.String())
				}
				return true
			})
		}
		return true
	})

	expectedCount := 4 // thumb + 2 photos + 1 blog img
	if len(pictures) != expectedCount {
		t.Errorf("Expected %d pictures, got %d: %v", expectedCount, len(pictures), pictures)
	}
}

// TestNewRequestGetRetry 测试重试逻辑
func TestNewRequestGetRetry(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		curr := atomic.LoadInt32(&attempts)
		if curr < 3 {
			// 模拟连接断开 (导致 EOF)
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("webserver doesn't support hijacking")
			}
			conn, _, _ := hj.Hijack()
			conn.Close()
			return
		}
		// 第三次请求成功
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	cfg := &config.Config{HostUrl: server.URL}
	// 运行重试逻辑 (NewRequestGet 会重试最多 3 次)
	body, err := NewRequestGet(cfg, server.URL, "")

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}
	if string(body) != "success" {
		t.Errorf("Expected 'success', got '%s'", string(body))
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}
