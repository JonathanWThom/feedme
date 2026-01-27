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
	http        *http.Client
	storyCache  map[int]*Item
	cacheMu     sync.RWMutex
	lastRequest time.Time
	requestMu   sync.Mutex
}

// NewLobstersClient creates a new Lobste.rs scraping client
func NewLobstersClient() *LobstersClient {
	return &LobstersClient{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		storyCache: make(map[int]*Item),
	}
}

// throttle ensures we don't make requests too quickly
func (c *LobstersClient) throttle() {
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
	// The item ID is actually a hash of the short ID, we store the short ID in Type field
	if item.Type != "" && item.Type != "story" {
		return fmt.Sprintf("%s/s/%s", lobstersBaseURL, item.Type)
	}
	return item.URL
}

// FetchStoryIDs fetches story "IDs" for a feed
// Since lobste.rs doesn't have numeric IDs, we fetch stories and cache them
// returning sequential pseudo-IDs
func (c *LobstersClient) FetchStoryIDs(feed string) ([]int, error) {
	// Fetch multiple pages worth of stories
	var allStories []*Item
	for page := 1; page <= 2; page++ { // Get 2 pages (~50 stories) to avoid rate limiting
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
func (c *LobstersClient) fetchStoriesPage(feed string, page int) ([]*Item, error) {
	// Throttle requests to avoid rate limiting
	c.throttle()

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
		time.Sleep(2 * time.Second) // Wait 2 seconds and retry
		c.throttle()
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

	return c.parseStories(doc)
}

// parseStories extracts stories from the HTML document
func (c *LobstersClient) parseStories(doc *goquery.Document) ([]*Item, error) {
	var stories []*Item

	doc.Find("ol.stories > li.story").Each(func(i int, s *goquery.Selection) {
		story := c.parseStory(s)
		if story != nil {
			stories = append(stories, story)
		}
	})

	return stories, nil
}

// parseStory extracts a single story from an HTML element
func (c *LobstersClient) parseStory(s *goquery.Selection) *Item {
	item := &Item{
		Type: "story",
	}

	// Get story link and title
	linkSel := s.Find("a.u-url")
	if linkSel.Length() == 0 {
		// Try alternative selector
		linkSel = s.Find(".link a").First()
	}
	if linkSel.Length() > 0 {
		item.Title = strings.TrimSpace(linkSel.Text())
		item.URL, _ = linkSel.Attr("href")
		// If URL is relative, make it absolute
		if strings.HasPrefix(item.URL, "/") {
			item.URL = lobstersBaseURL + item.URL
		}
	}

	// Get short ID from the story class or data attribute
	if shortID, exists := s.Attr("data-shortid"); exists {
		item.Type = shortID // Store short ID in Type field for later use
		// Generate a pseudo-ID from the short ID
		item.ID = hashShortID(shortID)
	}

	// Get score/votes - lobste.rs uses a.upvoter with the score as text
	scoreSel := s.Find(".voters a.upvoter")
	if scoreSel.Length() > 0 {
		scoreText := strings.TrimSpace(scoreSel.Text())
		if score, err := strconv.Atoi(scoreText); err == nil {
			item.Score = score
		}
	}

	// Get author
	authorSel := s.Find(".byline a.u-author")
	if authorSel.Length() == 0 {
		authorSel = s.Find(".byline a[href^='/~']")
	}
	if authorSel.Length() > 0 {
		item.By = strings.TrimSpace(authorSel.Text())
	}

	// Get time from the time element
	timeSel := s.Find(".byline time")
	if timeSel.Length() > 0 {
		// Try to get the title attribute which has the full timestamp
		if title, exists := timeSel.Attr("title"); exists {
			if t, err := parseTime(title); err == nil {
				item.Time = t.Unix()
			}
		}
		// If no title, try to parse the text
		if item.Time == 0 {
			timeText := strings.TrimSpace(timeSel.Text())
			item.Time = parseRelativeTime(timeText)
		}
	}

	// Get comment count from span.comments_label
	commentSel := s.Find(".comments_label")
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

	// Get tags (store in Text field for now, as a comma-separated list)
	var tags []string
	s.Find(".tags a.tag").Each(func(i int, tagSel *goquery.Selection) {
		tag := strings.TrimSpace(tagSel.Text())
		if tag != "" {
			tags = append(tags, tag)
		}
	})
	if len(tags) > 0 {
		item.Text = "[" + strings.Join(tags, ", ") + "]"
	}

	// Skip if we don't have a title
	if item.Title == "" {
		return nil
	}

	return item
}

// FetchItem fetches a cached item by pseudo-ID
func (c *LobstersClient) FetchItem(id int) (*Item, error) {
	c.cacheMu.RLock()
	item, ok := c.storyCache[id]
	c.cacheMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("item %d not found in cache", id)
	}

	return item, nil
}

// FetchItems fetches multiple cached items by pseudo-ID
func (c *LobstersClient) FetchItems(ids []int) ([]*Item, error) {
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
func (c *LobstersClient) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	// Throttle requests to avoid rate limiting
	c.throttle()

	// Get the short ID from the Type field
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

	return c.parseComments(doc)
}

// parseComments extracts comments from a story page
func (c *LobstersClient) parseComments(doc *goquery.Document) ([]*Comment, error) {
	var comments []*Comment

	// Comments are in div.comment elements within li.comments_subtree
	doc.Find("div.comment[data-shortid]").Each(func(i int, s *goquery.Selection) {
		comment := c.parseComment(s)
		if comment != nil {
			comments = append(comments, comment)
		}
	})

	return comments, nil
}

// parseComment extracts a single comment
func (c *LobstersClient) parseComment(s *goquery.Selection) *Comment {
	item := &Item{
		Type: "comment",
	}

	// Get author from byline
	authorSel := s.Find(".byline a[href^='/~']")
	if authorSel.Length() > 0 {
		item.By = strings.TrimSpace(authorSel.Text())
	}

	// Get comment text
	textSel := s.Find(".comment_text")
	if textSel.Length() > 0 {
		html, _ := textSel.Html()
		item.Text = html
	}

	// Get time from time element
	timeSel := s.Find(".byline time")
	if timeSel.Length() > 0 {
		if title, exists := timeSel.Attr("title"); exists {
			if t, err := parseTime(title); err == nil {
				item.Time = t.Unix()
			}
		}
		if item.Time == 0 {
			timeText := strings.TrimSpace(timeSel.Text())
			item.Time = parseRelativeTime(timeText)
		}
	}

	// Get depth by counting parent ol.comments elements
	depth := 0
	s.Parents().Each(func(i int, parent *goquery.Selection) {
		if parent.Is("ol.comments") {
			depth++
		}
	})
	// Subtract 1 because the top-level comments list counts as 1
	if depth > 0 {
		depth--
	}

	// Get score from upvoter
	scoreSel := s.Find(".voters a.upvoter")
	if scoreSel.Length() > 0 {
		scoreText := strings.TrimSpace(scoreSel.Text())
		if score, err := strconv.Atoi(scoreText); err == nil {
			item.Score = score
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

// Helper functions

// hashShortID creates a pseudo-numeric ID from a short ID string
func hashShortID(shortID string) int {
	hash := 0
	for i, c := range shortID {
		hash += int(c) * (i + 1) * 31
	}
	return hash
}

// parseTime attempts to parse a timestamp string
func parseTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05 -0700",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// parseRelativeTime converts relative time strings to Unix timestamp
func parseRelativeTime(s string) int64 {
	now := time.Now()
	s = strings.ToLower(strings.TrimSpace(s))

	re := regexp.MustCompile(`(\d+)\s*(second|minute|hour|day|week|month|year)s?\s*ago`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 3 {
		return now.Unix()
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return now.Unix()
	}

	var duration time.Duration
	switch matches[2] {
	case "second":
		duration = time.Duration(num) * time.Second
	case "minute":
		duration = time.Duration(num) * time.Minute
	case "hour":
		duration = time.Duration(num) * time.Hour
	case "day":
		duration = time.Duration(num) * 24 * time.Hour
	case "week":
		duration = time.Duration(num) * 7 * 24 * time.Hour
	case "month":
		duration = time.Duration(num) * 30 * 24 * time.Hour
	case "year":
		duration = time.Duration(num) * 365 * 24 * time.Hour
	}

	return now.Add(-duration).Unix()
}
