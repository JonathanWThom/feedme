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
