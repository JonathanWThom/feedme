package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Reddit feed types (correspond to URL paths)
const (
	RedditFeedHot    = "hot"
	RedditFeedNew    = "new"
	RedditFeedTop    = "top"
	RedditFeedRising = "rising"
	RedditFeedBest   = "best"
)

var RedditFeedNames = []string{RedditFeedHot, RedditFeedNew, RedditFeedTop, RedditFeedRising, RedditFeedBest}
var RedditFeedLabels = []string{"Hot", "New", "Top", "Rising", "Best"}

// RedditClient fetches data from Reddit's JSON API
type RedditClient struct {
	CachedSource
	http       *http.Client
	subreddit  string
	idToReddit map[int]string // Maps pseudo-ID to Reddit post ID
}

// NewRedditClient creates a new Reddit API client for a subreddit
func NewRedditClient(subreddit string) *RedditClient {
	subreddit = strings.TrimPrefix(subreddit, "r/")
	subreddit = strings.TrimPrefix(subreddit, "/r/")

	return &RedditClient{
		CachedSource: NewCachedSource(1 * time.Second),
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		subreddit:  subreddit,
		idToReddit: make(map[int]string),
	}
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
	if item.Type != "" && strings.HasPrefix(item.Type, "/r/") {
		return fmt.Sprintf("https://www.reddit.com%s", item.Type)
	}
	return item.URL
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

	ids := c.StoreItems(stories)

	c.idToReddit = make(map[int]string)
	for i, story := range stories {
		c.idToReddit[i+1] = story.Type
	}

	return ids, nil
}

// fetchStories fetches stories from Reddit
func (c *RedditClient) fetchStories(feed string) ([]*Item, error) {
	c.Throttle()

	url := fmt.Sprintf("https://www.reddit.com/r/%s/%s.json?limit=100", c.subreddit, feed)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "feedme:v1.0 (terminal news reader)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch r/%s: %w", c.subreddit, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		time.Sleep(2 * time.Second)
		c.Throttle()
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

	return parseRedditStories(listing), nil
}

// FetchCommentTree fetches comments for a story
func (c *RedditClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	c.Throttle()

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

	var listings []redditCommentListing
	if err := json.NewDecoder(resp.Body).Decode(&listings); err != nil {
		return nil, fmt.Errorf("failed to decode comments: %w", err)
	}

	if len(listings) < 2 {
		return nil, fmt.Errorf("unexpected response format")
	}

	return parseRedditComments(listings[1], maxDepth)
}
