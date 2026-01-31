package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const devtoBaseURL = "https://dev.to"
const devtoAPIURL = "https://dev.to/api"

// Dev.to feed types
const (
	DevtoFeedTop    = "top"    // Default - featured articles by popularity
	DevtoFeedLatest = "latest" // Fresh articles (state=fresh)
	DevtoFeedRising = "rising" // Rising articles (state=rising)
	DevtoFeedWeek   = "week"   // Top articles from last 7 days
)

var DevtoFeedNames = []string{DevtoFeedTop, DevtoFeedLatest, DevtoFeedRising, DevtoFeedWeek}
var DevtoFeedLabels = []string{"Top", "Latest", "Rising", "Week"}

// DevtoClient fetches data from the Dev.to API
type DevtoClient struct {
	http        *http.Client
	storyCache  map[int]*Item
	idToDevto   map[int]int // Maps pseudo-ID to Dev.to article ID
	cacheMu     sync.RWMutex
	lastRequest time.Time
	requestMu   sync.Mutex
}

// NewDevtoClient creates a new Dev.to API client
func NewDevtoClient() *DevtoClient {
	return &DevtoClient{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		storyCache: make(map[int]*Item),
		idToDevto:  make(map[int]int),
	}
}

// throttle ensures we don't make requests too quickly
func (c *DevtoClient) throttle() {
	c.requestMu.Lock()
	defer c.requestMu.Unlock()

	// Wait at least 500ms between requests to be polite
	minDelay := 500 * time.Millisecond
	elapsed := time.Since(c.lastRequest)
	if elapsed < minDelay {
		time.Sleep(minDelay - elapsed)
	}
	c.lastRequest = time.Now()
}

// Name returns the display name of the source
func (c *DevtoClient) Name() string {
	return "DEV"
}

// FeedNames returns the available feed names
func (c *DevtoClient) FeedNames() []string {
	return DevtoFeedNames
}

// FeedLabels returns the display labels for feeds
func (c *DevtoClient) FeedLabels() []string {
	return DevtoFeedLabels
}

// StoryURL returns the URL for viewing a story on Dev.to
func (c *DevtoClient) StoryURL(item *Item) string {
	// The Dev.to path is stored in the Type field
	if item.Type != "" && item.Type != "article" {
		return fmt.Sprintf("%s%s", devtoBaseURL, item.Type)
	}
	return item.URL
}

// devtoArticle represents a Dev.to article from the API
type devtoArticle struct {
	ID                   int       `json:"id"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	URL                  string    `json:"url"`
	Path                 string    `json:"path"`
	PublishedAt          time.Time `json:"published_at"`
	PublicReactionsCount int       `json:"public_reactions_count"`
	CommentsCount        int       `json:"comments_count"`
	ReadingTimeMinutes   int       `json:"reading_time_minutes"`
	TagList              []string  `json:"tag_list"`
	User                 struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"user"`
}

// FetchStoryIDs fetches story "IDs" for a feed
func (c *DevtoClient) FetchStoryIDs(feed string) ([]int, error) {
	stories, err := c.fetchStories(feed)
	if err != nil {
		return nil, err
	}

	if len(stories) == 0 {
		return nil, fmt.Errorf("no stories found for feed %q", feed)
	}

	// Cache stories and return pseudo-IDs
	ids := make([]int, len(stories))
	c.cacheMu.Lock()
	// Clear old cache
	c.storyCache = make(map[int]*Item)
	c.idToDevto = make(map[int]int)
	for i, story := range stories {
		id := i + 1 // 1-indexed pseudo-IDs
		c.storyCache[id] = story
		c.idToDevto[id] = story.ID
		ids[i] = id
	}
	c.cacheMu.Unlock()

	return ids, nil
}

