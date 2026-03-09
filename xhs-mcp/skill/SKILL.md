---
name: xiaohongshu-mcp
description: 使用小红书 MCP 服务器进行内容发布、浏览、搜索和互动。支持图文/视频发布、Feed浏览、用户互动等功能。
---

# 小红书 MCP (Xiaohongshu MCP)

## 概述

小红书 MCP 是一个基于 Chrome CDP 自动化的小红书内容管理工具，支持通过 MCP 协议与 Claude 集成。

**核心功能:**
- 图文/视频内容发布
- Feed 浏览与搜索
- 用户互动（点赞、收藏、评论）
- 用户主页查看

**使用前声明:** "我正在使用小红书 MCP skill 来帮助你管理小红书内容。"

---

## 前置要求

### 1. 系统要求
- macOS 或 Linux
- 已安装 Chrome 浏览器
- Go 1.21+ (如需从源码构建)

### 2. 配置文件

确保 cookies 文件存在（可以为空对象 `{}`）：

```bash
mkdir -p cookies
echo '{}' > cookies/cookies.json
```

### 3. 二进制文件

检查二进制文件是否存在：

```bash
ls xiaohongshu-mcp 2>/dev/null || echo "需要构建"
```

如不存在，需要构建：

```bash
go build -o xiaohongshu-mcp
```

---

## 服务启动

### 启动 MCP 服务

```bash
./xiaohongshu-mcp --port 18060 --log-level info
```

**常用参数:**
- `--port`: 服务端口（默认 18060）
- `--log-level`: 日志级别（debug/info/warn/error）
- `--headless`: 是否无头模式（默认 true）

### 后台运行

```bash
# 后台启动
nohup ./xiaohongshu-mcp --port 18060 > mcp.log 2>&1 &

# 检查是否运行
lsof -i :18060

# 停止服务
pkill -f xiaohongshu-mcp
```

### 检查服务状态

```bash
curl http://localhost:18060/health
```

---

## Claude Code 配置

### 添加 MCP 服务器

编辑 Claude Code 配置：

```bash
# macOS
vim ~/Library/Application\ Support/Claude/settings.json

# 或 Linux
vim ~/.config/Claude/settings.json
```

添加以下配置（注意替换为实际路径）：

```json
{
  "mcpServers": {
    "xiaohongshu": {
      "command": "/ABSOLUTE/PATH/TO/xiaohongshu-mcp",
      "args": ["--port", "18060", "--log-level", "info"],
      "env": {},
      "cwd": "/ABSOLUTE/PATH/TO/xiaohongshu-mcp/DIRECTORY"
    }
  }
}
```

**配置示例:**
```json
{
  "mcpServers": {
    "xiaohongshu": {
      "command": "/Users/getui/Desktop/repo/xhs-mcp/mcp/xiaohongshu-mcp",
      "args": ["--port", "18060", "--log-level", "info"],
      "env": {},
      "cwd": "/Users/getui/Desktop/repo/xhs-mcp/mcp"
    }
  }
}
```

### 验证连接

配置完成后，重启 Claude Code 或在设置中刷新 MCP 服务器，应该能看到可用工具列表。

---

## HTTP API 直接调用

除了 MCP 工具，服务也支持直接 HTTP API 调用，便于调试和脚本使用。

### 基础信息
- **基础 URL**: `http://localhost:18060`
- **API 前缀**: `/api/v1`

### API 端点

#### 检查登录状态
```bash
curl http://localhost:18060/api/v1/login/status
```

#### 获取登录二维码
```bash
curl http://localhost:18060/api/v1/login/qrcode_image
# 返回二维码图片 base64
```

#### 搜索内容
```bash
curl -X POST http://localhost:18060/api/v1/feeds/search \
  -H "Content-Type: application/json" \
  -d '{
    "keyword": "关键词",
    "type": "user",
    "page": 1,
    "page_size": 10,
    "filters": []
  }'
```

#### 获取用户主页
```bash
curl -X POST http://localhost:18060/api/v1/user/profile \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "用户ID",
    "xsec_token": "xsec_token"
  }'
```

