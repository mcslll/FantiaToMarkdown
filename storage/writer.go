package storage

import (
	"FantiaToMarkdown/config"
	"FantiaToMarkdown/fantia"
	"FantiaToMarkdown/utils"
	"context"
	"fmt"
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
func SavePost(cfg *config.Config, post fantia.Post, outputDir string, converter *md.Converter) error {
	filePath := filepath.Join(outputDir, post.Title+".md")

	// 1. 转换正文 HTML 为 Markdown
	markdown, err := converter.ConvertString(post.Content)
	if err != nil {
		slog.Warn("HTML to Markdown conversion failed, using raw HTML", "title", post.Title)
		markdown = post.Content
	}

	// 2. 下载图片并获取本地引用路径
	picMarkdown, err := downloadAndGetImgMarkdown(post.Title, outputDir, post.Pictures)
	if err != nil {
		slog.Error("Failed to download images", "title", post.Title, "error", err)
	}

	// 3. 组装最终内容
	header := fmt.Sprintf("# %s\n\nURL: %s\n\n---\n\n", post.Title, post.Url)
	finalContent := header + markdown
	if picMarkdown != "" {
		finalContent += "\n\n### 图片附件\n\n" + picMarkdown
	}

	// 4. 写入文件
	return os.WriteFile(filePath, []byte(finalContent), 0644)
}

func downloadAndGetImgMarkdown(postTitle string, outputDir string, urls []string) (string, error) {
	if len(urls) == 0 {
		return "", nil
	}

	assetsDir := filepath.Join(outputDir, ImgDir)
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return "", err
	}

	var mdRefs []string
	for i, url := range urls {
		// 确定后缀名
		ext := ".jpg"
		if strings.Contains(url, ".webp") {
			ext = ".webp"
		} else if strings.Contains(url, ".png") {
			ext = ".png"
		}

		fileName := fmt.Sprintf("%s_%d%s", utils.ToSafeFilename(postTitle), i, ext)
		localFilePath := filepath.Join(assetsDir, fileName)

		slog.Debug("Downloading image", "url", url, "dest", localFilePath)

		// 增加 60 秒下载超时
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// 下载图片
		err := requests.URL(url).
			Header("User-Agent", fantia.ChromeUserAgent).
			ToFile(localFilePath).
			Fetch(ctx)


		if err != nil {
			slog.Error("Download failed", "url", url, "error", err)
			mdRefs = append(mdRefs, fmt.Sprintf("![image](%s) (Download Failed)", url))
			continue
		}

		// 使用相对路径引用
		relPath := filepath.Join(ImgDir, fileName)
		mdRefs = append(mdRefs, fmt.Sprintf("![image](%s)", relPath))
	}

	return strings.Join(mdRefs, "\n\n"), nil
}
