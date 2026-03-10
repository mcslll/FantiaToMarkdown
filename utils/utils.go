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

func DefaultDataDir(appDir string) string {
	return filepath.Join(appDir, "data")
}

func DefaultCookiePath(appDir string) string {
	return filepath.Join(appDir, "cookies.json")
}

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
