package api

import (
	"testing"
	"time"
)

// TestHackerNewsClientInterface verifies that Client implements Source interface
func TestHackerNewsClientInterface(t *testing.T) {
	var _ Source = (*Client)(nil)
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.http == nil {
		t.Fatal("HTTP client is nil")
	}
}

func TestClientName(t *testing.T) {
	client := NewClient()
	name := client.Name()
	if name != "HN" {
		t.Errorf("expected name 'HN', got '%s'", name)
	}
}

func TestClientFeedNames(t *testing.T) {
	client := NewClient()
	feedNames := client.FeedNames()

	expected := []string{FeedTop, FeedNew, FeedBest, FeedAsk, FeedShow}
	if len(feedNames) != len(expected) {
		t.Errorf("expected %d feed names, got %d", len(expected), len(feedNames))
	}

	for i, name := range expected {
		if feedNames[i] != name {
			t.Errorf("expected feed name %s at index %d, got %s", name, i, feedNames[i])
		}
	}
}

func TestClientFeedLabels(t *testing.T) {
	client := NewClient()
	labels := client.FeedLabels()

	expected := []string{"Top", "New", "Best", "Ask", "Show"}
	if len(labels) != len(expected) {
		t.Errorf("expected %d labels, got %d", len(expected), len(labels))
	}

	for i, label := range expected {
		if labels[i] != label {
			t.Errorf("expected label %s at index %d, got %s", label, i, labels[i])
		}
	}
}

func TestClientStoryURL(t *testing.T) {
	client := NewClient()
	item := &Item{ID: 12345}
	url := client.StoryURL(item)

	expected := "https://news.ycombinator.com/item?id=12345"
	if url != expected {
		t.Errorf("expected URL '%s', got '%s'", expected, url)
	}
}

// TestFetchStoryIDs tests fetching story IDs from real HN API
func TestFetchStoryIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	testCases := []string{FeedTop, FeedNew, FeedBest, FeedAsk, FeedShow}
	for _, feed := range testCases {
		t.Run(feed, func(t *testing.T) {
			ids, err := client.FetchStoryIDs(feed)
			if err != nil {
				t.Fatalf("FetchStoryIDs failed for %s: %v", feed, err)
			}
			if len(ids) == 0 {
				t.Errorf("expected non-empty story IDs for feed %s", feed)
			}
			// HN typically returns up to 500 IDs
			if len(ids) > 500 {
				t.Errorf("unexpected number of IDs: %d", len(ids))
			}
			// Verify IDs are positive
			for _, id := range ids[:min(10, len(ids))] {
				if id <= 0 {
					t.Errorf("invalid story ID: %d", id)
				}
			}
		})
	}
}

