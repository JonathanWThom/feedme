package api

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// RSS feed structures for parsing XML

// RSSFeed represents an RSS 2.0 feed
type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel RSSChannel `xml:"channel"`
}

// RSSChannel represents an RSS channel
type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

// RSSItem represents an RSS item
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Author      string `xml:"author"`
	Creator     string `xml:"creator"` // Dublin Core dc:creator
	Source      string `xml:"source"`
}

// AtomFeed represents an Atom feed
type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Entries []AtomEntry `xml:"entry"`
}

// AtomEntry represents an Atom entry
type AtomEntry struct {
	Title     string     `xml:"title"`
	Link      AtomLink   `xml:"link"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
	ID        string     `xml:"id"`
	Author    AtomAuthor `xml:"author"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
}

// AtomLink represents an Atom link
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

// AtomAuthor represents an Atom author
type AtomAuthor struct {
	Name string `xml:"name"`
}

// RSSClient is a base client for RSS-based news sources
type RSSClient struct {
	http        *http.Client
	name        string
	feedURLs    map[string]string // feed name -> URL
	feedNames   []string
	feedLabels  []string
	storyCache  map[int]*Item
	urlToID     map[string]int
	nextID      int
	baseURL     string
	lastRequest time.Time
}

// NewRSSClient creates a new RSS client
func NewRSSClient(name string, baseURL string, feeds map[string]string, feedNames, feedLabels []string) *RSSClient {
	return &RSSClient{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		name:       name,
		baseURL:    baseURL,
		feedURLs:   feeds,
		feedNames:  feedNames,
		feedLabels: feedLabels,
		storyCache: make(map[int]*Item),
		urlToID:    make(map[string]int),
		nextID:     1,
	}
}

// Name returns the display name of the source
func (c *RSSClient) Name() string {
	return c.name
}

// FeedNames returns the available feed names
func (c *RSSClient) FeedNames() []string {
	return c.feedNames
}

// FeedLabels returns the display labels for feeds
func (c *RSSClient) FeedLabels() []string {
	return c.feedLabels
}

// StoryURL returns the URL for viewing a story
func (c *RSSClient) StoryURL(item *Item) string {
	return item.URL
}

// FetchStoryIDs fetches story IDs for a feed
func (c *RSSClient) FetchStoryIDs(feed string) ([]int, error) {
	feedURL, ok := c.feedURLs[feed]
	if !ok {
		return nil, fmt.Errorf("unknown feed: %s", feed)
	}

	items, err := c.fetchRSSFeed(feedURL)
	if err != nil {
		return nil, err
	}

	// Clear cache for new fetch
	c.storyCache = make(map[int]*Item)
	c.urlToID = make(map[string]int)
	c.nextID = 1

	ids := make([]int, len(items))
	for i, item := range items {
		id := c.nextID
		c.nextID++
		c.storyCache[id] = item
		c.urlToID[item.URL] = id
		ids[i] = id
	}

	return ids, nil
}

// fetchRSSFeed fetches and parses an RSS feed
func (c *RSSClient) fetchRSSFeed(url string) ([]*Item, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; feedme/1.0)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("RSS feed returned status %d", resp.StatusCode)
	}

	// Try parsing as RSS first
	var rssFeed RSSFeed
	decoder := xml.NewDecoder(resp.Body)
	decoder.Strict = false
	if err := decoder.Decode(&rssFeed); err == nil && len(rssFeed.Channel.Items) > 0 {
		return c.convertRSSItems(rssFeed.Channel.Items), nil
	}

	// If RSS parsing failed, try refetching and parsing as Atom
	resp2, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refetch feed: %w", err)
	}
	defer resp2.Body.Close()

	var atomFeed AtomFeed
	decoder2 := xml.NewDecoder(resp2.Body)
	decoder2.Strict = false
	if err := decoder2.Decode(&atomFeed); err == nil && len(atomFeed.Entries) > 0 {
		return c.convertAtomEntries(atomFeed.Entries), nil
	}

	return nil, fmt.Errorf("failed to parse feed as RSS or Atom")
}

// convertRSSItems converts RSS items to our Item type
func (c *RSSClient) convertRSSItems(rssItems []RSSItem) []*Item {
	items := make([]*Item, 0, len(rssItems))
	for _, rss := range rssItems {
		item := &Item{
			Title: html.UnescapeString(strings.TrimSpace(rss.Title)),
			URL:   strings.TrimSpace(rss.Link),
			By:    c.extractAuthor(rss),
			Time:  c.parseRSSDate(rss.PubDate),
			Text:  cleanDescription(rss.Description),
			Type:  rss.GUID,
		}
		if item.Title != "" && item.URL != "" {
			items = append(items, item)
		}
	}
	return items
}

// convertAtomEntries converts Atom entries to our Item type
func (c *RSSClient) convertAtomEntries(entries []AtomEntry) []*Item {
	items := make([]*Item, 0, len(entries))
	for _, entry := range entries {
		pubTime := entry.Published
		if pubTime == "" {
			pubTime = entry.Updated
		}

		item := &Item{
			Title: html.UnescapeString(strings.TrimSpace(entry.Title)),
			URL:   strings.TrimSpace(entry.Link.Href),
			By:    entry.Author.Name,
			Time:  c.parseRSSDate(pubTime),
			Text:  cleanDescription(entry.Summary),
			Type:  entry.ID,
		}
		if item.Title != "" && item.URL != "" {
			items = append(items, item)
		}
	}
	return items
}

// extractAuthor extracts author from RSS item
func (c *RSSClient) extractAuthor(rss RSSItem) string {
	if rss.Author != "" {
		return rss.Author
	}
	if rss.Creator != "" {
		return rss.Creator
	}
	if rss.Source != "" {
		return rss.Source
	}
	return ""
}

// parseRSSDate parses various RSS date formats
func (c *RSSClient) parseRSSDate(dateStr string) int64 {
	if dateStr == "" {
		return time.Now().Unix()
	}

	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2 Jan 2006 15:04:05 -0700",
		"2006-01-02 15:04:05",
	}

	dateStr = strings.TrimSpace(dateStr)
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Unix()
		}
	}

	return time.Now().Unix()
}

// FetchItem fetches a cached item by ID
func (c *RSSClient) FetchItem(id int) (*Item, error) {
	if item, ok := c.storyCache[id]; ok {
		return item, nil
	}
	return nil, fmt.Errorf("item %d not found in cache", id)
}

// FetchItems fetches multiple cached items by ID
func (c *RSSClient) FetchItems(ids []int) ([]*Item, error) {
	items := make([]*Item, len(ids))
	for i, id := range ids {
		if item, ok := c.storyCache[id]; ok {
			items[i] = item
		}
	}
	return items, nil
}

// FetchCommentTree returns empty comments (news sites don't have integrated comments)
func (c *RSSClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	// Most news sites don't have comments in their RSS feeds
	return []*Comment{}, nil
}

// cleanDescription removes HTML tags and cleans up description text
func cleanDescription(desc string) string {
	if desc == "" {
		return ""
	}
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(desc, "")
	// Unescape HTML entities
	text = html.UnescapeString(text)
	// Clean up whitespace
	text = strings.Join(strings.Fields(text), " ")
	// Truncate if too long
	if len(text) > 300 {
		text = text[:297] + "..."
	}
	return text
}
