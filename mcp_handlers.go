package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ajia1206/xhs-mcp/xiaohongshu"
	"github.com/sirupsen/logrus"
)

// mcpLogEntry 创建 MCP 操作的日志条目
func mcpLogEntry(toolName string) *logrus.Entry {
	return logrus.WithField("mcp_tool", toolName)
}

// formatJSON 将数据格式化为 JSON 字符串
func formatJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// handleCheckLoginStatus 处理检查登录状态
func (s *AppServer) handleCheckLoginStatus(ctx context.Context) *MCPToolResult {
	logger := mcpLogEntry("check_login_status")
	logger.Info("执行检查登录状态")

	status, err := s.xiaohongshuService.CheckLoginStatus(ctx)
	if err != nil {
		logger.WithError(err).Error("检查登录状态失败")
		return newErrorResult("检查登录状态失败: " + err.Error())
	}

	return newSuccessResult(fmt.Sprintf("登录状态检查成功: 是否登录=%v, 用户名=%s", status.IsLoggedIn, status.Username))
}

// handleGetLoginQrcode 处理获取登录二维码请求
func (s *AppServer) handleGetLoginQrcode(ctx context.Context) *MCPToolResult {
	logger := mcpLogEntry("get_login_qrcode")
	logger.Info("执行获取登录二维码")

	result, err := s.xiaohongshuService.GetLoginQrcode(ctx)
	if err != nil {
		logger.WithError(err).Error("获取登录二维码失败")
		return newErrorResult("获取登录二维码失败: " + err.Error())
	}

	if result.IsLoggedIn {
		return newSuccessResult("你当前已处于登录状态")
	}

	// 构建响应内容
	contents := []MCPContent{
		{Type: "text", Text: fmt.Sprintf("请用小红书 App 扫码登录，二维码有效时间: %s", result.Timeout)},
	}

	if result.Img != "" {
		contents = append(contents, MCPContent{
			Type:     "image",
			MimeType: "image/png",
			Data:     result.Img,
		})
	}

	return &MCPToolResult{Content: contents}
}

// handlePublishContent 处理发布内容
func (s *AppServer) handlePublishContent(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logger := mcpLogEntry("publish_content")

	// 解析参数
	req, err := parsePublishRequest(args)
	if err != nil {
		logger.WithError(err).Warn("参数解析失败")
		return newErrorResult("参数错误: " + err.Error())
	}

	logger.Infof("发布内容 - 标题: %s, 图片数量: %d, 标签数量: %d", req.Title, len(req.Images), len(req.Tags))

	// 执行发布
	result, err := s.xiaohongshuService.PublishContent(ctx, req)
	if err != nil {
		logger.WithError(err).Error("发布内容失败")
		return newErrorResult("发布失败: " + err.Error())
	}

	return newSuccessResult(fmt.Sprintf("内容发布成功! 标题: %s, 图片数: %d, 状态: %s", result.Title, result.Images, result.Status))
}

// parsePublishRequest 解析发布请求参数
func parsePublishRequest(args map[string]interface{}) (*PublishRequest, error) {
	title, _ := args["title"].(string)
	content, _ := args["content"].(string)

	var imagePaths []string
	if images, ok := args["images"].([]interface{}); ok {
		for _, path := range images {
			if pathStr, ok := path.(string); ok {
				imagePaths = append(imagePaths, pathStr)
			}
		}
	}

	var tags []string
	if tagList, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tagList {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	if title == "" {
		return nil, fmt.Errorf("标题不能为空")
	}
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("至少需要一张图片")
	}

	return &PublishRequest{
		Title:   title,
		Content: content,
		Images:  imagePaths,
		Tags:    tags,
	}, nil
}

