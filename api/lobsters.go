package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const lobstersBaseURL = "https://lobste.rs"
const lobstersUserAgent = "Mozilla/5.0 (compatible; hn-tui/1.0)"

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
	doc, err := c.fetchDocument(lobstersPageURL(feed, page))
	if err != nil {
		return nil, err
	}
	return parseLobstersStories(doc)
}

func lobstersPageURL(feed string, page int) string {
	if feed == "" && page == 1 {
		return lobstersBaseURL
	}
	if feed == "" {
		return fmt.Sprintf("%s/page/%d", lobstersBaseURL, page)
	}
	if page == 1 {
		return fmt.Sprintf("%s/%s", lobstersBaseURL, feed)
	}
	return fmt.Sprintf("%s/%s/page/%d", lobstersBaseURL, feed, page)
}

// FetchCommentTree fetches comments for a story
func (c *LobstersClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	c.Throttle()

	shortID := item.Type
	if shortID == "" || shortID == "story" {
		return nil, fmt.Errorf("no story ID available")
	}

	url := fmt.Sprintf("%s/s/%s", lobstersBaseURL, shortID)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, err
	}
	return parseLobstersComments(doc)
}

func (c *LobstersClient) fetchDocument(url string) (*goquery.Document, error) {
	resp, err := doWithRetry(c.http, url, lobstersUserAgent, &c.CachedSource)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()
	return goquery.NewDocumentFromReader(resp.Body)
}
