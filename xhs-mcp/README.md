# xhs-mcp Workspace

这个仓库现在按两个主要项目重新整理：

- `mcp/`：小红书 MCP 服务端源码（Go）
- `skill/`：配套的 Claude/Codex Skill 文档与安装说明

本次整理刻意没有处理现有的数据分析/可视化工具链，`scripts/`、`web/`、`data/`、`artifacts/` 等目录保持原样。

## 目录说明

```text
xhs-mcp/
├── mcp/                 # Go MCP 项目
├── skill/               # Skill 项目
├── scripts/             # 现有辅助脚本（含 xhs-ready）
├── web/                 # 现有可视化页面
├── data/                # 现有数据目录
└── artifacts/           # 现有分析产物
```

## 使用入口

- MCP 项目说明见 [mcp/README.md](/Users/getui/Desktop/repo/xhs-mcp/mcp/README.md)
- Skill 项目说明见 [skill/README.md](/Users/getui/Desktop/repo/xhs-mcp/skill/README.md)

## 快速命令

构建 MCP：

```bash
cd mcp
go build -o xiaohongshu-mcp .
```

启动就绪脚本：

```bash
scripts/xhs-ready.sh
```
