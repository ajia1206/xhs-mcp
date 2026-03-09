package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
)

// 应用版本常量
const (
	Version = "2.0.0"
)

// Config 应用配置结构体
type Config struct {
	Headless bool
	BinPath  string
	Port     string
	LogLevel string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Headless: true,
		Port:     ":18060",
		LogLevel: "info",
	}
}

// LoadFromEnv 从环境变量加载配置（仅当值非空时设置）
func (c *Config) LoadFromEnv() {
	if v := os.Getenv("ROD_BROWSER_BIN"); v != "" {
		c.BinPath = v
	}
	if v := os.Getenv("XHS_MCP_PORT"); v != "" {
		c.Port = v
	}
	if v := os.Getenv("XHS_MCP_LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	if v := os.Getenv("XHS_MCP_HEADLESS"); v != "" {
		if headless, err := strconv.ParseBool(v); err == nil {
			c.Headless = headless
		}
	}
}

func main() {
	// 1. 加载默认配置
	cfg := DefaultConfig()

	// 2. 从环境变量加载（优先级：默认值 < 环境变量）
	cfg.LoadFromEnv()

	// 3. 设置命令行参数（会覆盖之前的值，优先级最高）
	flag.BoolVar(&cfg.Headless, "headless", cfg.Headless, "是否无头模式")
	flag.StringVar(&cfg.BinPath, "bin", cfg.BinPath, "浏览器二进制文件路径")
	flag.StringVar(&cfg.Port, "port", cfg.Port, "服务端口")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "日志级别 (debug/info/warn/error)")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本
	if *showVersion {
		fmt.Printf("xiaohongshu-mcp version %s\n", Version)
		os.Exit(0)
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.Warnf("无效的日志级别 %s，使用 info", cfg.LogLevel)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 初始化配置
	configs.InitHeadless(cfg.Headless)
	configs.SetBinPath(cfg.BinPath)

	logrus.Infof("🚀 启动小红书 MCP 服务 v%s", Version)
	logrus.Infof("配置: headless=%v, port=%s, log_level=%s", cfg.Headless, cfg.Port, cfg.LogLevel)

	// 初始化服务
	xiaohongshuService := NewXiaohongshuService()

	// 创建并启动应用服务器
	appServer := NewAppServer(xiaohongshuService)
	if err := appServer.Start(cfg.Port); err != nil {
		logrus.Fatalf("failed to run server: %v", err)
	}
}
