package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const baseURL = "https://hacker-news.firebaseio.com/v0"

// Feed types
const (
	FeedTop  = "topstories"
	FeedNew  = "newstories"
	FeedBest = "beststories"
	FeedAsk  = "askstories"
	FeedShow = "showstories"
)

var FeedNames = []string{FeedTop, FeedNew, FeedBest, FeedAsk, FeedShow}

// Item represents a HN item (story, comment, job, poll)
type Item struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Text        string `json:"text"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Score       int    `json:"score"`
	Kids        []int  `json:"kids"`
	Parent      int    `json:"parent"`
	Descendants int    `json:"descendants"`
	Deleted     bool   `json:"deleted"`
	Dead        bool   `json:"dead"`
}

// TimeAgo returns a human-readable time ago string
func (i *Item) TimeAgo() string {
	t := time.Unix(i.Time, 0)
	d := time.Since(t)

	switch {
	case d.Hours() >= 24*365:
		years := int(d.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	case d.Hours() >= 24*30:
		months := int(d.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	case d.Hours() >= 24:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case d.Hours() >= 1:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case d.Minutes() >= 1:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	default:
		return "just now"
	}
}

// Domain extracts the domain from the URL
func (i *Item) Domain() string {
	if i.URL == "" {
		return ""
	}
	// Simple domain extraction
	url := i.URL
	// Remove protocol
	if len(url) > 8 && url[:8] == "https://" {
		url = url[8:]
	} else if len(url) > 7 && url[:7] == "http://" {
		url = url[7:]
	}
	// Remove www.
	if len(url) > 4 && url[:4] == "www." {
		url = url[4:]
	}
	// Get domain only
	for i, c := range url {
		if c == '/' {
			return url[:i]
		}
	}
	return url
}

// Client is the HN API client
type Client struct {
	http *http.Client
}

// NewClient creates a new HN API client
func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the display name of the source
func (c *Client) Name() string {
	return "HN"
}

// FeedNames returns the available feed names
func (c *Client) FeedNames() []string {
	return FeedNames
}

// FeedLabels returns the display labels for feeds
func (c *Client) FeedLabels() []string {
	return []string{"Top", "New", "Best", "Ask", "Show"}
}

// StoryURL returns the URL for viewing a story on HN
func (c *Client) StoryURL(item *Item) string {
	return fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID)
}

// FetchStoryIDs fetches the list of story IDs for a given feed
func (c *Client) FetchStoryIDs(feed string) ([]int, error) {
	url := fmt.Sprintf("%s/%s.json", baseURL, feed)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", feed, err)
	}
	defer resp.Body.Close()

	var ids []int
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", feed, err)
	}

	return ids, nil
}

// FetchItem fetches a single item by ID
func (c *Client) FetchItem(id int) (*Item, error) {
	url := fmt.Sprintf("%s/item/%d.json", baseURL, id)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item %d: %w", id, err)
	}
	defer resp.Body.Close()

	var item Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode item %d: %w", id, err)
	}

	return &item, nil
}

// FetchItems fetches multiple items concurrently
func (c *Client) FetchItems(ids []int) ([]*Item, error) {
	items := make([]*Item, len(ids))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	// Limit concurrency
	sem := make(chan struct{}, 10)

	for i, id := range ids {
		wg.Add(1)
		go func(idx, itemID int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			item, err := c.FetchItem(itemID)
			mu.Lock()
			if err != nil && firstErr == nil {
				firstErr = err
			}
			if item != nil {
				items[idx] = item
			}
			mu.Unlock()
		}(i, id)
	}

	wg.Wait()
	return items, firstErr
}

// FetchComments fetches all comments for a story recursively
func (c *Client) FetchComments(item *Item) ([]*Item, error) {
	if len(item.Kids) == 0 {
		return nil, nil
	}

	var allComments []*Item
	comments, err := c.FetchItems(item.Kids)
	if err != nil {
		return nil, err
	}

	for _, comment := range comments {
		if comment == nil || comment.Deleted || comment.Dead {
			continue
		}
		allComments = append(allComments, comment)
	}

	return allComments, nil
}

// Comment represents a comment with its depth for rendering
type Comment struct {
	*Item
	Depth    int
	Children []*Comment
}

// FetchCommentTree fetches the full comment tree for a story
func (c *Client) FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error) {
	return c.fetchCommentsRecursive(item.Kids, 0, maxDepth)
}

func (c *Client) fetchCommentsRecursive(ids []int, depth, maxDepth int) ([]*Comment, error) {
	if len(ids) == 0 || (maxDepth > 0 && depth >= maxDepth) {
		return nil, nil
	}

	items, err := c.FetchItems(ids)
	if err != nil {
		return nil, err
	}

	var comments []*Comment
	for _, item := range items {
		if item == nil || item.Deleted || item.Dead {
			continue
		}

		comment := &Comment{
			Item:  item,
			Depth: depth,
		}

		// Fetch children
		if len(item.Kids) > 0 {
			children, err := c.fetchCommentsRecursive(item.Kids, depth+1, maxDepth)
			if err != nil {
				// Continue even if some children fail
				continue
			}
			comment.Children = children
		}

		comments = append(comments, comment)
	}

	return comments, nil
}
