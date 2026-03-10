package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ImgDir = ".assets"
)

// ResolveAppDir 解析程序所在目录（可执行文件目录或工作目录）
func ResolveAppDir() (string, error) {
	ex, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(ex)
		if !strings.Contains(execDir, "go-build") {
			return execDir, nil
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return wd, nil
}

// DefaultDataDir 返回默认的数据目录（appDir/data）
func DefaultDataDir(appDir string) string {
	return filepath.Join(appDir, "data")
}

// DefaultCookiePath 返回默认的 cookie 文件路径（appDir/cookies.json）
func DefaultCookiePath(appDir string) string {
	return filepath.Join(appDir, "cookies.json")
}

// DefaultLogPath 返回默认的日志文件路径（appDir/crawler.log）
func DefaultLogPath(appDir string) string {
	return filepath.Join(appDir, "crawler.log")
}

// GetExecutionTime 计算执行耗时并格式化
func GetExecutionTime(startTime, endTime time.Time) string {
	duration := endTime.Sub(startTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := duration.Seconds() - float64(hours*3600+minutes*60)

	result := ""
	if hours > 0 {
		result += fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dmin", minutes)
	}
	result += fmt.Sprintf("%.2fs", seconds)
	return result
}

// ToSafeFilename 将输入字符串转换为安全的文件名
func ToSafeFilename(in string) string {
	rp := strings.NewReplacer(
		"/", "_",
		`\`, "_",
		"<", "_",
		">", "_",
		":", "_",
		`"`, "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)
	return rp.Replace(in)
}

// UniqueStrings 字符串切片去重
func UniqueStrings(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)
	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}

// IsNumeric 检查字符串是否由纯数字组成且不为空
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