// TestFetchItem tests fetching a single item from real HN API
func TestFetchItem(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// First get some story IDs
	ids, err := client.FetchStoryIDs(FeedTop)
	if err != nil {
		t.Fatalf("FetchStoryIDs failed: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("no story IDs returned")
	}

	// Fetch the first item
	item, err := client.FetchItem(ids[0])
	if err != nil {
		t.Fatalf("FetchItem failed: %v", err)
	}
	if item == nil {
		t.Fatal("FetchItem returned nil")
	}

	// Validate item fields
	if item.ID == 0 {
		t.Error("item ID is 0")
	}
	if item.Title == "" {
		t.Error("item title is empty")
	}
	if item.By == "" {
		t.Error("item author is empty")
	}
	if item.Time == 0 {
		t.Error("item time is 0")
	}
}

// TestFetchItems tests concurrent item fetching
func TestFetchItems(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Get some story IDs
	ids, err := client.FetchStoryIDs(FeedTop)
	if err != nil {
		t.Fatalf("FetchStoryIDs failed: %v", err)
	}

	// Fetch first 5 items
	batchSize := min(5, len(ids))
	items, err := client.FetchItems(ids[:batchSize])
	if err != nil {
		t.Fatalf("FetchItems failed: %v", err)
	}

	if len(items) != batchSize {
		t.Errorf("expected %d items, got %d", batchSize, len(items))
	}

	// Count non-nil items
	nonNil := 0
	for _, item := range items {
		if item != nil {
			nonNil++
		}
	}
	if nonNil == 0 {
		t.Error("all items are nil")
	}
}

// TestFetchCommentTree tests fetching comments
func TestFetchCommentTree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()

	// Get story IDs
	ids, err := client.FetchStoryIDs(FeedTop)
	if err != nil {
		t.Fatalf("FetchStoryIDs failed: %v", err)
	}

	// Find a story with comments
	var storyWithComments *Item
	items, _ := client.FetchItems(ids[:min(10, len(ids))])
	for _, item := range items {
		if item != nil && item.Descendants > 0 {
			storyWithComments = item
			break
		}
	}

	if storyWithComments == nil {
		t.Skip("no story with comments found")
	}

	// Fetch comment tree with max depth of 2
	comments, err := client.FetchCommentTree(storyWithComments, 2)
	if err != nil {
		t.Fatalf("FetchCommentTree failed: %v", err)
	}

	// Should have at least one comment
	if len(comments) == 0 {
		t.Log("story has descendant count but no comments returned (possibly deleted)")
	} else {
		// Verify comment structure
		for _, c := range comments {
			if c.Item == nil {
				t.Error("comment has nil Item")
				continue
			}
			if c.Depth != 0 {
				t.Errorf("expected top-level comment depth 0, got %d", c.Depth)
			}
		}
	}
}

// TestItemTimeAgo tests the TimeAgo method
func TestItemTimeAgo(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"just now", now, "just now"},
		{"30 seconds ago", now.Add(-30 * time.Second), "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute), "1 minute ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1 hour ago"},
		{"3 hours ago", now.Add(-3 * time.Hour), "3 hours ago"},
		{"1 day ago", now.Add(-24 * time.Hour), "1 day ago"},
		{"7 days ago", now.Add(-7 * 24 * time.Hour), "7 days ago"},
		{"1 month ago", now.Add(-30 * 24 * time.Hour), "1 month ago"},
		{"6 months ago", now.Add(-6 * 30 * 24 * time.Hour), "6 months ago"},
		{"1 year ago", now.Add(-365 * 24 * time.Hour), "1 year ago"},
		{"2 years ago", now.Add(-2 * 365 * 24 * time.Hour), "2 years ago"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := &Item{Time: tc.time.Unix()}
			result := item.TimeAgo()
			if result != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestItemDomain tests the Domain method
func TestItemDomain(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		{"empty URL", "", ""},
		{"https with www", "https://www.example.com/path", "example.com"},
		{"https without www", "https://example.com/path", "example.com"},
		{"http with www", "http://www.example.com/path", "example.com"},
		{"http without www", "http://example.com/path", "example.com"},
		{"subdomain", "https://blog.example.com/path", "blog.example.com"},
		{"no path", "https://example.com", "example.com"},
		{"with port", "https://example.com:8080/path", "example.com:8080"},
		{"complex path", "https://example.com/foo/bar?query=1", "example.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := &Item{URL: tc.url}
			result := item.Domain()
			if result != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestFeedConstants tests that feed constants are correct
func TestFeedConstants(t *testing.T) {
	if FeedTop != "topstories" {
		t.Errorf("FeedTop should be 'topstories', got '%s'", FeedTop)
	}
	if FeedNew != "newstories" {
		t.Errorf("FeedNew should be 'newstories', got '%s'", FeedNew)
	}
	if FeedBest != "beststories" {
		t.Errorf("FeedBest should be 'beststories', got '%s'", FeedBest)
	}
	if FeedAsk != "askstories" {
		t.Errorf("FeedAsk should be 'askstories', got '%s'", FeedAsk)
	}
	if FeedShow != "showstories" {
		t.Errorf("FeedShow should be 'showstories', got '%s'", FeedShow)
	}
}