// handlePublishVideo 处理发布视频内容
func (s *AppServer) handlePublishVideo(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logger := mcpLogEntry("publish_video")

	title, _ := args["title"].(string)
	content, _ := args["content"].(string)
	videoPath, _ := args["video"].(string)

	if videoPath == "" {
		return newErrorResult("发布失败: 缺少本地视频文件路径")
	}

	var tags []string
	if tagList, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tagList {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	logger.Infof("发布视频 - 标题: %s, 标签数量: %d", title, len(tags))

	req := &PublishVideoRequest{
		Title:   title,
		Content: content,
		Video:   videoPath,
		Tags:    tags,
	}

	result, err := s.xiaohongshuService.PublishVideo(ctx, req)
	if err != nil {
		logger.WithError(err).Error("发布视频失败")
		return newErrorResult("发布失败: " + err.Error())
	}

	return newSuccessResult(fmt.Sprintf("视频发布成功! 标题: %s, 视频: %s, 状态: %s", result.Title, result.Video, result.Status))
}

// handleListFeeds 处理获取Feeds列表
func (s *AppServer) handleListFeeds(ctx context.Context) *MCPToolResult {
	logger := mcpLogEntry("list_feeds")
	logger.Info("执行获取 Feeds 列表")

	result, err := s.xiaohongshuService.ListFeeds(ctx)
	if err != nil {
		logger.WithError(err).Error("获取 Feeds 列表失败")
		return newErrorResult("获取 Feeds 列表失败: " + err.Error())
	}

	jsonStr, err := formatJSON(result)
	if err != nil {
		return newErrorResult("序列化结果失败: " + err.Error())
	}

	return newSuccessResult(jsonStr)
}

// handleSearchFeeds 处理搜索Feeds
func (s *AppServer) handleSearchFeeds(ctx context.Context, args SearchFeedsArgs) *MCPToolResult {
	logger := mcpLogEntry("search_feeds")

	if args.Keyword == "" {
		return newErrorResult("搜索失败: 缺少关键词参数")
	}

	logger.Infof("搜索 Feeds - 关键词: %s, 筛选条件数量: %d", args.Keyword, len(args.Filters))

	var filters []xiaohongshu.FilterOption
	for _, filter := range args.Filters {
		filterOption, err := xiaohongshu.NewFilterOption(xiaohongshu.GetFilterGroupIndex(filter.FiltersIndex), filter.TagsIndex)
		if err != nil {
			return newErrorResult(fmt.Sprintf("筛选条件错误: 组=%v, 标签=%v, 错误=%v", filter.FiltersIndex, filter.TagsIndex, err))
		}
		filters = append(filters, filterOption)
	}

	result, err := s.xiaohongshuService.SearchFeeds(ctx, args.Keyword, filters...)
	if err != nil {
		logger.WithError(err).Error("搜索 Feeds 失败")
		return newErrorResult("搜索失败: " + err.Error())
	}

	jsonStr, err := formatJSON(result)
	if err != nil {
		return newErrorResult("序列化结果失败: " + err.Error())
	}

	return newSuccessResult(jsonStr)
}

// handleGetFeedDetail 处理获取Feed详情
func (s *AppServer) handleGetFeedDetail(ctx context.Context, args map[string]any) *MCPToolResult {
	logger := mcpLogEntry("get_feed_detail")

	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return newErrorResult("获取 Feed 详情失败: 缺少 feed_id 参数")
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return newErrorResult("获取 Feed 详情失败: 缺少 xsec_token 参数")
	}

	logger.Infof("获取 Feed 详情 - Feed ID: %s", feedID)

	result, err := s.xiaohongshuService.GetFeedDetail(ctx, feedID, xsecToken)
	if err != nil {
		logger.WithError(err).Error("获取 Feed 详情失败")
		return newErrorResult("获取 Feed 详情失败: " + err.Error())
	}

	jsonStr, err := formatJSON(result)
	if err != nil {
		return newErrorResult("序列化结果失败: " + err.Error())
	}

	return newSuccessResult(jsonStr)
}

