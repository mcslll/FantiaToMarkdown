package main

import (
	"FantiaToMarkdown/config"
	"FantiaToMarkdown/fantia"
	"FantiaToMarkdown/logger"
	"FantiaToMarkdown/storage"
	"FantiaToMarkdown/utils"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/urfave/cli/v3"
	"golang.org/x/exp/slog"
)

var (
	fantiaHost     string
	dataDirFlag    string
	cookiePathFlag string
	fanclubID      string
	debugMode      bool

	cfg *config.Config
)

// GetPosts 调度抓取逻辑：列表 -> 详情 -> 转换 -> 保存
func GetPosts(cfg *config.Config, fanclubID string, cookieString string) error {
	maxPage, csrfToken, err := fantia.GetMaxPage(cfg, fanclubID, cookieString)
	if err != nil {
		return err
	}
	slog.Info("Detected Max Page", "maxPage", maxPage)
	if csrfToken != "" {
		slog.Debug("CSRF token extracted", "token", csrfToken)
	}

	// 创建存储目录
	outputDir := filepath.Join(cfg.DataDir, fanclubID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	converter := md.NewConverter("", true, nil)

	for page := 1; page <= maxPage; page++ {
		slog.Info("Fetching page", "page", page)
		posts, err := fantia.CollectPostsFromPage(cfg, fanclubID, page, cookieString)
		if err != nil {
			slog.Error("Failed to fetch page", "page", page, "error", err)
			continue
		}

		for _, p := range posts {
			// 从 URL 提取 ID 用于文件名
			reID := regexp.MustCompile(`/posts/(\d+)`)
			postID := "0"
			if m := reID.FindStringSubmatch(p.Url); len(m) > 1 {
				postID = m[1]
			}

			fileName := fmt.Sprintf("%s_%s.md", postID, p.Title)
			filePath := filepath.Join(outputDir, fileName)

			// 如果文件已存在则跳过
			if _, err := os.Stat(filePath); err == nil {
				slog.Debug("Post already exists, skipping", "title", p.Title)
				continue
			}

			slog.Info("Downloading post", "title", p.Title)
			htmlContent, pics, err := fantia.GetPostContent(cfg, p.Url, cookieString, csrfToken)
			if err != nil {
				slog.Error("Failed to get post content", "url", p.Url, "error", err)
				continue
			}

			postData := fantia.Post{
				Title:    p.Title,
				Url:      p.Url,
				Content:  htmlContent,
				Pictures: pics,
			}

			// 调用 storage 模块保存
			if err := storage.SavePost(cfg, postData, outputDir, converter); err != nil {
				slog.Error("Failed to save post", "title", p.Title, "error", err)
			}

			time.Sleep(time.Duration(fantia.DelayMs) * time.Millisecond)
		}

		time.Sleep(time.Duration(fantia.DelayMs) * time.Millisecond)
	}

	return nil
}

func main() {
	// 记录开始时间
	startTime := time.Now()
	cmd := &cli.Command{
		Name:  "FantiaToMarkdown",
		Usage: "Fantia crawler in Go, inspired by AfdianToMarkdown",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "host", Destination: &fantiaHost, Value: "fantia.jp", Usage: "Fantia host"},
			&cli.StringFlag{Name: "dir", Destination: &dataDirFlag, Value: "", Usage: "Data directory"},
			&cli.StringFlag{Name: "cookie", Destination: &cookiePathFlag, Value: "", Usage: "Path to cookies.json"},
			&cli.BoolFlag{Name: "debug", Destination: &debugMode, Value: false, Usage: "Enable debug logging"},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// 设置日志级别
			logLevel := slog.LevelInfo
			if debugMode {
				logLevel = slog.LevelDebug
			}
			slog.SetDefault(logger.SetupLogger(logLevel))

			// 解析程序目录
			appDir, err := utils.ResolveAppDir()
			if err != nil {
				return ctx, fmt.Errorf("failed to resolve app directory: %w", err)
			}

			// 确定数据目录
			dataDir := dataDirFlag
			if dataDir == "" {
				dataDir = utils.DefaultDataDir(appDir)
			}

			// 确定 Cookie 路径
			cookiePath := cookiePathFlag
			if cookiePath == "" {
				cookiePath = utils.DefaultCookiePath(appDir)
			}

			cfg = config.NewConfig(fantiaHost, dataDir, cookiePath)
			return ctx, nil
		},
		After: func(ctx context.Context, cmd *cli.Command) error {
			// 记录结束时间
			endTime := time.Now()
			slog.Info("处理完毕", "time cost", utils.GetExecutionTime(startTime, endTime))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "fanclub",
				Usage: "抓取指定 Fanclub 的帖子",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "id", Destination: &fanclubID, Value: "", Usage: "Fanclub ID", Required: true},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cookieString, err := fantia.GetCookies(cfg.CookiePath)
					if err != nil {
						return err
					}
					return GetPosts(cfg, fanclubID, cookieString)
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// 如果没有子命令，显示帮助
			_ = cli.ShowAppHelp(cmd)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error())
	}
}