// fetchStories fetches stories from Dev.to
func (c *DevtoClient) fetchStories(feed string) ([]*Item, error) {
	c.throttle()

	// Build URL based on feed type
	var url string
	switch feed {
	case DevtoFeedLatest:
		url = fmt.Sprintf("%s/articles?per_page=50&state=fresh", devtoAPIURL)
	case DevtoFeedRising:
		url = fmt.Sprintf("%s/articles?per_page=50&state=rising", devtoAPIURL)
	case DevtoFeedWeek:
		url = fmt.Sprintf("%s/articles?per_page=50&top=7", devtoAPIURL)
	default: // DevtoFeedTop
		url = fmt.Sprintf("%s/articles?per_page=50", devtoAPIURL)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Dev.to requires a User-Agent header
	req.Header.Set("User-Agent", "feedme:v1.0 (terminal news reader)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dev.to articles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		// Rate limited, wait and retry
		time.Sleep(2 * time.Second)
		c.throttle()
		req, _ = http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "feedme:v1.0 (terminal news reader)")
		resp, err = c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch dev.to articles after retry: %w", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("dev.to returned status %d", resp.StatusCode)
	}

	var articles []devtoArticle
	if err := json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		return nil, fmt.Errorf("failed to decode dev.to response: %w", err)
	}

	var stories []*Item
	for _, article := range articles {
		item := &Item{
			ID:          article.ID,
			Type:        article.Path, // Store path for later
			Title:       article.Title,
			By:          article.User.Username,
			Score:       article.PublicReactionsCount,
			URL:         article.URL,
			Time:        article.PublishedAt.Unix(),
			Descendants: article.CommentsCount,
		}

		// Store tags in Text field if present
		if len(article.TagList) > 0 {
			item.Text = "[" + joinTags(article.TagList) + "]"
		}

		stories = append(stories, item)
	}

	return stories, nil
}

// joinTags joins tags with comma separator
func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := tags[0]
	for i := 1; i < len(tags); i++ {
		result += ", " + tags[i]
	}
	return result
}

// FetchItem fetches a cached item by pseudo-ID
func (c *DevtoClient) FetchItem(id int) (*Item, error) {
	c.cacheMu.RLock()
	item, ok := c.storyCache[id]
	c.cacheMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("item %d not found in cache", id)
	}

	return item, nil
}

// FetchItems fetches multiple cached items by pseudo-ID
func (c *DevtoClient) FetchItems(ids []int) ([]*Item, error) {
	items := make([]*Item, len(ids))
	c.cacheMu.RLock()
	for i, id := range ids {
		if item, ok := c.storyCache[id]; ok {
			items[i] = item
		}
	}
	c.cacheMu.RUnlock()
	return items, nil
}

// devtoComment represents a Dev.to comment from the API
type devtoComment struct {
	IDCode    string `json:"id_code"`
	BodyHTML  string `json:"body_html"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"user"`
	Children []devtoComment `json:"children"`
}

// FetchCommentTree fetches comments for a story
func (c *DevtoClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	c.throttle()

	// Get the Dev.to article ID from the item
	articleID := item.ID

	url := fmt.Sprintf("%s/comments?a_id=%d", devtoAPIURL, articleID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "feedme:v1.0 (terminal news reader)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("dev.to returned status %d", resp.StatusCode)
	}

	var devtoComments []devtoComment
	if err := json.NewDecoder(resp.Body).Decode(&devtoComments); err != nil {
		return nil, fmt.Errorf("failed to decode comments: %w", err)
	}

	return c.parseComments(devtoComments, 0, maxDepth), nil
}

// parseComments converts Dev.to comments to our Comment type
func (c *DevtoClient) parseComments(devtoComments []devtoComment, depth int, maxDepth int) []*Comment {
	var comments []*Comment

	for _, dc := range devtoComments {
		comment := c.parseComment(dc, depth, maxDepth)
		if comment != nil {
			comments = append(comments, comment)
		}
	}

	return comments
}

// parseComment converts a single Dev.to comment to our Comment type
func (c *DevtoClient) parseComment(dc devtoComment, depth int, maxDepth int) *Comment {
	if dc.User.Username == "" {
		return nil
	}

	// Parse the created_at time
	var timestamp int64
	if t, err := time.Parse(time.RFC3339, dc.CreatedAt); err == nil {
		timestamp = t.Unix()
	}

	item := &Item{
		ID:   hashShortID(dc.IDCode),
		Type: "comment",
		By:   dc.User.Username,
		Text: dc.BodyHTML,
		Time: timestamp,
	}

	comment := &Comment{
		Item:  item,
		Depth: depth,
	}

	// Parse nested replies if present and within depth limit
	if (maxDepth <= 0 || depth < maxDepth) && len(dc.Children) > 0 {
		comment.Children = c.parseComments(dc.Children, depth+1, maxDepth)
	}

	return comment
}
