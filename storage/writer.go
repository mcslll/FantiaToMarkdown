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
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/carlmjohnson/requests"
	"golang.org/x/exp/slog"
)

const ImgDir = ".assets"

// SavePost 保存帖子到本地，包括图片下载
func SavePost(cfg *config.Config, post fantia.Post, fileName string, outputDir string, converter *md.Converter) error {
	filePath := filepath.Join(outputDir, fileName)

	// 1. 转换正文 HTML 为 Markdown
	markdown, err := converter.ConvertString(post.Content)
	if err != nil {
		slog.Warn("HTML to Markdown conversion failed, using raw HTML", "title", post.Title)
		markdown = post.Content
	}

	// 2. 下载图片并获取本地引用路径
	picMarkdown, err := downloadAndGetImgMarkdown(cfg, post.Title, outputDir, post.Pictures)
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
	slog.Info("Saving markdown file", "path", filePath)
	return os.WriteFile(filePath, []byte(finalContent), 0644)
}

func downloadAndGetImgMarkdown(cfg *config.Config, postTitle string, outputDir string, urls []string) (string, error) {
	if len(urls) == 0 {
		return "", nil
	}

	assetsDir := filepath.Join(outputDir, ImgDir)
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return "", err
	}

	var mdRefs []string
	for i, urlStr := range urls {
		// 确定后缀名
		ext := ".jpg"
		if strings.Contains(urlStr, ".webp") {
			ext = ".webp"
		} else if strings.Contains(urlStr, ".png") {
			ext = ".png"
		}

		fileName := fmt.Sprintf("%s_%d%s", utils.ToSafeFilename(postTitle), i, ext)
		localFilePath := filepath.Join(assetsDir, fileName)

		// 检查图片是否已存在，存在则跳过下载
		if _, err := os.Stat(localFilePath); err == nil {
			slog.Debug("Image already exists, skipping download", "path", localFilePath)
			relPath := filepath.Join(ImgDir, fileName)
			mdRefs = append(mdRefs, fmt.Sprintf("![image](%s)", relPath))
			continue
		}

		slog.Debug("Downloading image", "url", urlStr, "dest", localFilePath)

		// 增加 60 秒下载超时
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		rb := requests.URL(urlStr).
			Header("User-Agent", fantia.ChromeUserAgent)

		// 设置代理
		if cfg.ProxyUrl != "" {
			proxy, err := url.Parse(cfg.ProxyUrl)
			if err != nil {
				slog.Error("Invalid proxy URL", "url", cfg.ProxyUrl, "error", err)
			} else {
				rb.Transport(&http.Transport{
					Proxy: http.ProxyURL(proxy),
				})
			}
		}

		// 下载图片
		err := rb.ToFile(localFilePath).Fetch(ctx)

		if err != nil {
			slog.Error("Download failed", "url", urlStr, "error", err)
			mdRefs = append(mdRefs, fmt.Sprintf("![image](%s) (Download Failed)", urlStr))
			continue
		}

		// 使用相对路径引用
		relPath := filepath.Join(ImgDir, fileName)
		mdRefs = append(mdRefs, fmt.Sprintf("![image](%s)", relPath))
	}

	return strings.Join(mdRefs, "\n\n"), nil
}
