package browser

import (
	"encoding/json"
	"time"

	"github.com/ajia1206/xhs-mcp/cookies"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// Browser 包装 rod.Browser，提供 cookie 管理
type Browser struct {
	browser *rod.Browser
	cookies []*proto.NetworkCookieParam
}

// NewBrowser 创建新的浏览器实例
func NewBrowser(headless bool, options ...Option) *Browser {
	cfg := &browserConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	l := launcher.New().
		Headless(headless).
		NoSandbox(true)

	if cfg.binPath != "" {
		l = l.Bin(cfg.binPath)
	}

	url := l.MustLaunch()

	rbrowser := rod.New().
		ControlURL(url).
		MustConnect()

	// 加载并设置 cookies（在浏览器级别）
	cookiePath := cookies.GetCookiesFilePath()
	cookieLoader := cookies.NewLoadCookie(cookiePath)

	var cookieParams []*proto.NetworkCookieParam

	if data, err := cookieLoader.LoadCookies(); err == nil {
		// 解析 cookie 文件
		var rawCookies []map[string]interface{}
		if err := json.Unmarshal(data, &rawCookies); err == nil {
			for _, c := range rawCookies {
				name, _ := c["name"].(string)
				value, _ := c["value"].(string)
				domain, _ := c["domain"].(string)
				path, _ := c["path"].(string)
				httpOnly, _ := c["httpOnly"].(bool)
				secure, _ := c["secure"].(bool)

				// 处理 expires 字段（proto.TimeSinceEpoch 是 float64）
				var expires proto.TimeSinceEpoch
				if exp, ok := c["expires"].(float64); ok {
					expires = proto.TimeSinceEpoch(exp)
				}

				if name != "" && domain != "" {
					cookieParams = append(cookieParams, &proto.NetworkCookieParam{
						Name:     name,
						Value:    value,
						Domain:   domain,
						Path:     path,
						HTTPOnly: httpOnly,
						Secure:   secure,
						Expires:  expires,
					})
				}
			}

			// 在浏览器级别设置 cookies
			// 需要将 NetworkCookieParam 转换为 NetworkCookie
			if len(cookieParams) > 0 {
				var networkCookies []*proto.NetworkCookie
				for _, cp := range cookieParams {
					networkCookies = append(networkCookies, &proto.NetworkCookie{
						Name:     cp.Name,
						Value:    cp.Value,
						Domain:   cp.Domain,
						Path:     cp.Path,
						HTTPOnly: cp.HTTPOnly,
						Secure:   cp.Secure,
					})
				}
				rbrowser.MustSetCookies(networkCookies...)
				logrus.Infof("[Browser] Loaded %d cookies", len(cookieParams))
			}
		} else {
			logrus.Warnf("[Browser] Failed to parse cookies: %v", err)
		}
	} else {
		logrus.Warnf("[Browser] Failed to load cookies: %v", err)
	}

	return &Browser{
		browser: rbrowser,
		cookies: cookieParams,
	}
}

// NewPage 创建新页面，并确保 cookies 生效
func (b *Browser) NewPage() *rod.Page {
	// 不使用 stealth 模式，直接创建页面
	page := b.browser.MustPage()

	// 在导航之前设置 cookies（关键！）
	if len(b.cookies) > 0 {
		page.Browser().SetCookies(b.cookies)
		logrus.Infof("[Browser] Set %d cookies before navigation", len(b.cookies))
	}

	// 访问首页
	page.MustNavigate("https://www.xiaohongshu.com")
	page.MustWaitLoad()
	time.Sleep(3 * time.Second)

	// 检查登录状态
	hasAvatar := page.MustEval(`() => document.querySelector('.avatar-img') !== null`).Bool()
	logrus.Infof("[Browser] NewPage - hasAvatar: %v", hasAvatar)

	return page
}

// Close 关闭浏览器
func (b *Browser) Close() {
	b.browser.MustClose()
}

type browserConfig struct {
	binPath string
}

type Option func(*browserConfig)

func WithBinPath(binPath string) Option {
	return func(c *browserConfig) {
		c.binPath = binPath
	}
}
