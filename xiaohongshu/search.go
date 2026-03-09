package xiaohongshu

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/ajia1206/xhs-mcp/errors"
	"github.com/go-rod/rod"
)

type SearchResult struct {
	Search struct {
		Feeds FeedsValue `json:"feeds"`
	} `json:"search"`
}

type FilterOption struct {
	FiltersIndex int    `json:"filters_index" jsonschema:"筛选组索引 1=排序依据, 2=笔记类型, 3=发布时间, 4=搜索范围, 5=位置距离"`
	TagsIndex    int    `json:"tags_index" jsonschema:"标签索引，根据不同的筛选组索引对应不同的选项: 1=排序依据(1-5), 2=笔记类型(1-3), 3=发布时间(1-4), 4=搜索范围(1-4), 5=位置距离(1-3)"`
	Text         string `json:"text" jsonschema:"标签文本描述"`
}

// 预定义的筛选选项映射表
var FilterOptionsMap = map[int][]FilterOption{
	1: { // 排序依据
		{FiltersIndex: 1, TagsIndex: 1, Text: "综合"},
		{FiltersIndex: 1, TagsIndex: 2, Text: "最新"},
		{FiltersIndex: 1, TagsIndex: 3, Text: "最多点赞"},
		{FiltersIndex: 1, TagsIndex: 4, Text: "最多评论"},
		{FiltersIndex: 1, TagsIndex: 5, Text: "最多收藏"},
	},
	2: { // 笔记类型
		{FiltersIndex: 2, TagsIndex: 1, Text: "不限"},
		{FiltersIndex: 2, TagsIndex: 2, Text: "视频"},
		{FiltersIndex: 2, TagsIndex: 3, Text: "图文"},
	},
	3: { // 发布时间
		{FiltersIndex: 3, TagsIndex: 1, Text: "不限"},
		{FiltersIndex: 3, TagsIndex: 2, Text: "一天内"},
		{FiltersIndex: 3, TagsIndex: 3, Text: "一周内"},
		{FiltersIndex: 3, TagsIndex: 4, Text: "半年内"},
	},
	4: { // 搜索范围
		{FiltersIndex: 4, TagsIndex: 1, Text: "不限"},
		{FiltersIndex: 4, TagsIndex: 2, Text: "已看过"},
		{FiltersIndex: 4, TagsIndex: 3, Text: "未看过"},
		{FiltersIndex: 4, TagsIndex: 4, Text: "已关注"},
	},
	5: { // 位置距离
		{FiltersIndex: 5, TagsIndex: 1, Text: "不限"},
		{FiltersIndex: 5, TagsIndex: 2, Text: "同城"},
		{FiltersIndex: 5, TagsIndex: 3, Text: "附近"},
	},
}

// 定义筛选组索引到中文描述的映射
var filterGroupMap = map[int]string{
	1: "排序依据",
	2: "笔记类型",
	3: "发布时间",
	4: "搜索范围",
	5: "位置距离",
}

// validateFilterOption 验证筛选选项是否在有效范围内
func validateFilterOption(filter FilterOption) error {
	// 检查筛选组索引是否有效
	if filter.FiltersIndex < 1 || filter.FiltersIndex > 5 {
		return fmt.Errorf("无效的筛选组索引 %d，有效范围为 1-5", filter.FiltersIndex)
	}

	// 检查标签索引是否在对应筛选组的有效范围内
	options, exists := FilterOptionsMap[filter.FiltersIndex]
	if !exists {
		return fmt.Errorf("筛选组 %d 不存在", filter.FiltersIndex)
	}

	if filter.TagsIndex < 1 || filter.TagsIndex > len(options) {
		return fmt.Errorf("筛选组 %d 的标签索引 %d 超出范围，有效范围为 1-%d",
			filter.FiltersIndex, filter.TagsIndex, len(options))
	}

	return nil
}

// 便利函数：根据文本创建筛选选项
func NewFilterOption(filtersIndex int, text string) (FilterOption, error) {
	options, exists := FilterOptionsMap[filtersIndex]
	if !exists {
		return FilterOption{}, fmt.Errorf("筛选组 %d 不存在", filtersIndex)
	}

	for _, option := range options {
		if option.Text == text {
			return option, nil
		}
	}

	return FilterOption{}, fmt.Errorf("在筛选组 %d 中未找到文本 '%s'", filtersIndex, text)
}

