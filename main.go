package main

import (
	"FantiaToMarkdown/config"
	"FantiaToMarkdown/fantia"
	"FantiaToMarkdown/logger"
	"FantiaToMarkdown/utils"
	"context"
	"fmt"
	"os"
	"time"

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
					return fantia.GetPosts(cfg, fanclubID, cookieString)
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
