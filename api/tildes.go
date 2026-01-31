package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const tildesBaseURL = "https://tildes.net"

// Tildes feed types (correspond to URL query parameters)
const (
	TildesFeedActivity = ""         // Default - recently active
	TildesFeedNew      = "new"      // Newest topics
	TildesFeedVotes    = "votes"    // Highest voted
	TildesFeedComments = "comments" // Most comments
)

var TildesFeedNames = []string{TildesFeedActivity, TildesFeedNew, TildesFeedVotes, TildesFeedComments}
var TildesFeedLabels = []string{"Activity", "New", "Votes", "Comments"}

// TildesClient scrapes tildes.net
type TildesClient struct {
	http        *http.Client
	storyCache  map[int]*Item
	cacheMu     sync.RWMutex
	lastRequest time.Time
	requestMu   sync.Mutex
	group       string // Optional group filter (e.g., "~tech")
}

// NewTildesClient creates a new Tildes scraping client
func NewTildesClient() *TildesClient {
	return &TildesClient{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		storyCache: make(map[int]*Item),
	}
}

// NewTildesClientWithGroup creates a new Tildes client for a specific group
func NewTildesClientWithGroup(group string) *TildesClient {
	// Normalize group name (remove ~ prefix if present)
	group = strings.TrimPrefix(group, "~")
	return &TildesClient{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		storyCache: make(map[int]*Item),
		group:      group,
	}
}

