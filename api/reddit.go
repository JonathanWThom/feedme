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

const redditUserAgent = "feedme:v1.0 (terminal news reader)"

// fetchStories fetches stories from Reddit
func (c *RedditClient) fetchStories(feed string) ([]*Item, error) {
	c.Throttle()

	url := fmt.Sprintf("https://www.reddit.com/r/%s/%s.json?limit=100", c.subreddit, feed)
	resp, err := doWithRetry(c.http, url, redditUserAgent, &c.CachedSource)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch r/%s: %w", c.subreddit, err)
	}
	defer resp.Body.Close()

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

	listings, err := c.fetchCommentListings(permalink)
	if err != nil {
		return nil, err
	}
	return parseRedditComments(listings[1], maxDepth)
}

func (c *RedditClient) fetchCommentListings(permalink string) ([]redditCommentListing, error) {
	url := fmt.Sprintf("https://www.reddit.com%s.json?limit=200", permalink)
	resp, err := doWithRetry(c.http, url, redditUserAgent, &c.CachedSource)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer resp.Body.Close()

	var listings []redditCommentListing
	if err := json.NewDecoder(resp.Body).Decode(&listings); err != nil {
		return nil, fmt.Errorf("failed to decode comments: %w", err)
	}
	if len(listings) < 2 {
		return nil, fmt.Errorf("unexpected response format")
	}
	return listings, nil
}