// 便利函数：创建常用的筛选选项
func SortBy(text string) (FilterOption, error) {
	return NewFilterOption(1, text) // 排序依据
}

func NoteType(text string) (FilterOption, error) {
	return NewFilterOption(2, text) // 笔记类型
}

func TimeRange(text string) (FilterOption, error) {
	return NewFilterOption(3, text) // 发布时间
}

func SearchScope(text string) (FilterOption, error) {
	return NewFilterOption(4, text) // 搜索范围
}

func LocationDistance(text string) (FilterOption, error) {
	return NewFilterOption(5, text) // 位置距离
}

// GetFilterGroupDescription 根据筛选组索引获取中文描述
func GetFilterGroupDescription(index int) string {
	if desc, exists := filterGroupMap[index]; exists {
		return desc
	}
	return "未知筛选组"
}

// GetFilterGroupIndex 根据中文描述获取筛选组索引
func GetFilterGroupIndex(text string) int {
	// 通过遍历filterGroupMap获取对应的索引
	for index, description := range filterGroupMap {
		if description == text {
			return index
		}
	}
	return -1 // 未找到匹配项时返回-1
}

type SearchAction struct {
	page *rod.Page
}

func NewSearchAction(page *rod.Page) *SearchAction {
	pp := page.Timeout(120 * time.Second)

	return &SearchAction{page: pp}
}

