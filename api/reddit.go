package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Reddit feed types (correspond to URL paths)
const (
	RedditFeedHot         = "hot"
	RedditFeedNew         = "new"
	RedditFeedTop         = "top"
	RedditFeedRising      = "rising"
	RedditFeedBest        = "best"
)

var RedditFeedNames = []string{RedditFeedHot, RedditFeedNew, RedditFeedTop, RedditFeedRising, RedditFeedBest}
var RedditFeedLabels = []string{"Hot", "New", "Top", "Rising", "Best"}

// RedditClient fetches data from Reddit's JSON API
type RedditClient struct {
	http        *http.Client
	subreddit   string
	storyCache  map[int]*Item
	idToReddit  map[int]string // Maps pseudo-ID to Reddit post ID
	cacheMu     sync.RWMutex
	lastRequest time.Time
	requestMu   sync.Mutex
}

// NewRedditClient creates a new Reddit API client for a subreddit
func NewRedditClient(subreddit string) *RedditClient {
	// Clean up subreddit name (remove r/ prefix if present)
	subreddit = strings.TrimPrefix(subreddit, "r/")
	subreddit = strings.TrimPrefix(subreddit, "/r/")

	return &RedditClient{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		subreddit:  subreddit,
		storyCache: make(map[int]*Item),
		idToReddit: make(map[int]string),
	}
}

// throttle ensures we don't make requests too quickly
func (c *RedditClient) throttle() {
	c.requestMu.Lock()
	defer c.requestMu.Unlock()

	// Wait at least 1 second between requests (Reddit rate limits)
	minDelay := 1 * time.Second
	elapsed := time.Since(c.lastRequest)
	if elapsed < minDelay {
		time.Sleep(minDelay - elapsed)
	}
	c.lastRequest = time.Now()
}

// Name returns the display name of the source
func (c *RedditClient) Name() string {
	return fmt.Sprintf("r/%s", c.subreddit)
}

// FeedNames returns the available feed names
func (c *RedditClient) FeedNames() []string {
	return RedditFeedNames
}

// FeedLabels returns the display labels for feeds
func (c *RedditClient) FeedLabels() []string {
	return RedditFeedLabels
}

// StoryURL returns the URL for viewing a story on Reddit
func (c *RedditClient) StoryURL(item *Item) string {
	// The Reddit permalink is stored in the Type field
	if item.Type != "" && strings.HasPrefix(item.Type, "/r/") {
		return fmt.Sprintf("https://www.reddit.com%s", item.Type)
	}
	return item.URL
}

// redditListing represents the top-level JSON structure
type redditListing struct {
	Data struct {
		Children []struct {
			Data redditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// redditPost represents a Reddit post
type redditPost struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Author       string  `json:"author"`
	Score        int     `json:"score"`
	URL          string  `json:"url"`
	Permalink    string  `json:"permalink"`
	NumComments  int     `json:"num_comments"`
	CreatedUTC   float64 `json:"created_utc"`
	Selftext     string  `json:"selftext"`
	IsSelf       bool    `json:"is_self"`
	Subreddit    string  `json:"subreddit"`
	Domain       string  `json:"domain"`
	LinkFlairText string `json:"link_flair_text"`
}

// FetchStoryIDs fetches story "IDs" for a feed
func (c *RedditClient) FetchStoryIDs(feed string) ([]int, error) {
	stories, err := c.fetchStories(feed)
	if err != nil {
		return nil, err
	}

	if len(stories) == 0 {
		return nil, fmt.Errorf("no stories found for r/%s/%s", c.subreddit, feed)
	}

	// Cache stories and return pseudo-IDs
	ids := make([]int, len(stories))
	c.cacheMu.Lock()
	// Clear old cache
	c.storyCache = make(map[int]*Item)
	c.idToReddit = make(map[int]string)
	for i, story := range stories {
		id := i + 1 // 1-indexed pseudo-IDs
		c.storyCache[id] = story
		c.idToReddit[id] = story.Type // Store Reddit ID (stored in Type field temporarily)
		ids[i] = id
	}
	c.cacheMu.Unlock()

	return ids, nil
}

// fetchStories fetches stories from Reddit
func (c *RedditClient) fetchStories(feed string) ([]*Item, error) {
	c.throttle()

	url := fmt.Sprintf("https://www.reddit.com/r/%s/%s.json?limit=100", c.subreddit, feed)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Reddit requires a User-Agent header
	req.Header.Set("User-Agent", "feedme:v1.0 (terminal news reader)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch r/%s: %w", c.subreddit, err)
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
			return nil, fmt.Errorf("failed to fetch r/%s after retry: %w", c.subreddit, err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("reddit returned status %d for r/%s", resp.StatusCode, c.subreddit)
	}

	var listing redditListing
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return nil, fmt.Errorf("failed to decode reddit response: %w", err)
	}

	var stories []*Item
	for _, child := range listing.Data.Children {
		post := child.Data
		item := &Item{
			ID:          hashShortID(post.ID),
			Type:        post.Permalink, // Store permalink for later
			Title:       post.Title,
			By:          post.Author,
			Score:       post.Score,
			URL:         post.URL,
			Time:        int64(post.CreatedUTC),
			Descendants: post.NumComments,
		}

		// Store flair in Text field if present
		if post.LinkFlairText != "" {
			item.Text = "[" + post.LinkFlairText + "]"
		}

		// For self posts, the URL should point to Reddit
		if post.IsSelf {
			item.URL = fmt.Sprintf("https://www.reddit.com%s", post.Permalink)
		}

		stories = append(stories, item)
	}

	return stories, nil
}

// FetchItem fetches a cached item by pseudo-ID
func (c *RedditClient) FetchItem(id int) (*Item, error) {
	c.cacheMu.RLock()
	item, ok := c.storyCache[id]
	c.cacheMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("item %d not found in cache", id)
	}

	return item, nil
}