#### 获取笔记详情
```bash
curl -X POST http://localhost:18060/api/v1/feeds/detail \
  -H "Content-Type: application/json" \
  -d '{
    "feed_id": "笔记ID",
    "xsec_token": "xsec_token"
  }'
```

---

## MCP 工具功能说明

### 1. 登录相关

#### check_login_status
检查当前登录状态。

**使用场景:**
- 开始操作前确认是否已登录
- 排查登录问题

#### get_login_qrcode
获取登录二维码。

**使用场景:**
- 首次登录
- Cookie 过期需要重新登录

**工作流程:**
1. 调用工具获取二维码
2. 用小红书 App 扫描二维码
3. 等待登录完成
4. Cookie 会自动保存到 `cookies/cookies.json`

**登录持久性:**
- Cookie 有效期约 **1 年**
- 关键凭证（web_session、id_token）在 2027-03 左右过期
- 主动退出、修改密码、安全检测可能导致提前失效

---

### 2. 内容发布

#### publish_content
发布图文笔记。

**参数:**
- `title`: 标题（最多20个中文字）
- `content`: 正文内容（不含标签）
- `images`: 图片路径列表（支持本地路径或 HTTP URL）
- `tags`: 话题标签列表（可选）

**图片支持格式:**
- 本地绝对路径: `/Users/user/Desktop/image.jpg`
- HTTP URL: `https://example.com/image.jpg`（自动下载）

**使用示例:**
```
发布一篇关于旅行的图文笔记，标题是"周末的杭州之旅"，
图片在 /Users/user/Desktop/hangzhou1.jpg 和 /Users/user/Desktop/hangzhou2.jpg，
标签是 [旅行, 杭州, 周末]
```

#### publish_with_video
发布视频笔记。

**参数:**
- `title`: 标题
- `content`: 正文内容
- `video`: 本地视频绝对路径
- `tags`: 话题标签列表（可选）

**限制:**
- 仅支持本地视频文件
- 单个视频

---

### 3. Feed 浏览

#### list_feeds
获取首页推荐 Feed 列表。

**返回:**
- 笔记列表（包含标题、作者、互动数据）
- 每个笔记的 `feed_id` 和 `xsec_token`（用于后续操作）

#### search_feeds
搜索小红书内容。

**参数:**
- `keyword`: 搜索关键词
- `type`: 搜索类型（`general` 综合 / `user` 用户）
- `page`: 页码（从1开始）
- `page_size`: 每页数量
- `filters`: 筛选条件（可选，传 `[]`）

**使用示例:**
```
搜索关键词"旅行"，类型为综合，第1页，每页20条
```

#### get_feed_detail
获取笔记详情。

**参数:**
- `feed_id`: 笔记 ID（从 list_feeds 或 search_feeds 获取）
- `xsec_token`: 访问令牌（从列表获取）

**返回:**
- 完整笔记内容
- 图片/视频列表
- 作者信息
- 互动数据（点赞/收藏/分享数）
- 评论列表

#### user_profile
获取用户主页。

**参数:**
- `user_id`: 用户 ID
- `xsec_token`: 访问令牌（可从搜索结果获取）

**返回:**
- 用户基本信息
- 关注/粉丝/获赞数
- 用户发布的笔记列表（自动滚动加载，最多约150-220篇）

---

### 4. 互动功能

#### like_feed
点赞或取消点赞笔记。

**参数:**
- `feed_id`: 笔记 ID
- `xsec_token`: 访问令牌
- `unlike`: true 表示取消点赞，false/省略表示点赞

#### favorite_feed
收藏或取消收藏笔记。

**参数:**
- `feed_id`: 笔记 ID
- `xsec_token`: 访问令牌
- `unfavorite`: true 表示取消收藏

#### post_comment_to_feed
发表评论。

**参数:**
- `feed_id`: 笔记 ID
- `xsec_token`: 访问令牌
- `content`: 评论内容

---

