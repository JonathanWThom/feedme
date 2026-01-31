package api

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// RSSClient fetches and parses RSS/Atom feeds
type RSSClient struct {
	feedURL    string
	feedTitle  string
	http       *http.Client
	storyCache map[int]*Item
	cacheMu    sync.RWMutex
}

// NewRSSClient creates a new RSS feed client
func NewRSSClient(feedURL string) *RSSClient {
	return &RSSClient{
		feedURL: feedURL,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		storyCache: make(map[int]*Item),
	}
}

// Name returns the display name of the source (feed title or domain)
func (c *RSSClient) Name() string {
	if c.feedTitle != "" {
		// Truncate long titles
		if len(c.feedTitle) > 20 {
			return c.feedTitle[:17] + "..."
		}
		return c.feedTitle
	}
	// Fall back to domain name
	if u, err := url.Parse(c.feedURL); err == nil {
		return u.Host
	}
	return "RSS"
}

// FeedNames returns the available feed names (just one for RSS)
func (c *RSSClient) FeedNames() []string {
	return []string{"feed"}
}

// FeedLabels returns the display labels for feeds
func (c *RSSClient) FeedLabels() []string {
	return []string{"Items"}
}

// StoryURL returns the URL for viewing a story
func (c *RSSClient) StoryURL(item *Item) string {
	// For RSS items, the URL field already contains the link
	if item.URL != "" {
		return item.URL
	}
	// Fall back to feed URL if no item URL
	return c.feedURL
}

// FetchStoryIDs fetches the RSS feed and returns pseudo-IDs
func (c *RSSClient) FetchStoryIDs(feed string) ([]int, error) {
	req, err := http.NewRequest("GET", c.feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "feedme/1.0 (Terminal RSS Reader)")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("RSS feed returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS feed: %w", err)
	}

	items, title, err := c.parseFeed(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	if title != "" {
		c.feedTitle = title
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no items found in RSS feed")
	}

	// Cache items and return pseudo-IDs
	ids := make([]int, len(items))
	c.cacheMu.Lock()
	c.storyCache = make(map[int]*Item)
	for i, item := range items {
		id := i + 1 // 1-indexed pseudo-IDs
		c.storyCache[id] = item
		ids[i] = id
	}
	c.cacheMu.Unlock()

	return ids, nil
}

// FetchItem fetches a cached item by pseudo-ID
func (c *RSSClient) FetchItem(id int) (*Item, error) {
	c.cacheMu.RLock()
	item, ok := c.storyCache[id]
	c.cacheMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("item %d not found in cache", id)
	}

	return item, nil
}

// FetchItems fetches multiple cached items by pseudo-ID
func (c *RSSClient) FetchItems(ids []int) ([]*Item, error) {
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

// FetchCommentTree returns empty for RSS (no comments support)
func (c *RSSClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	// RSS feeds don't have comments
	return nil, nil
}

// parseFeed attempts to parse the feed as RSS 2.0, then Atom
func (c *RSSClient) parseFeed(data []byte) ([]*Item, string, error) {
	// Try RSS 2.0 first
	items, title, err := c.parseRSS(data)
	if err == nil && len(items) > 0 {
		return items, title, nil
	}

	// Try Atom
	items, title, err = c.parseAtom(data)
	if err == nil && len(items) > 0 {
		return items, title, nil
	}

	return nil, "", fmt.Errorf("unable to parse feed as RSS or Atom")
}

// RSS 2.0 structures
type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	Creator     string `xml:"creator"` // dc:creator
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Comments    string `xml:"comments"`
}

func (c *RSSClient) parseRSS(data []byte) ([]*Item, string, error) {
	var feed rssFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, "", err
	}

	if len(feed.Channel.Items) == 0 {
		return nil, "", fmt.Errorf("no items in RSS feed")
	}

	items := make([]*Item, 0, len(feed.Channel.Items))
	for _, rssItem := range feed.Channel.Items {
		item := &Item{
			Title: cleanText(rssItem.Title),
			URL:   rssItem.Link,
			By:    getAuthor(rssItem.Author, rssItem.Creator),
			Time:  parseRSSTime(rssItem.PubDate),
			Text:  cleanText(rssItem.Description),
			Type:  "story",
		}

		// Skip items without title
		if item.Title == "" {
			continue
		}

		items = append(items, item)
	}

	return items, cleanText(feed.Channel.Title), nil
}

// Atom structures
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	Author    atomAuthor `xml:"author"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
	ID        string     `xml:"id"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

func (c *RSSClient) parseAtom(data []byte) ([]*Item, string, error) {
	var feed atomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, "", err
	}

	if len(feed.Entries) == 0 {
		return nil, "", fmt.Errorf("no entries in Atom feed")
	}

	items := make([]*Item, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		item := &Item{
			Title: cleanText(entry.Title),
			URL:   getAtomLink(entry.Links),
			By:    entry.Author.Name,
			Time:  parseAtomTime(entry.Published, entry.Updated),
			Text:  cleanText(firstNonEmpty(entry.Summary, entry.Content)),
			Type:  "story",
		}

		// Skip items without title
		if item.Title == "" {
			continue
		}

		items = append(items, item)
	}

	return items, cleanText(feed.Title), nil
}

// Helper functions

func cleanText(s string) string {
	// Unescape HTML entities
	s = html.UnescapeString(s)
	// Trim whitespace
	s = strings.TrimSpace(s)
	return s
}

func getAuthor(author, creator string) string {
	if author != "" {
		return author
	}
	return creator
}

func getAtomLink(links []atomLink) string {
	// Prefer alternate link, fall back to first link
	for _, link := range links {
		if link.Rel == "alternate" || link.Rel == "" {
			return link.Href
		}
	}
	if len(links) > 0 {
		return links[0].Href
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseRSSTime(s string) int64 {
	if s == "" {
		return time.Now().Unix()
	}

	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t.Unix()
		}
	}

	return time.Now().Unix()
}

func parseAtomTime(published, updated string) int64 {
	s := published
	if s == "" {
		s = updated
	}
	if s == "" {
		return time.Now().Unix()
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t.Unix()
		}
	}

	return time.Now().Unix()
}