// FetchItems fetches multiple cached items by pseudo-ID
func (c *RedditClient) FetchItems(ids []int) ([]*Item, error) {
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

// redditCommentListing represents the comments JSON structure
type redditCommentListing struct {
	Data struct {
		Children []struct {
			Kind string          `json:"kind"`
			Data json.RawMessage `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// redditComment represents a Reddit comment
type redditComment struct {
	ID         string  `json:"id"`
	Author     string  `json:"author"`
	Body       string  `json:"body"`
	Score      int     `json:"score"`
	CreatedUTC float64 `json:"created_utc"`
	Depth      int     `json:"depth"`
	Replies    any `json:"replies"` // Can be "" or a listing
}

// FetchCommentTree fetches comments for a story
func (c *RedditClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	c.throttle()

	// Get the permalink from the Type field
	permalink := item.Type
	if permalink == "" || !strings.HasPrefix(permalink, "/r/") {
		return nil, fmt.Errorf("no permalink available")
	}

	url := fmt.Sprintf("https://www.reddit.com%s.json?limit=200", permalink)

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
		return nil, fmt.Errorf("reddit returned status %d", resp.StatusCode)
	}

	// Reddit returns an array: [post_listing, comments_listing]
	var listings []redditCommentListing
	if err := json.NewDecoder(resp.Body).Decode(&listings); err != nil {
		return nil, fmt.Errorf("failed to decode comments: %w", err)
	}

	if len(listings) < 2 {
		return nil, fmt.Errorf("unexpected response format")
	}

	return c.parseComments(listings[1], maxDepth)
}

// parseComments extracts comments from the listing
func (c *RedditClient) parseComments(listing redditCommentListing, maxDepth int) ([]*Comment, error) {
	var comments []*Comment

	for _, child := range listing.Data.Children {
		if child.Kind != "t1" { // t1 = comment
			continue
		}

		var rc redditComment
		if err := json.Unmarshal(child.Data, &rc); err != nil {
			continue
		}

		comment := c.parseComment(rc, maxDepth)
		if comment != nil {
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

// parseComment converts a Reddit comment to our Comment type
func (c *RedditClient) parseComment(rc redditComment, maxDepth int) *Comment {
	if rc.Author == "" || rc.Author == "[deleted]" {
		return nil
	}

	item := &Item{
		ID:    hashShortID(rc.ID),
		Type:  "comment",
		By:    rc.Author,
		Text:  rc.Body,
		Score: rc.Score,
		Time:  int64(rc.CreatedUTC),
	}

	comment := &Comment{
		Item:  item,
		Depth: rc.Depth,
	}

	// Parse nested replies if present and within depth limit
	if maxDepth <= 0 || rc.Depth < maxDepth {
		if replies, ok := rc.Replies.(map[string]any); ok {
			if data, ok := replies["data"].(map[string]any); ok {
				if children, ok := data["children"].([]any); ok {
					for _, child := range children {
						if childMap, ok := child.(map[string]any); ok {
							if childMap["kind"] == "t1" {
								if childData, ok := childMap["data"].(map[string]any); ok {
									childJSON, _ := json.Marshal(childData)
									var childRC redditComment
									if err := json.Unmarshal(childJSON, &childRC); err == nil {
										if childComment := c.parseComment(childRC, maxDepth); childComment != nil {
											comment.Children = append(comment.Children, childComment)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return comment
}
