package storage

import (
	"FantiaToMarkdown/config"
	"FantiaToMarkdown/fantia"
	"os"
	"path/filepath"
	"strings"
	"testing"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

func TestSavePost(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{DataDir: tmpDir}
	post := fantia.Post{
		Title:   "TestPost",
		Url:     "http://example.com/p1",
		Content: "<h1>Hello</h1>",
	}
	converter := md.NewConverter("", true, nil)

	// 测试保存功能
	err = SavePost(cfg, post, tmpDir, converter)
	if err != nil {
		t.Errorf("SavePost failed: %v", err)
	}

	// 验证文件是否存在
	expectedFile := filepath.Join(tmpDir, "TestPost.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s was not created", expectedFile)
	}

	// 验证内容是否包含基本信息
	content, _ := os.ReadFile(expectedFile)
	contentStr := string(content)
	if !strings.Contains(contentStr, "TestPost") || !strings.Contains(contentStr, "Hello") {
		t.Errorf("File content incorrect: %s", contentStr)
	}
}
