package storage

import (
	"FantiaToMarkdown/config"
	"FantiaToMarkdown/fantia"
	"FantiaToMarkdown/utils"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/carlmjohnson/requests"
	"golang.org/x/exp/slog"
)

const ImgDir = ".assets"

// SavePost 保存帖子到本地，包括图片下载
func SavePost(cfg *config.Config, post fantia.Post, postID string, fileName string, outputDir string, converter *md.Converter) error {
	filePath := filepath.Join(outputDir, fileName)

	// 1. 转换正文 HTML 为 Markdown
	markdown, err := converter.ConvertString(post.Content)
	if err != nil {
		slog.Warn("HTML to Markdown conversion failed, using raw HTML", "title", post.Title)
		markdown = post.Content
	}

	// 2. 下载图片并获取本地引用路径
	picMarkdown, err := downloadAndGetImgMarkdown(cfg, postID, post.Title, outputDir, post.Pictures)
	if err != nil {
		slog.Error("Failed to download images", "title", post.Title, "error", err)
	}

	// 3. 组装最终内容
	tagsStr := ""
	if len(post.Tags) > 0 {
		tagsStr = fmt.Sprintf("\nTags: %s", strings.Join(post.Tags, ", "))
	}

	// 格式化日期
	displayDate := post.PostedAt
	if t, err := time.Parse(time.RFC1123Z, post.PostedAt); err == nil {
		displayDate = t.Format("2006-01-02 15:04:05")
	}

	header := fmt.Sprintf("# %s\n\nURL: %s\nDate: %s%s\n\n---\n\n", post.Title, post.Url, displayDate, tagsStr)
	
	finalContent := header
	if picMarkdown != "" {
		finalContent += picMarkdown + "\n\n---\n\n"
	}
	finalContent += markdown

	// 4. 写入文件
	slog.Info("Saving", "path", filePath)
	return os.WriteFile(filePath, []byte(finalContent), 0644)
}

func downloadAndGetImgMarkdown(cfg *config.Config, postID string, postTitle string, outputDir string, urls []string) (string, error) {
	if len(urls) == 0 {
		return "", nil
	}

	assetsDir := filepath.Join(outputDir, ImgDir)
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return "", err
	}

	mdRefs := make([]string, len(urls))
	var wg sync.WaitGroup
	// 限制并发：最多同时下载 5 张图片
	sem := make(chan struct{}, 5)

	for i, urlStr := range urls {
		wg.Add(1)
		go func(i int, urlStr string) {
			defer wg.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			// 确定后缀名
			ext := ".jpg"
			if strings.Contains(urlStr, ".webp") {
				ext = ".webp"
			} else if strings.Contains(urlStr, ".png") {
				ext = ".png"
			}

			// 使用 postID_Title_Index 格式以确保唯一性
			fileName := fmt.Sprintf("%s_%s_%d%s", postID, utils.ToSafeFilename(postTitle), i, ext)
			localFilePath := filepath.Join(assetsDir, fileName)

			// 检查图片是否已存在
			if _, err := os.Stat(localFilePath); err == nil {
				slog.Debug("Image exists, skip", "path", localFilePath)
				relPath := filepath.Join(ImgDir, fileName)
				mdRefs[i] = fmt.Sprintf("![image](%s)", relPath)
				return
			}

			// 带有重试逻辑的下载
			var err error
			for retry := 0; retry < 3; retry++ {
				if retry > 0 {
					slog.Debug("Retrying image download", "url", urlStr, "attempt", retry+1)
					time.Sleep(time.Second * time.Duration(retry)) // 退避重试
				}

				err = downloadToFile(cfg, urlStr, localFilePath)
				if err == nil {
					break
				}
			}

			if err != nil {
				slog.Error("Image download failed", "url", urlStr, "error", err)
				mdRefs[i] = fmt.Sprintf("![image](%s) (Download Failed)", urlStr)
				return
			}

			// 使用相对路径引用
			relPath := filepath.Join(ImgDir, fileName)
			mdRefs[i] = fmt.Sprintf("![image](%s)", relPath)
		}(i, urlStr)
	}

	wg.Wait()
	return strings.Join(mdRefs, "\n\n"), nil
}

func downloadToFile(cfg *config.Config, urlStr string, destPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rb := requests.URL(urlStr).
		Header("User-Agent", fantia.ChromeUserAgent)

	// 设置代理
	if cfg.ProxyUrl != "" {
		proxy, err := url.Parse(cfg.ProxyUrl)
		if err != nil {
			return fmt.Errorf("invalid proxy: %w", err)
		}
		rb.Transport(&http.Transport{
			Proxy: http.ProxyURL(proxy),
		})
	}

	return rb.ToFile(destPath).Fetch(ctx)
}
