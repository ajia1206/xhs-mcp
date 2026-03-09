package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ajia1206/xhs-mcp/browser"
	"github.com/ajia1206/xhs-mcp/configs"
	"github.com/ajia1206/xhs-mcp/cookies"
	"github.com/ajia1206/xhs-mcp/pkg/downloader"
	"github.com/ajia1206/xhs-mcp/xiaohongshu"
	"github.com/go-rod/rod"
	"github.com/mattn/go-runewidth"
	"github.com/sirupsen/logrus"
)

var (
	// ErrTitleTooLong 标题过长错误
	ErrTitleTooLong = errors.New("标题长度超过限制（最多40个字符宽度）")
	// ErrVideoNotFound 视频文件不存在
	ErrVideoNotFound = errors.New("视频文件不存在或不可访问")
	// ErrEmptyVideoPath 视频路径为空
	ErrEmptyVideoPath = errors.New("必须提供本地视频文件路径")
	// ErrNotLoggedIn 未登录错误
	ErrNotLoggedIn = errors.New("用户未登录")
)

// XiaohongshuService 小红书业务服务
type XiaohongshuService struct{}

// NewXiaohongshuService 创建小红书服务实例
func NewXiaohongshuService() *XiaohongshuService {
	return &XiaohongshuService{}
}

// PublishRequest 发布请求
type PublishRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Images  []string `json:"images" binding:"required,min=1"`
	Tags    []string `json:"tags,omitempty"`
}

// LoginStatusResponse 登录状态响应
type LoginStatusResponse struct {
	IsLoggedIn bool   `json:"is_logged_in"`
	Username   string `json:"username,omitempty"`
}

// LoginQrcodeResponse 登录扫码二维码
type LoginQrcodeResponse struct {
	Timeout    string `json:"timeout"`
	IsLoggedIn bool   `json:"is_logged_in"`
	Img        string `json:"img,omitempty"`
}

// PublishResponse 发布响应
type PublishResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Images  int    `json:"images"`
	Status  string `json:"status"`
	PostID  string `json:"post_id,omitempty"`
}

// PublishVideoRequest 发布视频请求（仅支持本地单个视频文件）
type PublishVideoRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Video   string   `json:"video" binding:"required"`
	Tags    []string `json:"tags,omitempty"`
}

// PublishVideoResponse 发布视频响应
type PublishVideoResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Video   string `json:"video"`
	Status  string `json:"status"`
	PostID  string `json:"post_id,omitempty"`
}

// FeedsListResponse Feeds列表响应
type FeedsListResponse struct {
	Feeds []xiaohongshu.Feed `json:"feeds"`
	Count int                `json:"count"`
}

// UserProfileResponse 用户主页响应
type UserProfileResponse struct {
	UserBasicInfo xiaohongshu.UserBasicInfo      `json:"userBasicInfo"`
	Interactions  []xiaohongshu.UserInteractions `json:"interactions"`
	Feeds         []xiaohongshu.Feed             `json:"feeds"`
}

// validateTitle 验证标题长度
func validateTitle(title string) error {
	// 小红书限制：最大40个单位长度
	// 中文/日文/韩文占2个单位，英文/数字占1个单位
	if titleWidth := runewidth.StringWidth(title); titleWidth > 40 {
		return fmt.Errorf("%w: 当前长度 %d", ErrTitleTooLong, titleWidth)
	}
	return nil
}

// withBrowser 创建浏览器上下文并执行操作
func withBrowser[T any](operation func(*rod.Page) (T, error)) (T, error) {
	var zero T

	b := newBrowser()
	defer func() {
		b.Close()
	}()

	page := b.NewPage()
	defer func() {
		page.Close()
	}()

	result, err := operation(page)
	if err != nil {
		return zero, err
	}
	return result, nil
}

// withBrowserNoResult 创建浏览器上下文并执行操作（无返回值）
func withBrowserNoResult(operation func(*rod.Page) error) error {
	_, err := withBrowser(func(page *rod.Page) (struct{}, error) {
		return struct{}{}, operation(page)
	})
	return err
}

