package cookies

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type Cookier interface {
	LoadCookies() ([]byte, error)
	SaveCookies(data []byte) error
}

type localCookie struct {
	path string
}

func NewLoadCookie(path string) Cookier {
	if path == "" {
		panic("path is required")
	}

	return &localCookie{
		path: path,
	}
}

// LoadCookies 从文件中加载 cookies。
func (c *localCookie) LoadCookies() ([]byte, error) {

	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cookies from tmp file")
	}

	return data, nil
}

// SaveCookies 保存 cookies 到文件中。
func (c *localCookie) SaveCookies(data []byte) error {
	dir := filepath.Dir(c.path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrap(err, "failed to create cookies directory")
		}
	}

	return os.WriteFile(c.path, data, 0644)
}

// GetCookiesFilePath 获取 cookies 文件路径。
// 优先使用显式配置的 COOKIES_PATH；
// 未配置时再兼容旧路径 /tmp/cookies.json，最后回退到当前目录下的 cookies.json。
func GetCookiesFilePath() string {
	path := os.Getenv("COOKIES_PATH")
	if path != "" {
		return path
	}

	// 旧路径：/tmp/cookies.json
	tmpDir := os.TempDir()
	oldPath := filepath.Join(tmpDir, "cookies.json")

	if _, err := os.Stat(oldPath); err == nil {
		return oldPath
	}

	return "cookies.json"
}
