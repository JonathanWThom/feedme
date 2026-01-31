package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// AP News doesn't have public RSS feeds, so we scrape their website
// AP provides straight wire news with minimal opinion

const apBaseURL = "https://apnews.com"

var apFeedPaths = map[string]string{
	"top":       "",
	"world":     "/world-news",
	"us":        "/us-news",
	"politics":  "/politics",
	"business":  "/business",
	"tech":      "/technology",
	"science":   "/science",
	"health":    "/health",
	"sports":    "/sports",
	"entertain": "/entertainment",
}

var apFeedNames = []string{"top", "world", "us", "politics", "business", "tech", "science", "health", "sports", "entertain"}
var apFeedLabels = []string{"Top", "World", "US", "Politics", "Business", "Tech", "Science", "Health", "Sports", "Arts"}

// APNewsClient scrapes AP News
type APNewsClient struct {
	http        *http.Client
	storyCache  map[int]*Item
	cacheMu     sync.RWMutex
	lastRequest time.Time
	requestMu   sync.Mutex
}

// NewAPNewsClient creates a new AP News scraping client
func NewAPNewsClient() *APNewsClient {
	return &APNewsClient{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		storyCache: make(map[int]*Item),
	}
}

// throttle ensures we don't make requests too quickly
func (c *APNewsClient) throttle() {
	c.requestMu.Lock()
	defer c.requestMu.Unlock()

	minDelay := 500 * time.Millisecond
	elapsed := time.Since(c.lastRequest)
	if elapsed < minDelay {
		time.Sleep(minDelay - elapsed)
	}
	c.lastRequest = time.Now()
}

// Name returns the display name of the source
func (c *APNewsClient) Name() string {
	return "AP News"
}

// FeedNames returns the available feed names
func (c *APNewsClient) FeedNames() []string {
	return apFeedNames
}

// FeedLabels returns the display labels for feeds
func (c *APNewsClient) FeedLabels() []string {
	return apFeedLabels
}

// StoryURL returns the URL for viewing a story on AP News
func (c *APNewsClient) StoryURL(item *Item) string {
	return item.URL
}

// FetchStoryIDs fetches story IDs for a feed
func (c *APNewsClient) FetchStoryIDs(feed string) ([]int, error) {
	path, ok := apFeedPaths[feed]
	if !ok {
		return nil, fmt.Errorf("unknown feed: %s", feed)
	}

	stories, err := c.fetchStories(path)
	if err != nil {
		return nil, err
	}

	// Cache stories and return IDs
	ids := make([]int, len(stories))
	c.cacheMu.Lock()
	c.storyCache = make(map[int]*Item)
	for i, story := range stories {
		id := i + 1
		c.storyCache[id] = story
		ids[i] = id
	}
	c.cacheMu.Unlock()

	return ids, nil
}

// fetchStories fetches stories from AP News
func (c *APNewsClient) fetchStories(path string) ([]*Item, error) {
	c.throttle()

	url := apBaseURL + path

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; feedme/1.0)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AP News: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("AP News returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return c.parseStories(doc)
}

// parseStories extracts stories from the HTML document
func (c *APNewsClient) parseStories(doc *goquery.Document) ([]*Item, error) {
	var stories []*Item
	seen := make(map[string]bool)

	// AP News uses various selectors for story cards
	selectors := []string{
		"a[data-key='card-headline']",
		"a.Link[href*='/article/']",
		"div.PagePromo a[href*='/article/']",
		"h2 a[href*='/article/']",
		"h3 a[href*='/article/']",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists {
				return
			}

			// Make URL absolute
			if strings.HasPrefix(href, "/") {
				href = apBaseURL + href
			}

			// Skip if already seen
			if seen[href] {
				return
			}
			seen[href] = true

			// Get title
			title := strings.TrimSpace(s.Text())
			if title == "" {
				title = strings.TrimSpace(s.Find("span").Text())
			}
			if title == "" {
				return
			}

			story := &Item{
				Title: title,
				URL:   href,
				By:    "AP",
				Time:  time.Now().Unix(),
				Type:  href,
			}

			stories = append(stories, story)
		})
	}

	// Limit to reasonable number
	if len(stories) > 50 {
		stories = stories[:50]
	}

	return stories, nil
}

// FetchItem fetches a cached item by ID
func (c *APNewsClient) FetchItem(id int) (*Item, error) {
	c.cacheMu.RLock()
	item, ok := c.storyCache[id]
	c.cacheMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("item %d not found in cache", id)
	}

	return item, nil
}

// FetchItems fetches multiple cached items by ID
func (c *APNewsClient) FetchItems(ids []int) ([]*Item, error) {
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

// FetchCommentTree returns empty comments (news sites don't have integrated comments)
func (c *APNewsClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	return []*Comment{}, nil
}