// handleUserProfile 获取用户主页
func (s *AppServer) handleUserProfile(ctx context.Context, args map[string]any) *MCPToolResult {
	logger := mcpLogEntry("user_profile")

	userID, ok := args["user_id"].(string)
	if !ok || userID == "" {
		return newErrorResult("获取用户主页失败: 缺少 user_id 参数")
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return newErrorResult("获取用户主页失败: 缺少 xsec_token 参数")
	}

	logger.Infof("获取用户主页 - User ID: %s", userID)

	result, err := s.xiaohongshuService.UserProfile(ctx, userID, xsecToken)
	if err != nil {
		logger.WithError(err).Error("获取用户主页失败")
		return newErrorResult("获取用户主页失败: " + err.Error())
	}

	jsonStr, err := formatJSON(result)
	if err != nil {
		return newErrorResult("序列化结果失败: " + err.Error())
	}

	return newSuccessResult(jsonStr)
}

// handleLikeFeed 处理点赞/取消点赞
func (s *AppServer) handleLikeFeed(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logger := mcpLogEntry("like_feed")

	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return newErrorResult("操作失败: 缺少 feed_id 参数")
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return newErrorResult("操作失败: 缺少 xsec_token 参数")
	}

	unlike, _ := args["unlike"].(bool)

	var res *ActionResult
	var err error

	if unlike {
		logger.Infof("取消点赞 - Feed ID: %s", feedID)
		res, err = s.xiaohongshuService.UnlikeFeed(ctx, feedID, xsecToken)
	} else {
		logger.Infof("点赞 - Feed ID: %s", feedID)
		res, err = s.xiaohongshuService.LikeFeed(ctx, feedID, xsecToken)
	}

	if err != nil {
		action := "点赞"
		if unlike {
			action = "取消点赞"
		}
		logger.WithError(err).Errorf("%s失败", action)
		return newErrorResult(action + "失败: " + err.Error())
	}

	return newSuccessResult(res.Message)
}

// handleFavoriteFeed 处理收藏/取消收藏
func (s *AppServer) handleFavoriteFeed(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logger := mcpLogEntry("favorite_feed")

	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return newErrorResult("操作失败: 缺少 feed_id 参数")
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return newErrorResult("操作失败: 缺少 xsec_token 参数")
	}

	unfavorite, _ := args["unfavorite"].(bool)

	var res *ActionResult
	var err error

	if unfavorite {
		logger.Infof("取消收藏 - Feed ID: %s", feedID)
		res, err = s.xiaohongshuService.UnfavoriteFeed(ctx, feedID, xsecToken)
	} else {
		logger.Infof("收藏 - Feed ID: %s", feedID)
		res, err = s.xiaohongshuService.FavoriteFeed(ctx, feedID, xsecToken)
	}

	if err != nil {
		action := "收藏"
		if unfavorite {
			action = "取消收藏"
		}
		logger.WithError(err).Errorf("%s失败", action)
		return newErrorResult(action + "失败: " + err.Error())
	}

	return newSuccessResult(res.Message)
}

// handlePostComment 处理发表评论到Feed
func (s *AppServer) handlePostComment(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logger := mcpLogEntry("post_comment")

	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return newErrorResult("发表评论失败: 缺少 feed_id 参数")
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return newErrorResult("发表评论失败: 缺少 xsec_token 参数")
	}

	content, ok := args["content"].(string)
	if !ok || content == "" {
		return newErrorResult("发表评论失败: 缺少 content 参数")
	}

	logger.Infof("发表评论 - Feed ID: %s, 内容长度: %d", feedID, len(content))

	result, err := s.xiaohongshuService.PostCommentToFeed(ctx, feedID, xsecToken, content)
	if err != nil {
		logger.WithError(err).Error("发表评论失败")
		return newErrorResult("发表评论失败: " + err.Error())
	}

	return newSuccessResult(result.Message)
}

// newSuccessResult 创建成功的 MCP 结果
func newSuccessResult(text string) *MCPToolResult {
	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: text,
		}},
		IsError: false,
	}
}

// newErrorResult 创建错误的 MCP 结果
func newErrorResult(text string) *MCPToolResult {
	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: text,
		}},
		IsError: true,
	}
}
