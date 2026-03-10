package fantia

import (
	"FantiaToMarkdown/config"
	"FantiaToMarkdown/utils"
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slog"
)

// GetMaxPage 获取 Fanclub 的最大页数、提取 CSRF Token 并获取 Fanclub 名称
func GetMaxPage(cfg *config.Config, fanclubID string, tag string, cookieString string) (int, string, string, error) {
	apiUrl := fmt.Sprintf("%s/fanclubs/%s/posts", cfg.HostUrl, fanclubID)
	if tag != "" {
		apiUrl += "?tag=" + url.QueryEscape(tag)
	}

	body, err := NewRequestGet(cfg, apiUrl, cookieString)
	if err != nil {
		return 1, "", "", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return 1, "", "", err
	}

	// 1. 提取 CSRF Token
	csrfToken, _ := doc.Find("meta[name='csrf-token']").Attr("content")
	if csrfToken == "" {
		slog.Warn("CSRF token not found in page metadata")
	}

	// 2. 提取 Fanclub 名称
	fanclubName := doc.Find("h1.fanclub-name").Text()
	fanclubName = strings.TrimSpace(fanclubName)
	if fanclubName == "" {
		// 备选选择器
		fanclubName = doc.Find(".fanclub-header h1").Text()
		fanclubName = strings.TrimSpace(fanclubName)
	}

	maxPage := 1
	re := regexp.MustCompile(`page=(\d+)`)

	// 3. 尝试寻找最大页数
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

	// 兜底尝试所有 page-link
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

	return maxPage, csrfToken, fanclubName, nil
}

// CollectPostsFromPage 抓取指定页面的帖子列表
func CollectPostsFromPage(cfg *config.Config, fanclubID string, page int, tag string, cookieString string) ([]Post, error) {
	apiUrl := fmt.Sprintf("%s/fanclubs/%s/posts?page=%d", cfg.HostUrl, fanclubID, page)
	if tag != "" {
		apiUrl += "&tag=" + url.QueryEscape(tag)
	}

	body, err := NewRequestGet(cfg, apiUrl, cookieString)
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

// GetPostContent 从 Fantia API 获取帖子的详情（含日期和标签）
func GetPostContent(cfg *config.Config, postUrl string, cookieString string, csrfToken string) (Post, error) {
	re := regexp.MustCompile(`/posts/(\d+)`)
	matches := re.FindStringSubmatch(postUrl)
	if len(matches) < 2 {
		return Post{}, fmt.Errorf("failed to parse post ID: %s", postUrl)
	}
	postID := matches[1]

	apiUrl := fmt.Sprintf("%s/api/v1/posts/%s", cfg.HostUrl, postID)

	extraHeaders := map[string]string{
		"X-Requested-With": "XMLHttpRequest",
		"X-CSRF-Token":     csrfToken,
		"Referer":          postUrl,
	}

	body, err := NewRequestGet(cfg, apiUrl, cookieString, extraHeaders)
	if err != nil {
		return Post{}, err
	}

	jsonData := gjson.ParseBytes(body)
	postJson := jsonData.Get("post")
	if !postJson.Exists() {
		return Post{}, fmt.Errorf("API response invalid for post %s", postID)
	}

	// 1. 提取基本信息
	title := postJson.Get("title").String()
	postedAt := postJson.Get("posted_at").String()

	// 2. 提取文本内容
	contentHtml := postJson.Get("comment").String()
	if contentHtml == "" {
		contentHtml = postJson.Get("description").String()
	}

	// 3. 提取标签
	var tags []string
	postJson.Get("tags").ForEach(func(key, value gjson.Result) bool {
		tags = append(tags, value.Get("name").String())
		return true
	})

	// 4. 提取图片列表
	var pictures []string
	if thumb := postJson.Get("thumb.original"); thumb.Exists() {
		pictures = append(pictures, thumb.String())
	} else if thumb := postJson.Get("thumb.main"); thumb.Exists() {
		pictures = append(pictures, thumb.String())
	}

	postJson.Get("post_contents").ForEach(func(key, value gjson.Result) bool {
		value.Get("post_content_photos").ForEach(func(k, v gjson.Result) bool {
			imgUrl := ""
			if original := v.Get("url.original"); original.Exists() {
				imgUrl = original.String()
			} else if webp := v.Get("url.main_webp"); webp.Exists() {
				imgUrl = webp.String()
			} else {
				imgUrl = v.Get("url.main").String()
			}
			if imgUrl != "" {
				pictures = append(pictures, imgUrl)
			}
			return true
		})

		commentStr := value.Get("comment").String()
		if strings.HasPrefix(commentStr, "{\"ops\":") {
			opsJson := gjson.Parse(commentStr)
			opsJson.Get("ops").ForEach(func(k, v gjson.Result) bool {
				if img := v.Get("insert.fantiaImage.original_url"); img.Exists() {
					pictures = append(pictures, img.String())
				}
				return true
			})
		}
		return true
	})

	pictures = utils.UniqueStrings(pictures)

	slog.Debug("Post parsing completed via API", "id", postID, "imagesCount", len(pictures), "tags", tags)

	return Post{
		Title:    utils.ToSafeFilename(title),
		Url:      postUrl,
		Content:  contentHtml,
		Pictures: pictures,
		PostedAt: postedAt,
		Tags:     tags,
	}, nil
}
