package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const lobstersBaseURL = "https://lobste.rs"

// Lobste.rs feed types (correspond to URL paths)
const (
	LobstersFeedHottest = ""        // Default front page
	LobstersFeedNewest  = "newest"  // /newest
	LobstersFeedRecent  = "recent"  // /recent (recently active)
)

var LobstersFeedNames = []string{LobstersFeedHottest, LobstersFeedNewest, LobstersFeedRecent}
var LobstersFeedLabels = []string{"Hot", "New", "Recent"}

// LobstersClient scrapes lobste.rs
type LobstersClient struct {
	CachedSource
	http *http.Client
}

// NewLobstersClient creates a new Lobste.rs scraping client
func NewLobstersClient() *LobstersClient {
	return &LobstersClient{
		CachedSource: NewCachedSource(500 * time.Millisecond),
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Name returns the display name of the source
func (c *LobstersClient) Name() string {
	return "Lobsters"
}

// FeedNames returns the available feed names
func (c *LobstersClient) FeedNames() []string {
	return LobstersFeedNames
}

// FeedLabels returns the display labels for feeds
func (c *LobstersClient) FeedLabels() []string {
	return LobstersFeedLabels
}

// StoryURL returns the URL for viewing a story on Lobste.rs
func (c *LobstersClient) StoryURL(item *Item) string {
	if item.Type != "" && item.Type != "story" {
		return fmt.Sprintf("%s/s/%s", lobstersBaseURL, item.Type)
	}
	return item.URL
}

// FetchStoryIDs fetches story "IDs" for a feed
func (c *LobstersClient) FetchStoryIDs(feed string) ([]int, error) {
	var allStories []*Item
	for page := 1; page <= 2; page++ {
		stories, err := c.fetchStoriesPage(feed, page)
		if err != nil {
			if page == 1 {
				return nil, fmt.Errorf("failed to fetch page 1 for feed %q: %w", feed, err)
			}
			break
		}
		allStories = append(allStories, stories...)
	}

	if len(allStories) == 0 {
		return nil, fmt.Errorf("no stories found for feed %q", feed)
	}

	return c.StoreItems(allStories), nil
}

// fetchStoriesPage fetches a single page of stories
func (c *LobstersClient) fetchStoriesPage(feed string, page int) ([]*Item, error) {
	c.Throttle()

	var url string
	if feed == "" {
		if page == 1 {
			url = lobstersBaseURL
		} else {
			url = fmt.Sprintf("%s/page/%d", lobstersBaseURL, page)
		}
	} else {
		if page == 1 {
			url = fmt.Sprintf("%s/%s", lobstersBaseURL, feed)
		} else {
			url = fmt.Sprintf("%s/%s/page/%d", lobstersBaseURL, feed, page)
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; hn-tui/1.0)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lobste.rs page: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting with retry
	if resp.StatusCode == 429 {
		resp.Body.Close()
		time.Sleep(2 * time.Second)
		c.Throttle()
		req, _ = http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; hn-tui/1.0)")
		resp, err = c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch lobste.rs page after retry: %w", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("lobste.rs returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return parseLobstersStories(doc)
}

// FetchCommentTree fetches comments for a story
func (c *LobstersClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	c.Throttle()

	shortID := item.Type
	if shortID == "" || shortID == "story" {
		return nil, fmt.Errorf("no story ID available")
	}

	url := fmt.Sprintf("%s/s/%s", lobstersBaseURL, shortID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; hn-tui/1.0)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch story page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return parseLobstersComments(doc)
}
