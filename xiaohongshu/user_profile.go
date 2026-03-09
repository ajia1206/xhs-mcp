package xiaohongshu

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rod/rod"
)

type UserProfileAction struct {
	page *rod.Page
}

func NewUserProfileAction(page *rod.Page) *UserProfileAction {
	pp := page.Timeout(60 * time.Second)
	return &UserProfileAction{page: pp}
}

// UserProfile 获取用户基本信息及帖子
func (u *UserProfileAction) UserProfile(ctx context.Context, userID, xsecToken string) (*UserProfileResponse, error) {
	page := u.page.Context(ctx)

	searchURL := makeUserProfileURL(userID, xsecToken)
	page.MustNavigate(searchURL)
	page.MustWaitStable()

	return u.extractUserProfileData(page)
}

// extractUserProfileData 从页面中提取用户资料数据的通用方法（支持滚动加载）
func (u *UserProfileAction) extractUserProfileData(page *rod.Page) (*UserProfileResponse, error) {
	page.MustWait(`() => window.__INITIAL_STATE__ !== undefined`)

	userDataResult := page.MustEval(`() => {
		if (window.__INITIAL_STATE__ &&
		    window.__INITIAL_STATE__.user &&
		    window.__INITIAL_STATE__.user.userPageData) {
			const userPageData = window.__INITIAL_STATE__.user.userPageData;
			const data = userPageData.value !== undefined ? userPageData.value : userPageData._value;
			if (data) {
				return JSON.stringify(data);
			}
		}
		return "";
	}`).String()

	if userDataResult == "" {
		return nil, fmt.Errorf("user.userPageData.value not found in __INITIAL_STATE__")
	}

	// 解析用户信息
	var userPageData struct {
		Interactions []UserInteractions `json:"interactions"`
		BasicInfo    UserBasicInfo      `json:"basicInfo"`
	}
	if err := json.Unmarshal([]byte(userDataResult), &userPageData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal userPageData: %w", err)
	}

	// 组装响应
	response := &UserProfileResponse{
		UserBasicInfo: userPageData.BasicInfo,
		Interactions:  userPageData.Interactions,
	}

	// 滚动加载更多笔记
	const maxScroll = 20              // 最多滚动次数
	const waitBetween = 2 * time.Second // 每次滚动后等待时间
	seen := make(map[string]struct{}) // 去重

	for i := 0; i < maxScroll; i++ {
		// 获取当前所有笔记
		notesResult := page.MustEval(`() => {
			if (window.__INITIAL_STATE__ &&
			    window.__INITIAL_STATE__.user &&
			    window.__INITIAL_STATE__.user.notes) {
				const notes = window.__INITIAL_STATE__.user.notes;
				const data = notes.value !== undefined ? notes.value : notes._value;
				if (data) {
					const seen = new WeakSet();
					return JSON.stringify(data, function(key, value) {
						if (typeof value === "object" && value !== null) {
							if (seen.has(value)) {
								return;
							}
							seen.add(value);
						}
						return value;
					});
				}
			}
			return "";
		}`).String()

		if notesResult != "" {
			// 解析帖子数据
			var notesFeeds [][]Feed
			if err := json.Unmarshal([]byte(notesResult), &notesFeeds); err == nil {
				// 展平并去重
				newCount := 0
				for _, feeds := range notesFeeds {
					for _, feed := range feeds {
						if _, ok := seen[feed.ID]; !ok {
							seen[feed.ID] = struct{}{}
							response.Feeds = append(response.Feeds, feed)
							newCount++
						}
					}
				}
				// 如果没有新数据，停止滚动
				if newCount == 0 && len(seen) > 0 {
					break
				}
			}
		}

		// 滚动到底部加载更多
		page.MustEval(`() => window.scrollTo(0, document.body.scrollHeight)`)
		time.Sleep(waitBetween)
	}

	return response, nil
}

func makeUserProfileURL(userID, xsecToken string) string {
	return fmt.Sprintf("https://www.xiaohongshu.com/user/profile/%s?xsec_token=%s&xsec_source=pc_note", userID, xsecToken)
}

func (u *UserProfileAction) GetMyProfileViaSidebar(ctx context.Context) (*UserProfileResponse, error) {
	page := u.page.Context(ctx)

	// 创建导航动作
	navigate := NewNavigate(page)

	// 通过侧边栏导航到个人主页
	if err := navigate.ToProfilePage(ctx); err != nil {
		return nil, fmt.Errorf("failed to navigate to profile page via sidebar: %w", err)
	}

	// 等待页面加载完成并获取 __INITIAL_STATE__
	page.MustWaitStable()

	return u.extractUserProfileData(page)
}