// CheckLoginStatus 检查登录状态
func (s *XiaohongshuService) CheckLoginStatus(ctx context.Context) (*LoginStatusResponse, error) {
	result, err := withBrowser(func(page *rod.Page) (bool, error) {
		loginAction := xiaohongshu.NewLogin(page)
		return loginAction.CheckLoginStatus(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("检查登录状态失败: %w", err)
	}

	return &LoginStatusResponse{
		IsLoggedIn: result,
		Username:   configs.Username,
	}, nil
}

// GetLoginQrcode 获取登录的扫码二维码
func (s *XiaohongshuService) GetLoginQrcode(ctx context.Context) (*LoginQrcodeResponse, error) {
	b := newBrowser()
	page := b.NewPage()

	deferFunc := func() {
		_ = page.Close()
		b.Close()
	}

	loginAction := xiaohongshu.NewLogin(page)

	img, loggedIn, err := loginAction.FetchQrcodeImage(ctx)
	if err != nil || loggedIn {
		deferFunc()
	}
	if err != nil {
		return nil, fmt.Errorf("获取二维码失败: %w", err)
	}

	timeout := 4 * time.Minute

	if !loggedIn {
		go func() {
			ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			defer deferFunc()

			if loginAction.WaitForLogin(ctxTimeout) {
				if er := saveCookies(page); er != nil {
					logrus.WithError(er).Error("保存 cookies 失败")
				}
			}
		}()
	}

	return &LoginQrcodeResponse{
		Timeout: func() string {
			if loggedIn {
				return "0s"
			}
			return timeout.String()
		}(),
		Img:        img,
		IsLoggedIn: loggedIn,
	}, nil
}

// PublishContent 发布内容
func (s *XiaohongshuService) PublishContent(ctx context.Context, req *PublishRequest) (*PublishResponse, error) {
	// 验证标题长度
	if err := validateTitle(req.Title); err != nil {
		return nil, err
	}

	// 处理图片
	imagePaths, err := s.processImages(req.Images)
	if err != nil {
		return nil, fmt.Errorf("处理图片失败: %w", err)
	}

	// 构建发布内容
	content := xiaohongshu.PublishImageContent{
		Title:      req.Title,
		Content:    req.Content,
		Tags:       req.Tags,
		ImagePaths: imagePaths,
	}

	// 执行发布
	if err := withBrowserNoResult(func(page *rod.Page) error {
		action, err := xiaohongshu.NewPublishImageAction(page)
		if err != nil {
			return fmt.Errorf("创建发布操作失败: %w", err)
		}
		return action.Publish(ctx, content)
	}); err != nil {
		logrus.WithError(err).Errorf("发布内容失败: title=%s", content.Title)
		return nil, err
	}

	return &PublishResponse{
		Title:   req.Title,
		Content: req.Content,
		Images:  len(imagePaths),
		Status:  "发布完成",
	}, nil
}

// processImages 处理图片列表，支持URL下载和本地路径
func (s *XiaohongshuService) processImages(images []string) ([]string, error) {
	processor := downloader.NewImageProcessor()
	return processor.ProcessImages(images)
}

// PublishVideo 发布视频（本地文件）
func (s *XiaohongshuService) PublishVideo(ctx context.Context, req *PublishVideoRequest) (*PublishVideoResponse, error) {
	// 标题长度校验
	if err := validateTitle(req.Title); err != nil {
		return nil, err
	}

	// 本地视频文件校验
	if req.Video == "" {
		return nil, ErrEmptyVideoPath
	}
	if _, err := os.Stat(req.Video); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrVideoNotFound, err)
	}

	// 构建发布内容
	content := xiaohongshu.PublishVideoContent{
		Title:     req.Title,
		Content:   req.Content,
		Tags:      req.Tags,
		VideoPath: req.Video,
	}

	// 执行发布
	if err := withBrowserNoResult(func(page *rod.Page) error {
		action, err := xiaohongshu.NewPublishVideoAction(page)
		if err != nil {
			return fmt.Errorf("创建视频发布操作失败: %w", err)
		}
		return action.PublishVideo(ctx, content)
	}); err != nil {
		return nil, err
	}

	return &PublishVideoResponse{
		Title:   req.Title,
		Content: req.Content,
		Video:   req.Video,
		Status:  "发布完成",
	}, nil
}

// ListFeeds 获取Feeds列表
func (s *XiaohongshuService) ListFeeds(ctx context.Context) (*FeedsListResponse, error) {
	feeds, err := withBrowser(func(page *rod.Page) ([]xiaohongshu.Feed, error) {
		action := xiaohongshu.NewFeedsListAction(page)
		return action.GetFeedsList(ctx)
	})
	if err != nil {
		logrus.WithError(err).Error("获取 Feeds 列表失败")
		return nil, fmt.Errorf("获取 Feeds 列表失败: %w", err)
	}

	return &FeedsListResponse{
		Feeds: feeds,
		Count: len(feeds),
	}, nil
}

// SearchFeeds 搜索Feeds
func (s *XiaohongshuService) SearchFeeds(ctx context.Context, keyword string, filters ...xiaohongshu.FilterOption) (*FeedsListResponse, error) {
	feeds, err := withBrowser(func(page *rod.Page) ([]xiaohongshu.Feed, error) {
		action := xiaohongshu.NewSearchAction(page)
		return action.Search(ctx, keyword, filters...)
	})
	if err != nil {
		return nil, fmt.Errorf("搜索 Feeds 失败: %w", err)
	}

	return &FeedsListResponse{
		Feeds: feeds,
		Count: len(feeds),
	}, nil
}

// GetFeedDetail 获取Feed详情
func (s *XiaohongshuService) GetFeedDetail(ctx context.Context, feedID, xsecToken string) (*FeedDetailResponse, error) {
	result, err := withBrowser(func(page *rod.Page) (any, error) {
		action := xiaohongshu.NewFeedDetailAction(page)
		return action.GetFeedDetail(ctx, feedID, xsecToken)
	})
	if err != nil {
		return nil, fmt.Errorf("获取 Feed 详情失败: %w", err)
	}

	return &FeedDetailResponse{
		FeedID: feedID,
		Data:   result,
	}, nil
}

