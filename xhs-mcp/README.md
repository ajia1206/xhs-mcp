# 小红书 MCP (Model Context Protocol)

小红书 MCP 服务器，支持通过 MCP 协议与小红书进行交互，包括发布图文/视频、搜索内容、获取笔记详情、用户主页、点赞收藏等功能。

> **项目说明**: 本项目最初基于 [xpzouying/xiaohongshu-mcp](https://github.com/xpzouying/xiaohongshu-mcp) 进行深度优化和重构，修复了多项关键问题并增强了功能。

## 功能特性

- **登录管理**：扫码登录、登录状态检查
- **内容发布**：支持发布图文笔记和本地视频
- **内容浏览**：获取首页 Feed 列表、搜索内容
- **笔记详情**：获取笔记完整信息、评论列表
- **用户主页**：查看用户信息、关注/粉丝数、历史笔记
- **互动功能**：点赞/取消点赞、收藏/取消收藏、发表评论

## 快速开始

### 环境要求

- Go 1.20+
- Chrome/Chromium 浏览器（或指定浏览器二进制路径）

### 安装

```bash
# 克隆仓库
git clone https://github.com/ajia1206/xhs-mcp.git
cd xhs-mcp

# 安装依赖
go mod download

# 构建
go build -o xiaohongshu-mcp ./build/xiaohongshu-mcp
```

### 运行

```bash
# 基本运行
./xiaohongshu-mcp

# 指定浏览器路径（推荐）
./xiaohongshu-mcp -bin "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

# 非无头模式（显示浏览器窗口，用于调试）
./xiaohongshu-mcp -headless=false

# 指定端口
./xiaohongshu-mcp -port :8080

# 设置日志级别
./xiaohongshu-mcp -log-level debug

# 查看版本
./xiaohongshu-mcp -version
```

## 配置

### 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-headless` | `true` | 是否使用无头模式 |
| `-bin` | `""` | 浏览器二进制文件路径 |
| `-port` | `:18060` | 服务端口 |
| `-log-level` | `info` | 日志级别 (debug/info/warn/error) |
| `-version` | - | 显示版本信息 |

### 环境变量

| 变量名 | 说明 |
|--------|------|
| `ROD_BROWSER_BIN` | 浏览器二进制文件路径 |
| `XHS_MCP_PORT` | 服务端口 |
| `XHS_MCP_LOG_LEVEL` | 日志级别 |
| `XHS_MCP_HEADLESS` | 是否无头模式 (true/false) |

**配置优先级**：命令行参数 > 环境变量 > 默认值

## API 接口

### HTTP API

服务启动后，可通过以下接口访问：

- `GET /health` - 健康检查
- `GET /api/v1/login/status` - 检查登录状态
- `GET /api/v1/login/qrcode` - 获取登录二维码
- `GET /api/v1/login/qrcode_image` - 获取登录二维码图片
- `POST /api/v1/publish` - 发布图文内容
- `POST /api/v1/publish_video` - 发布视频
- `GET /api/v1/feeds/list` - 获取 Feed 列表
- `GET/POST /api/v1/feeds/search` - 搜索内容
- `POST /api/v1/feeds/detail` - 获取笔记详情
- `POST /api/v1/feeds/comment` - 发表评论
- `POST /api/v1/user/profile` - 获取用户主页
- `GET /api/v1/user/me` - 获取当前登录用户信息

### MCP 端点

- `/mcp` - MCP 协议端点（支持 Streamable HTTP）

## MCP 工具

连接 MCP 服务器后，可以使用以下工具：

| 工具名 | 描述 |
|--------|------|
| `check_login_status` | 检查小红书登录状态 |
| `get_login_qrcode` | 获取登录二维码 |
| `publish_content` | 发布图文内容 |
| `publish_with_video` | 发布视频内容 |
| `list_feeds` | 获取首页 Feed 列表 |
| `search_feeds` | 搜索小红书内容 |
| `get_feed_detail` | 获取笔记详情 |
| `user_profile` | 获取用户主页 |
| `post_comment_to_feed` | 发表评论 |
| `like_feed` | 点赞/取消点赞 |
| `favorite_feed` | 收藏/取消收藏 |

## 项目结构

```
xhs-mcp/
├── build/xiaohongshu-mcp/    # 主程序代码
│   ├── main.go               # 程序入口
│   ├── service.go            # 业务服务层
│   ├── mcp_handlers.go       # MCP 工具处理函数
│   ├── mcp_server.go         # MCP 服务器初始化
│   ├── handlers_api.go       # HTTP API 处理函数
│   ├── app_server.go         # 应用服务器
│   ├── routes.go             # 路由配置
│   ├── middleware.go         # 中间件
│   └── types.go              # 类型定义
├── configs/                   # 配置管理
├── cookies/                   # Cookie 管理
├── browser/                   # 浏览器封装
├── pkg/downloader/            # 图片下载器
└── xiaohongshu/               # 小红书操作封装
```

## 技术栈

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [go-rod/rod](https://github.com/go-rod/rod) - 浏览器自动化
- [logrus](https://github.com/sirupsen/logren) - 日志库
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) - MCP 协议实现

## 优化改进

本项目在原始代码基础上进行了以下关键优化：

### 🔧 核心修复
- **Cookie 持久化修复** - 修复了 cookie expires 字段类型不匹配问题，确保登录状态能正确保存约 1 年
- **搜索功能重构** - 移除了有问题的网络监听代码，改用页面 JS 执行获取数据，提高稳定性
- **用户主页增强** - 添加滚动加载逻辑，可获取全部笔记（从 30 篇提升到 150-220 篇）

### 🚀 性能优化
- 重构服务层代码结构，提高可维护性
- 优化 MCP 处理器响应速度
- 改进浏览器实例管理

### ✨ 功能增强
- 完善 HTTP API 端点，支持直接 API 调用
- 添加详细的错误处理和日志记录
- 优化配置文件管理

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 注意事项

1. 首次使用需要扫码登录小红书
2. 登录状态会保存在 `cookies.json` 文件中
3. 发布内容需要遵守小红书社区规范
4. 建议指定浏览器路径以获得更稳定的表现