// throttle ensures we don't make requests too quickly
func (c *TildesClient) throttle() {
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
func (c *TildesClient) Name() string {
	if c.group != "" {
		return "Tildes ~" + c.group
	}
	return "Tildes"
}

// FeedNames returns the available feed names
func (c *TildesClient) FeedNames() []string {
	return TildesFeedNames
}

// FeedLabels returns the display labels for feeds
func (c *TildesClient) FeedLabels() []string {
	return TildesFeedLabels
}

// StoryURL returns the URL for viewing a story on Tildes
func (c *TildesClient) StoryURL(item *Item) string {
	// The topic ID36 is stored in the Type field
	if item.Type != "" && item.Type != "story" {
		return fmt.Sprintf("%s%s", tildesBaseURL, item.Type)
	}
	return item.URL
}

// FetchStoryIDs fetches story "IDs" for a feed
// Since Tildes doesn't have numeric IDs, we fetch stories and cache them
// returning sequential pseudo-IDs
func (c *TildesClient) FetchStoryIDs(feed string) ([]int, error) {
	// Fetch multiple pages worth of stories
	var allStories []*Item
	for page := 1; page <= 2; page++ { // Get 2 pages (~50 stories)
		stories, err := c.fetchStoriesPage(feed, page)
		if err != nil {
			if page == 1 {
				return nil, fmt.Errorf("failed to fetch page 1 for feed %q: %w", feed, err)
			}
			break // If we got at least page 1, continue with what we have
		}
		allStories = append(allStories, stories...)
	}

	if len(allStories) == 0 {
		return nil, fmt.Errorf("no stories found for feed %q", feed)
	}

	// Cache stories and return pseudo-IDs
	ids := make([]int, len(allStories))
	c.cacheMu.Lock()
	// Clear old cache
	c.storyCache = make(map[int]*Item)
	for i, story := range allStories {
		id := i + 1 // 1-indexed pseudo-IDs
		c.storyCache[id] = story
		ids[i] = id
	}
	c.cacheMu.Unlock()

	return ids, nil
}

// fetchStoriesPage fetches a single page of stories
func (c *TildesClient) fetchStoriesPage(feed string, page int) ([]*Item, error) {
	// Throttle requests to avoid rate limiting
	c.throttle()

	var url string
	baseURL := tildesBaseURL
	if c.group != "" {
		baseURL = fmt.Sprintf("%s/~%s", tildesBaseURL, c.group)
	}

	// Build URL with feed sorting and pagination
	params := []string{}
	if feed != "" {
		params = append(params, fmt.Sprintf("order=%s", feed))
	}
	if page > 1 {
		params = append(params, fmt.Sprintf("page=%d", page))
	}

	if len(params) > 0 {
		url = fmt.Sprintf("%s?%s", baseURL, strings.Join(params, "&"))
	} else {
		url = baseURL
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; feedme/1.0)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tildes page: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting with retry
	if resp.StatusCode == 429 {
		resp.Body.Close()
		time.Sleep(2 * time.Second) // Wait 2 seconds and retry
		c.throttle()
		req, _ = http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; feedme/1.0)")
		resp, err = c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tildes page after retry: %w", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("tildes returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return c.parseStories(doc)
}

// parseStories extracts stories from the HTML document
func (c *TildesClient) parseStories(doc *goquery.Document) ([]*Item, error) {
	var stories []*Item

	// Topics are in <article> elements within the topic-listing
	doc.Find("article.topic").Each(func(i int, s *goquery.Selection) {
		story := c.parseStory(s)
		if story != nil {
			stories = append(stories, story)
		}
	})

	return stories, nil
}

// parseStory extracts a single story from an HTML element
func (c *TildesClient) parseStory(s *goquery.Selection) *Item {
	item := &Item{
		Type: "story",
	}

	// Get topic ID from the article id attribute (format: topic-{id36})
	if id, exists := s.Attr("id"); exists {
		if strings.HasPrefix(id, "topic-") {
			topicID36 := strings.TrimPrefix(id, "topic-")
			// Store the full path for linking
			if path, exists := s.Find("a.topic-title").Attr("href"); exists {
				item.Type = path // Store path like "/~tech/1abc/topic-title"
			} else {
				item.Type = topicID36
			}
			item.ID = hashShortID(topicID36)
		}
	}

	// Get title
	titleSel := s.Find("a.topic-title, h1.topic-title a")
	if titleSel.Length() > 0 {
		item.Title = strings.TrimSpace(titleSel.Text())
	}

	// Get URL - for link topics, the title links to external URL
	// For text topics, the title links to the topic itself
	if href, exists := titleSel.Attr("href"); exists {
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			item.URL = href
		} else {
			// It's an internal link (text post)
			item.URL = tildesBaseURL + href
		}
	}

	// Get vote count
	voteSel := s.Find(".topic-voting-votes")
	if voteSel.Length() > 0 {
		voteText := strings.TrimSpace(voteSel.Text())
		if votes, err := strconv.Atoi(voteText); err == nil {
			item.Score = votes
		}
	}

	// Get comment count
	commentSel := s.Find(".topic-info-comments a")
	if commentSel.Length() > 0 {
		commentText := strings.TrimSpace(commentSel.Text())
		// Extract number from text like "17 comments"
		re := regexp.MustCompile(`(\d+)`)
		if matches := re.FindStringSubmatch(commentText); len(matches) > 1 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				item.Descendants = count
			}
		}
	}

	// Get author from data attribute or topic-info-source
	if author, exists := s.Attr("data-topic-posted-by"); exists {
		item.By = author
	} else {
		authorSel := s.Find(".topic-info-source a")
		if authorSel.Length() > 0 {
			item.By = strings.TrimSpace(authorSel.Text())
		}
	}

	// Get timestamp from time element
	timeSel := s.Find("time")
	if timeSel.Length() > 0 {
		// Try datetime attribute first
		if datetime, exists := timeSel.Attr("datetime"); exists {
			if t, err := time.Parse(time.RFC3339, datetime); err == nil {
				item.Time = t.Unix()
			}
		}
		// Fallback to title attribute
		if item.Time == 0 {
			if title, exists := timeSel.Attr("title"); exists {
				if t, err := parseTime(title); err == nil {
					item.Time = t.Unix()
				}
			}
		}
		// Last resort: parse relative time from text
		if item.Time == 0 {
			timeText := strings.TrimSpace(timeSel.Text())
			item.Time = parseRelativeTime(timeText)
		}
	}

	// Get group
	groupSel := s.Find(".topic-group a")
	if groupSel.Length() > 0 {
		group := strings.TrimSpace(groupSel.Text())
		if group != "" {
			item.Text = "[" + group + "]"
		}
	}

	// Get tags (append to group if present)
	var tags []string
	s.Find(".topic-tags a.label-topic-tag").Each(func(i int, tagSel *goquery.Selection) {
		tag := strings.TrimSpace(tagSel.Text())
		if tag != "" {
			tags = append(tags, tag)
		}
	})
	if len(tags) > 0 {
		if item.Text != "" {
			item.Text += " "
		}
		item.Text += strings.Join(tags, ", ")
	}

	// Skip if we don't have a title
	if item.Title == "" {
		return nil
	}

	return item
}