func (s *SearchAction) Search(ctx context.Context, keyword string, filters ...FilterOption) ([]Feed, error) {
	page := s.page.Context(ctx)

	// 直接跳转到搜索页面（NewPage 已经访问过首页）
	searchURL := makeSearchURL(keyword)
	page.MustNavigate(searchURL)
	page.MustWaitStable()

	page.MustWait(`() => window.__INITIAL_STATE__ !== undefined`)

	// 等待页面数据加载（搜索页面需要更多时间）
	time.Sleep(3 * time.Second)

	// 如果 search.feeds 为空，尝试滚动触发加载
	hasSearch := page.MustEval(`() => !!(window.__INITIAL_STATE__?.search?.feeds)`).Bool()
	if hasSearch {
		dataLen := page.MustEval(`() => {
			const feeds = window.__INITIAL_STATE__?.search?.feeds;
			const data = feeds?.value || feeds?._value;
			return data ? data.length : -1;
		}`).Int()

		// 如果为空，尝试滚动加载
		if dataLen == 0 {
			page.MustEval(`() => window.scrollTo(0, document.body.scrollHeight)`)
			time.Sleep(2 * time.Second)
		}
	}

	// 如果有筛选条件，则应用筛选
	if len(filters) > 0 {
		// 验证所有筛选选项
		for _, filter := range filters {
			if err := validateFilterOption(filter); err != nil {
				return nil, fmt.Errorf("筛选选项验证失败: %w", err)
			}
		}

		// 悬停在筛选按钮上
		filterButton := page.MustElement(`div.filter`)
		filterButton.MustHover()

		// 等待筛选面板出现
		page.MustWait(`() => document.querySelector('div.filter-panel') !== null`)

		// 应用所有筛选条件
		for _, filter := range filters {
			selector := fmt.Sprintf(`div.filter-panel div.filters:nth-child(%d) div.tags:nth-child(%d)`,
				filter.FiltersIndex, filter.TagsIndex)
			option := page.MustElement(selector)
			option.MustClick()
		}

		// 等待页面更新
		page.MustWaitStable()
		// 重新等待 __INITIAL_STATE__ 更新
		page.MustWait(`() => window.__INITIAL_STATE__ !== undefined`)
	}

	fetchFeeds := func() ([]Feed, error) {
		// 尝试从 search.feeds 获取
		result := page.MustEval(`() => {
			if (window.__INITIAL_STATE__ &&
			    window.__INITIAL_STATE__.search &&
			    window.__INITIAL_STATE__.search.feeds) {
				const feeds = window.__INITIAL_STATE__.search.feeds;
				const feedsData = feeds.value !== undefined ? feeds.value : feeds._value;
				if (feedsData && feedsData.length > 0) {
					const seen = new WeakSet();
					return JSON.stringify(feedsData, function(key, value) {
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

		// 如果 search.feeds 为空，尝试从 feed.feeds 获取
		if result == "" {
			result = page.MustEval(`() => {
				if (window.__INITIAL_STATE__ &&
				    window.__INITIAL_STATE__.feed &&
				    window.__INITIAL_STATE__.feed.feeds) {
					const feeds = window.__INITIAL_STATE__.feed.feeds;
					const feedsData = feeds.value !== undefined ? feeds.value : feeds._value;
					if (feedsData && feedsData.length > 0) {
						const seen = new WeakSet();
						return JSON.stringify(feedsData, function(key, value) {
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
		}

		if result == "" {
			return nil, errors.ErrNoFeeds
		}

		var feeds []Feed
		if err := json.Unmarshal([]byte(result), &feeds); err != nil {
			return nil, fmt.Errorf("failed to unmarshal feeds: %w", err)
		}

		return feeds, nil
	}

	// 增加滚动抓取：循环下拉，直到没有新数据或达到最大滚动次数
	const maxScroll = 10              // 最多滚动次数（优化：从30减少到10）
	const noNewThreshold = 3          // 连续无新增阈值（优化：从5减少到3）
	const waitBetween = time.Second   // 滚动后等待时间
	seen := make(map[string]struct{}) // 去重
	var collected []Feed
	noNewCount := 0
	lastTotal := 0

	// TODO: 网络监听功能暂时禁用，需要更新 Rod API 用法
	// 使用页面 JavaScript 获取数据已经足够

	for i := 0; i < maxScroll; i++ {
		feeds, err := fetchFeeds()
		if err == nil {
			for _, f := range feeds {
				if _, ok := seen[f.ID]; !ok {
					seen[f.ID] = struct{}{}
					collected = append(collected, f)
				}
			}
		}

		if len(collected) == lastTotal {
			noNewCount++
		} else {
			noNewCount = 0
			lastTotal = len(collected)
		}

		if noNewCount >= noNewThreshold {
			break
		}

		// 触发下拉加载更多
		page.MustEval(`() => window.scrollTo(0, document.body.scrollHeight)`)
		page.MustWaitStable()
		time.Sleep(waitBetween)
	}

	if len(collected) == 0 {
		return nil, errors.ErrNoFeeds
	}

	return collected, nil
}

func makeSearchURL(keyword string) string {

	values := url.Values{}
	values.Set("keyword", keyword)
	values.Set("source", "web_explore_feed")

	//https://www.xiaohongshu.com/search_result?keyword=%25E7%258E%258B%25E5%25AD%2590&source=web_search_result_notes
	//https://www.xiaohongshu.com/search_result?keyword=%25E7%258E%258B%25E5%25AD%2590&source=web_explore_feed
	return fmt.Sprintf("https://www.xiaohongshu.com/search_result?%s", values.Encode())
}

// parseFeedsFromAPIBodies 尝试从搜索 API 响应体中提取 Feed 列表（结构未知时兜底）
func parseFeedsFromAPIBodies(bodies []string) []Feed {
	seen := make(map[string]struct{})
	var feeds []Feed

	var walk func(v any)
	walk = func(v any) {
		switch val := v.(type) {
		case map[string]any:
			if _, has := val["id"]; has {
				buf, err := json.Marshal(val)
				if err == nil {
					var f Feed
					if err := json.Unmarshal(buf, &f); err == nil && f.ID != "" {
						if _, ok := seen[f.ID]; !ok {
							seen[f.ID] = struct{}{}
							feeds = append(feeds, f)
						}
					}
				}
			}
			for _, vv := range val {
				walk(vv)
			}
		case []any:
			for _, vv := range val {
				walk(vv)
			}
		}
	}

	for _, body := range bodies {
		var root any
		if err := json.Unmarshal([]byte(body), &root); err != nil {
			continue
		}
		walk(root)
	}

	return feeds
}
