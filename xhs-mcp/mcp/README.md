# 小红书 MCP (Model Context Protocol)

小红书 MCP 服务器，支持通过 MCP 协议与小红书进行交互，包括发布图文/视频、搜索内容、获取笔记详情、用户主页、点赞收藏等功能。

> 项目说明: 本项目最初基于 [xpzouying/xiaohongshu-mcp](https://github.com/xpzouying/xiaohongshu-mcp) 进行优化和重构。

## 功能特性

- 登录管理：扫码登录、登录状态检查
- 内容发布：支持发布图文笔记和本地视频
- 内容浏览：获取首页 Feed 列表、搜索内容
- 笔记详情：获取笔记完整信息、评论列表
- 用户主页：查看用户信息、关注/粉丝数、历史笔记
- 互动功能：点赞/取消点赞、收藏/取消收藏、发表评论

## 快速开始

### 环境要求

- Go 1.20+
- Chrome/Chromium 浏览器

### 安装

```bash
git clone https://github.com/ajia1206/xhs-mcp.git
cd xhs-mcp/mcp
go mod download
go build -o xiaohongshu-mcp .
```

### 运行

```bash
./xiaohongshu-mcp
./xiaohongshu-mcp -bin "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
./xiaohongshu-mcp -headless=false
./xiaohongshu-mcp -port :8080
./xiaohongshu-mcp -log-level debug
./xiaohongshu-mcp -version
```

## 配置

### 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-headless` | `true` | 是否使用无头模式 |
| `-bin` | `""` | 浏览器二进制文件路径 |
| `-port` | `:18060` | 服务端口 |
| `-log-level` | `info` | 日志级别 |
| `-version` | - | 显示版本信息 |

### 环境变量

| 变量名 | 说明 |
|--------|------|
| `ROD_BROWSER_BIN` | 浏览器二进制文件路径 |
| `XHS_MCP_PORT` | 服务端口 |
| `XHS_MCP_LOG_LEVEL` | 日志级别 |
| `XHS_MCP_HEADLESS` | 是否无头模式 |

## API 接口

### HTTP API

- `GET /health`
- `GET /api/v1/login/status`
- `GET /api/v1/login/qrcode`
- `GET /api/v1/login/qrcode_image`
- `POST /api/v1/publish`
- `POST /api/v1/publish_video`
- `GET /api/v1/feeds/list`
- `GET/POST /api/v1/feeds/search`
- `POST /api/v1/feeds/detail`
- `POST /api/v1/feeds/comment`
- `POST /api/v1/user/profile`
- `GET /api/v1/user/me`

### MCP 端点

- `/mcp`

## MCP 工具

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

```text
mcp/
├── main.go
├── service.go
├── mcp_handlers.go
├── mcp_server.go
├── handlers_api.go
├── app_server.go
├── routes.go
├── middleware.go
├── types.go
├── browser/
├── configs/
├── cookies/
├── errors/
├── pkg/downloader/
└── xiaohongshu/
```

## 注意事项

1. 首次使用需要扫码登录小红书。
2. 登录状态会保存在 `cookies.json` 或 `cookies/cookies.json` 中。
3. 发布内容需要遵守小红书社区规范。
4. 建议指定浏览器路径以获得更稳定的表现。