// FetchItem fetches a cached item by pseudo-ID
func (c *TildesClient) FetchItem(id int) (*Item, error) {
	c.cacheMu.RLock()
	item, ok := c.storyCache[id]
	c.cacheMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("item %d not found in cache", id)
	}

	return item, nil
}

// FetchItems fetches multiple cached items by pseudo-ID
func (c *TildesClient) FetchItems(ids []int) ([]*Item, error) {
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

// FetchCommentTree fetches comments for a story
func (c *TildesClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	// Throttle requests to avoid rate limiting
	c.throttle()

	// Get the topic path from the Type field
	topicPath := item.Type
	if topicPath == "" || topicPath == "story" {
		return nil, fmt.Errorf("no topic ID available")
	}

	var url string
	if strings.HasPrefix(topicPath, "/") {
		url = tildesBaseURL + topicPath
	} else {
		// If it's just an ID36, we need to find the topic page somehow
		// This shouldn't happen with our implementation
		return nil, fmt.Errorf("cannot fetch comments without full topic path")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; feedme/1.0)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch topic page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("tildes returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return c.parseComments(doc)
}

// parseComments extracts comments from a topic page
func (c *TildesClient) parseComments(doc *goquery.Document) ([]*Comment, error) {
	var comments []*Comment

	// Comments are in article.comment elements
	doc.Find("article.comment").Each(func(i int, s *goquery.Selection) {
		comment := c.parseComment(s)
		if comment != nil {
			comments = append(comments, comment)
		}
	})

	return comments, nil
}

// parseComment extracts a single comment
func (c *TildesClient) parseComment(s *goquery.Selection) *Comment {
	item := &Item{
		Type: "comment",
	}

	// Get depth from data-comment-depth attribute
	depth := 0
	if depthStr, exists := s.Attr("data-comment-depth"); exists {
		if d, err := strconv.Atoi(depthStr); err == nil {
			depth = d
		}
	}

	// Get author from comment-header
	authorSel := s.Find(".comment-header a.link-user")
	if authorSel.Length() > 0 {
		item.By = strings.TrimSpace(authorSel.Text())
	}

	// Get comment text
	textSel := s.Find(".comment-text")
	if textSel.Length() > 0 {
		html, _ := textSel.Html()
		item.Text = html
	}

	// Get timestamp from time element in header
	timeSel := s.Find(".comment-posted-time time, time.comment-posted-time")
	if timeSel.Length() > 0 {
		// Try datetime attribute first
		if datetime, exists := timeSel.Attr("datetime"); exists {
			if t, err := time.Parse(time.RFC3339, datetime); err == nil {
				item.Time = t.Unix()
			}
		}
		// Fallback to title attribute
		if item.Time == 0 {
			if title, exists := timeSel.Attr("title"); exists {
				if t, err := parseTime(title); err == nil {
					item.Time = t.Unix()
				}
			}
		}
		// Last resort: parse relative time from text
		if item.Time == 0 {
			timeText := strings.TrimSpace(timeSel.Text())
			item.Time = parseRelativeTime(timeText)
		}
	}

	// Get vote count from comment-votes
	voteSel := s.Find(".comment-votes")
	if voteSel.Length() > 0 {
		voteText := strings.TrimSpace(voteSel.Text())
		// Extract number, format might be "3 votes" or just "3"
		re := regexp.MustCompile(`(\d+)`)
		if matches := re.FindStringSubmatch(voteText); len(matches) > 1 {
			if votes, err := strconv.Atoi(matches[1]); err == nil {
				item.Score = votes
			}
		}
	}

	if item.By == "" && item.Text == "" {
		return nil
	}

	return &Comment{
		Item:  item,
		Depth: depth,
	}
}
