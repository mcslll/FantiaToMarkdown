package fantia

import (
	"FantiaToMarkdown/config"
	"FantiaToMarkdown/utils"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slog"
)

// GetMaxPage 获取 Fanclub 的最大页数并提取 CSRF Token
func GetMaxPage(cfg *config.Config, fanclubID string, cookieString string) (int, string, error) {
	apiUrl := fmt.Sprintf("%s/fanclubs/%s/posts", cfg.HostUrl, fanclubID)
	body, err := NewRequestGet(cfg.Host, apiUrl, cookieString)
	if err != nil {
		return 1, "", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return 1, "", err
	}

	// 提取 CSRF Token
	csrfToken, _ := doc.Find("meta[name='csrf-token']").Attr("content")
	if csrfToken == "" {
		slog.Warn("CSRF token not found in page metadata")
	}

	maxPage := 1
	re := regexp.MustCompile(`page=(\d+)`)

	// 尝试寻找 fa-angle-double-right
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

	return maxPage, csrfToken, nil
}

// CollectPostsFromPage 抓取指定页面的帖子列表
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

// GetPostContent 从 Fantia API 获取帖子的详情和图片
func GetPostContent(cfg *config.Config, postUrl string, cookieString string, csrfToken string) (string, []string, error) {
	re := regexp.MustCompile(`/posts/(\d+)`)
	matches := re.FindStringSubmatch(postUrl)
	if len(matches) < 2 {
		return "", nil, fmt.Errorf("failed to parse post ID: %s", postUrl)
	}
	postID := matches[1]

	apiUrl := fmt.Sprintf("%s/api/v1/posts/%s", cfg.HostUrl, postID)

	extraHeaders := map[string]string{
		"X-Requested-With": "XMLHttpRequest",
		"X-CSRF-Token":     csrfToken,
		"Referer":          postUrl,
	}

	body, err := NewRequestGet(cfg.Host, apiUrl, cookieString, extraHeaders)
	if err != nil {
		return "", nil, err
	}

	jsonData := gjson.ParseBytes(body)
	postJson := jsonData.Get("post")
	if !postJson.Exists() {
		return "", nil, fmt.Errorf("API response invalid for post %s", postID)
	}

	// 1. 提取文本内容
	contentHtml := postJson.Get("comment").String()
	if contentHtml == "" {
		contentHtml = postJson.Get("description").String()
	}

	var pictures []string

	// 2. 提取封面图 (thumb)
	if thumb := postJson.Get("thumb.original"); thumb.Exists() {
		pictures = append(pictures, thumb.String())
	} else if thumb := postJson.Get("thumb.main"); thumb.Exists() {
		pictures = append(pictures, thumb.String())
	}

	// 3. 提取内容区块中的图片 (photo_gallery 类型)
	postJson.Get("post_contents").ForEach(func(key, value gjson.Result) bool {
		// 遍历照片数组
		value.Get("post_content_photos").ForEach(func(k, v gjson.Result) bool {
			// 优先取原图地址
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

		// 4. 处理 Blog 类型的嵌套图片 (Quill Delta 格式)
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

	// 去重图片列表
	pictures = utils.UniqueStrings(pictures)

	slog.Debug("Post parsing completed via API", "id", postID, "imagesCount", len(pictures))
	return contentHtml, pictures, nil
}
