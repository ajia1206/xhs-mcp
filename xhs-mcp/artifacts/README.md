# XHS 分析产物说明（2025-10-17 17:42）

## 文件清单（位于 `2025-10-17-1742_xhs-analysis/`）
- `xhs_samples.json`：8 个关键词共 173 条样本的汇总数据（含 `feed_id`、`xsecToken`、互动数据）
- `xhs_report.json` / `xhs_report.txt`：关键词级指标统计与观察结论
- `xhs_top_details.json`：点赞最高 15 篇笔记的完整详情（正文、评论、配图信息）
- `xhs_top_cta.json` / `xhs_top_cta.txt`：高热样本的 CTA / 结构模式统计

## 新增任务（2025-10-17 18:00 `2025-10-17-1800_ai-plus-analysis/`）
- `ai_plus_feeds.json`：以 “AI+行业” 为关键词抓取的 738 条笔记（去重后）
- `ai_plus_analysis.json`：方向标签统计与样本明细
- `ai_plus_summary.txt`：Top 10 AI+ 方向摘要

> 提示：后续分析请在 `artifacts/YYYY-MM-DD-HHmm_task/` 中查找对应数据，并更新本 README 以记录任务概况。

## 新增任务（2025-10-17 22:03 `2025-10-17-2203_ai-self-study/`）
- `self_study_feeds.json`：围绕 AI 自学/学习/知识库等关键词抓取的 253 条笔记
- `self_study_analysis.json`：含标准化互动数据、关键词汇总与高赞列表
- `self_study_summary.txt`：Top 15 高赞笔记与关键词覆盖概览

## 新增任务（2025-10-17 22:11 `2025-10-17-2211_ai-study-method-details/`）
- `top_feeds.json`：命中“AI 学习方法”且点赞最高的 10 条笔记
- `top_feed_details.json`：上述笔记的正文/配图/评论详情（MCP `get_feed_detail`）
- `top_feed_summary.json` / `top_feed_summary.txt`：CTA、结构、图文形态汇总及逐条摘要

## 新增任务（2025-10-17 22:16 `2025-10-17-2216_ai-self-study-template/`）
- `ai_self_study_workflow.md`：AI 自学研究的一站式流程模板与 Prompt 集
- `xhs_post_01_efficiency.md`：小红书稿件（效率倍增主题）大纲
- `xhs_post_02_self_coach.md`：小红书稿件（AI 自律打卡主题）大纲

## 新增任务（2025-10-18 15:03 `2025-10-18-1503_ai-self-study-deepdive/`）
- `deepdive_feeds.json`：AI 自学细分主题（学英语/编程/数学、考研、语言考试等）共 322 条笔记
- `deepdive_analysis.json`：标准化互动数据、Top40 高赞笔记、关键词统计
- `deepdive_summary.txt`：高赞笔记&关键词覆盖摘要

## 使用指南
1. **追踪高热笔记**：基于 `xhs_top_details.json` 的 `feed_id` 与 `xsecToken`，可调用 MCP `get_feed_detail` 更新互动数据。
2. **标题 / CTA 设计参考**：`xhs_top_cta.txt` 总结了高赞笔记的结构与 CTA 词频，可直接提炼标题模板。
3. **数据扩展**：如需追加关键词，可沿用 `xhs_samples.json` 的结构继续 append（建议保留 `keyword`、`likes_norm` 字段用于统计）。

## 备注
- 数据采样时间：2025-10-17 17:25～17:40 (UTC+8)
- MCP 二进制已修复 `search_feeds` / `list_feeds` 循环引用问题（替换于同日 17:20）
- 后续任务请在 `artifacts/YYYY-MM-DD-HHmm_task-name/` 下统一保存，并附类似 README 说明。