// UserProfile 获取用户信息
func (s *XiaohongshuService) UserProfile(ctx context.Context, userID, xsecToken string) (*UserProfileResponse, error) {
	result, err := withBrowser(func(page *rod.Page) (*xiaohongshu.UserProfileResponse, error) {
		action := xiaohongshu.NewUserProfileAction(page)
		return action.UserProfile(ctx, userID, xsecToken)
	})
	if err != nil {
		return nil, fmt.Errorf("获取用户主页失败: %w", err)
	}

	return &UserProfileResponse{
		UserBasicInfo: result.UserBasicInfo,
		Interactions:  result.Interactions,
		Feeds:         result.Feeds,
	}, nil
}

// PostCommentToFeed 发表评论到Feed
func (s *XiaohongshuService) PostCommentToFeed(ctx context.Context, feedID, xsecToken, content string) (*PostCommentResponse, error) {
	err := withBrowserNoResult(func(page *rod.Page) error {
		action := xiaohongshu.NewCommentFeedAction(page)
		return action.PostComment(ctx, feedID, xsecToken, content)
	})
	if err != nil {
		return nil, fmt.Errorf("发表评论失败: %w", err)
	}

	return &PostCommentResponse{FeedID: feedID, Success: true, Message: "评论发表成功"}, nil
}

// LikeFeed 点赞笔记
func (s *XiaohongshuService) LikeFeed(ctx context.Context, feedID, xsecToken string) (*ActionResult, error) {
	err := withBrowserNoResult(func(page *rod.Page) error {
		action := xiaohongshu.NewLikeAction(page)
		return action.Like(ctx, feedID, xsecToken)
	})
	if err != nil {
		return nil, fmt.Errorf("点赞失败: %w", err)
	}
	return &ActionResult{FeedID: feedID, Success: true, Message: "点赞成功或已点赞"}, nil
}

// UnlikeFeed 取消点赞笔记
func (s *XiaohongshuService) UnlikeFeed(ctx context.Context, feedID, xsecToken string) (*ActionResult, error) {
	err := withBrowserNoResult(func(page *rod.Page) error {
		action := xiaohongshu.NewLikeAction(page)
		return action.Unlike(ctx, feedID, xsecToken)
	})
	if err != nil {
		return nil, fmt.Errorf("取消点赞失败: %w", err)
	}
	return &ActionResult{FeedID: feedID, Success: true, Message: "取消点赞成功或未点赞"}, nil
}

// FavoriteFeed 收藏笔记
func (s *XiaohongshuService) FavoriteFeed(ctx context.Context, feedID, xsecToken string) (*ActionResult, error) {
	err := withBrowserNoResult(func(page *rod.Page) error {
		action := xiaohongshu.NewFavoriteAction(page)
		return action.Favorite(ctx, feedID, xsecToken)
	})
	if err != nil {
		return nil, fmt.Errorf("收藏失败: %w", err)
	}
	return &ActionResult{FeedID: feedID, Success: true, Message: "收藏成功或已收藏"}, nil
}

// UnfavoriteFeed 取消收藏笔记
func (s *XiaohongshuService) UnfavoriteFeed(ctx context.Context, feedID, xsecToken string) (*ActionResult, error) {
	err := withBrowserNoResult(func(page *rod.Page) error {
		action := xiaohongshu.NewFavoriteAction(page)
		return action.Unfavorite(ctx, feedID, xsecToken)
	})
	if err != nil {
		return nil, fmt.Errorf("取消收藏失败: %w", err)
	}
	return &ActionResult{FeedID: feedID, Success: true, Message: "取消收藏成功或未收藏"}, nil
}

func newBrowser() *browser.Browser {
	return browser.NewBrowser(configs.IsHeadless(), browser.WithBinPath(configs.GetBinPath()))
}

func saveCookies(page *rod.Page) error {
	cks, err := page.Browser().GetCookies()
	if err != nil {
		return fmt.Errorf("获取 cookies 失败: %w", err)
	}

	data, err := json.Marshal(cks)
	if err != nil {
		return fmt.Errorf("序列化 cookies 失败: %w", err)
	}

	cookieLoader := cookies.NewLoadCookie(cookies.GetCookiesFilePath())
	if err := cookieLoader.SaveCookies(data); err != nil {
		return fmt.Errorf("保存 cookies 失败: %w", err)
	}
	return nil
}

// GetMyProfile 获取当前登录用户的个人信息
func (s *XiaohongshuService) GetMyProfile(ctx context.Context) (*UserProfileResponse, error) {
	result, err := withBrowser(func(page *rod.Page) (*xiaohongshu.UserProfileResponse, error) {
		action := xiaohongshu.NewUserProfileAction(page)
		return action.GetMyProfileViaSidebar(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("获取我的主页失败: %w", err)
	}

	return &UserProfileResponse{
		UserBasicInfo: result.UserBasicInfo,
		Interactions:  result.Interactions,
		Feeds:         result.Feeds,
	}, nil
}