## 常见工作流程

### 发布图文笔记

```
1. check_login_status - 确认已登录
2. publish_content - 发布内容
3. 等待发布完成
```

### 浏览并互动

```
1. list_feeds - 获取推荐列表
2. get_feed_detail - 查看感兴趣的笔记详情
3. like_feed 或 favorite_feed - 点赞/收藏
4. post_comment_to_feed - 发表评论
```

### 搜索并查看用户

```
1. search_feeds - 搜索关键词（type=user）
2. 从结果中获取 user_id 和 xsec_token
3. user_profile - 获取用户主页和所有笔记
```

### 数据分析示例

```bash
# 1. 搜索用户获取 ID
curl -X POST http://localhost:18060/api/v1/feeds/search \
  -H "Content-Type: application/json" \
  -d '{"keyword":"用户名","type":"user","page":1,"page_size":10,"filters":[]}'

# 2. 获取用户完整资料
curl -X POST http://localhost:18060/api/v1/user/profile \
  -H "Content-Type: application/json" \
  -d '{"user_id":"用户ID","xsec_token":"xsec_token"}'
```

---

## 注意事项

### 频率限制
- 避免短时间内大量操作
- 模拟人类操作间隔（建议至少 2-3 秒）
- 批量操作时建议添加延迟

### Cookie 管理
- Cookie 自动保存到 `cookies/cookies.json`
- Cookie 有效期约 1 年，但可能因安全策略提前失效
- 不要手动修改 cookie 文件
- Cookie 文件不要提交到 Git（已添加到 .gitignore）

### 标题限制
- 最多 20 个中文字或英文单词
- 超出会被截断

### 图片处理
- 网络图片自动下载
- 本地图片需要绝对路径
- 建议图片尺寸符合小红书要求（3:4 比例最佳）

### 用户主页限制
- 用户主页笔记通过滚动加载获取
- 实际获取数量取决于页面加载逻辑（通常 150-220 篇）
- 小红书 Web 端本身有展示限制

### 错误处理
- 如遇到错误，先检查登录状态
- 查看服务端日志获取详细信息
- 二维码登录需要 App 配合
- 429 错误表示请求过于频繁，请稍后重试

---

## 故障排除

### 无法启动 MCP

**检查端口占用:**
```bash
lsof -i :18060
```

**手动启动测试:**
```bash
./xiaohongshu-mcp --port 18060 --log-level debug
```

**查看日志:**
```bash
tail -f mcp.log
```

### 登录状态失效

1. 调用 `get_login_qrcode` 或访问 `http://localhost:18060/api/v1/login/qrcode_image`
2. 用小红书 App 扫码
3. 重新尝试操作

**检查当前登录状态:**
```bash
curl http://localhost:18060/api/v1/login/status
```

### 发布失败

常见原因:
- 未登录或登录过期
- 图片路径错误
- 标题超过 20 字限制
- 网络问题
- Chrome 浏览器未安装

**排查步骤:**
1. 检查登录状态
2. 确认图片路径可访问
3. 检查标题长度
4. 查看详细错误信息
5. 检查 Chrome 是否安装: `google-chrome --version` 或 `/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --version`

### API 返回 404

- 确认服务已启动：`curl http://localhost:18060/health`
- 确认使用正确端口（默认 18060）
- 确认 API 路径正确（以 `/api/v1` 开头）

---

## 安全提醒

- Cookie 文件包含登录凭证，不要提交到 Git
- 仅在自己的账号上使用
- 遵守小红书平台规则
- 避免频繁操作导致账号受限
- 不要在公共环境分享二维码图片

---

## 更新日志

### v2.1.0
- 添加 HTTP API 直接调用支持
- 修正默认端口为 18060
- 添加登录持久性说明（约1年）
- 添加服务启动/停止说明

### v2.0.0
- 使用官方 MCP Go SDK
- 添加 11 个 MCP 工具
- 支持图文/视频发布
- 支持 Feed 浏览、搜索、互动
- 支持用户主页查看（自动滚动加载）
