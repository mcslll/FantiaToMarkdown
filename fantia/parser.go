package fantia

import (
	"FantiaToMarkdown/config"
	"FantiaToMarkdown/utils"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/exp/slog"
)

func GetMaxPage(cfg *config.Config, fanclubID string, cookieString string) (int, error) {
	apiUrl := fmt.Sprintf("%s/fanclubs/%s/posts", cfg.HostUrl, fanclubID)
	body, err := NewRequestGet(cfg.Host, apiUrl, cookieString)
	if err != nil {
		return 1, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return 1, err
	}

	maxPage := 1
	re := regexp.MustCompile(`page=(\d+)`)

	// Try fa-angle-double-right
	doc.Find("i.fa-angle-double-right").Each(func(i int, s *goquery.Selection) {
		parent := s.Parent()
		if parent.Is("a.page-link") {
			href, _ := parent.Attr("href")
			match := re.FindStringSubmatch(href)
			if len(match) > 1 {
				p, _ := strconv.Atoi(match[1])
				if p > maxPage {
					maxPage = p
				}
			}
		}
	})

	// Fallback to all page-links
	doc.Find("a.page-link").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		match := re.FindStringSubmatch(href)
		if len(match) > 1 {
			p, _ := strconv.Atoi(match[1])
			if p > maxPage {
				maxPage = p
			}
		}
	})

	return maxPage, nil
}

func CollectPostsFromPage(cfg *config.Config, fanclubID string, page int, cookieString string) ([]Post, error) {
	apiUrl := fmt.Sprintf("%s/fanclubs/%s/posts?page=%d", cfg.HostUrl, fanclubID, page)
	body, err := NewRequestGet(cfg.Host, apiUrl, cookieString)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var posts []Post
	doc.Find("a.link-block").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Attr("title")
		href, _ := s.Attr("href")
		if strings.Contains(href, "/posts/") {
			fullUrl := cfg.HostUrl + href
			posts = append(posts, Post{
				Title: utils.ToSafeFilename(title),
				Url:   fullUrl,
			})
		}
	})

	return posts, nil
}

func GetPostContent(cfg *config.Config, postUrl string, cookieString string) (string, error) {
	body, err := NewRequestGet(cfg.Host, postUrl, cookieString)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	// 提取正文
	contentHtml, err := doc.Find(".post-content, .post-body, .post-description").Html()
	if err != nil {
		return "", fmt.Errorf("failed to find post content: %w", err)
	}

	return contentHtml, nil
}

func GetPosts(cfg *config.Config, fanclubID string, cookieString string) error {
	maxPage, err := GetMaxPage(cfg, fanclubID, cookieString)
	if err != nil {
		return err
	}
	slog.Info("Detected Max Page", "maxPage", maxPage)

	outputDir := filepath.Join(cfg.DataDir, fanclubID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	converter := md.NewConverter("", true, nil)

	for page := 1; page <= maxPage; page++ {
		slog.Info("Fetching page", "page", page)
		posts, err := CollectPostsFromPage(cfg, fanclubID, page, cookieString)
		if err != nil {
			slog.Error("Failed to fetch page", "page", page, "error", err)
			continue
		}

		for _, post := range posts {
			filePath := filepath.Join(outputDir, post.Title+".md")
			// 如果文件已存在则跳过
			if _, err := os.Stat(filePath); err == nil {
				slog.Debug("Post already exists, skipping", "title", post.Title)
				continue
			}

			slog.Info("Downloading post", "title", post.Title)
			htmlContent, err := GetPostContent(cfg, post.Url, cookieString)
			if err != nil {
				slog.Error("Failed to get post content", "url", post.Url, "error", err)
				continue
			}

			markdown, err := converter.ConvertString(htmlContent)
			if err != nil {
				slog.Error("Failed to convert HTML to Markdown", "title", post.Title, "error", err)
				markdown = htmlContent // 降级保存 HTML
			}

			// 写入文件
			header := fmt.Sprintf("# %s\n\nURL: %s\n\n---\n\n", post.Title, post.Url)
			if err := os.WriteFile(filePath, []byte(header+markdown), 0644); err != nil {
				slog.Error("Failed to save file", "file", filePath, "error", err)
			}

			time.Sleep(time.Duration(DelayMs) * time.Millisecond)
		}

		time.Sleep(time.Duration(DelayMs) * time.Millisecond)
	}

	return nil
}
