package api

import (
	"testing"
	"time"
)

// TestRedditClientInterface verifies that RedditClient implements Source interface
func TestRedditClientInterface(t *testing.T) {
	var _ Source = (*RedditClient)(nil)
}

func TestNewRedditClient(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedSubred string
	}{
		{"plain name", "golang", "golang"},
		{"with r/ prefix", "r/golang", "golang"},
		{"with /r/ prefix", "/r/golang", "golang"},
		{"uppercase", "GOLANG", "GOLANG"},
		{"mixed case", "GoLang", "GoLang"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewRedditClient(tc.input)
			if client == nil {
				t.Fatal("NewRedditClient returned nil")
			}
			if client.subreddit != tc.expectedSubred {
				t.Errorf("expected subreddit '%s', got '%s'", tc.expectedSubred, client.subreddit)
			}
			if client.http == nil {
				t.Error("HTTP client is nil")
			}
			if client.storyCache == nil {
				t.Error("Story cache is nil")
			}
			if client.idToReddit == nil {
				t.Error("ID map is nil")
			}
		})
	}
}

func TestRedditClientName(t *testing.T) {
	client := NewRedditClient("golang")
	name := client.Name()
	if name != "r/golang" {
		t.Errorf("expected name 'r/golang', got '%s'", name)
	}
}

func TestRedditClientFeedNames(t *testing.T) {
	client := NewRedditClient("golang")
	feedNames := client.FeedNames()

	expected := []string{RedditFeedHot, RedditFeedNew, RedditFeedTop, RedditFeedRising, RedditFeedBest}
	if len(feedNames) != len(expected) {
		t.Errorf("expected %d feed names, got %d", len(expected), len(feedNames))
	}

	for i, name := range expected {
		if feedNames[i] != name {
			t.Errorf("expected feed name '%s' at index %d, got '%s'", name, i, feedNames[i])
		}
	}
}

func TestRedditClientFeedLabels(t *testing.T) {
	client := NewRedditClient("golang")
	labels := client.FeedLabels()

	expected := []string{"Hot", "New", "Top", "Rising", "Best"}
	if len(labels) != len(expected) {
		t.Errorf("expected %d labels, got %d", len(expected), len(labels))
	}

	for i, label := range expected {
		if labels[i] != label {
			t.Errorf("expected label '%s' at index %d, got '%s'", label, i, labels[i])
		}
	}
}

func TestRedditClientStoryURL(t *testing.T) {
	client := NewRedditClient("golang")

	testCases := []struct {
		name     string
		item     *Item
		expected string
	}{
		{
			name:     "with permalink",
			item:     &Item{Type: "/r/golang/comments/abc123/test_post", URL: "https://example.com"},
			expected: "https://www.reddit.com/r/golang/comments/abc123/test_post",
		},
		{
			name:     "without permalink",
			item:     &Item{Type: "not_a_permalink", URL: "https://example.com"},
			expected: "https://example.com",
		},
		{
			name:     "empty type",
			item:     &Item{Type: "", URL: "https://example.com"},
			expected: "https://example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := client.StoryURL(tc.item)
			if url != tc.expected {
				t.Errorf("expected URL '%s', got '%s'", tc.expected, url)
			}
		})
	}
}

// TestRedditFetchStoryIDs tests fetching story IDs from real Reddit API
func TestRedditFetchStoryIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewRedditClient("golang")

	testCases := []string{RedditFeedHot, RedditFeedNew, RedditFeedTop}
	for _, feed := range testCases {
		t.Run(feed, func(t *testing.T) {
			ids, err := client.FetchStoryIDs(feed)
			if err != nil {
				t.Fatalf("FetchStoryIDs failed for %s: %v", feed, err)
			}
			if len(ids) == 0 {
				t.Errorf("expected non-empty story IDs for feed %s", feed)
			}
			// Should get up to 100 stories
			if len(ids) > 100 {
				t.Errorf("unexpected number of IDs: %d (expected max 100)", len(ids))
			}
			// Verify IDs are positive (pseudo-IDs start at 1)
			for _, id := range ids {
				if id <= 0 {
					t.Errorf("invalid story ID: %d", id)
				}
			}
		})
	}
}

// TestRedditFetchStoryIDsMultipleSubreddits tests various subreddits
func TestRedditFetchStoryIDsMultipleSubreddits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	subreddits := []string{"programming", "technology", "news"}

	for _, sub := range subreddits {
		t.Run(sub, func(t *testing.T) {
			client := NewRedditClient(sub)
			ids, err := client.FetchStoryIDs(RedditFeedHot)
			if err != nil {
				t.Fatalf("FetchStoryIDs failed for r/%s: %v", sub, err)
			}
			if len(ids) == 0 {
				t.Errorf("expected non-empty story IDs for r/%s", sub)
			}
			t.Logf("r/%s: got %d stories", sub, len(ids))
		})
	}
}

// TestRedditFetchItem tests fetching cached items
func TestRedditFetchItem(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewRedditClient("golang")

	// First fetch story IDs to populate cache
	ids, err := client.FetchStoryIDs(RedditFeedHot)
	if err != nil {
		t.Fatalf("FetchStoryIDs failed: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("no story IDs returned")
	}

	// Fetch the first item from cache
	item, err := client.FetchItem(ids[0])
	if err != nil {
		t.Fatalf("FetchItem failed: %v", err)
	}
	if item == nil {
		t.Fatal("FetchItem returned nil")
	}

	// Validate item fields
	if item.Title == "" {
		t.Error("item title is empty")
	}
	if item.By == "" {
		t.Error("item author is empty")
	}
	// URL might be empty for self posts, but Type (permalink) should be set
	if item.Type == "" {
		t.Error("item Type (permalink) is empty")
	}
	if item.Time == 0 {
		t.Error("item time is 0")
	}
}

// TestRedditFetchItemNotInCache tests fetching non-existent item
func TestRedditFetchItemNotInCache(t *testing.T) {
	client := NewRedditClient("golang")

	// Try to fetch an item that doesn't exist in cache
	_, err := client.FetchItem(99999)
	if err == nil {
		t.Error("expected error for non-existent item, got nil")
	}
}

// TestRedditFetchItems tests fetching multiple items
func TestRedditFetchItems(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewRedditClient("golang")

	// Fetch story IDs first
	ids, err := client.FetchStoryIDs(RedditFeedHot)
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

// TestRedditFetchCommentTree tests fetching comments
func TestRedditFetchCommentTree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewRedditClient("golang")

	// Fetch story IDs
	ids, err := client.FetchStoryIDs(RedditFeedHot)
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

	// Fetch comment tree
	comments, err := client.FetchCommentTree(storyWithComments, 0)
	if err != nil {
		t.Fatalf("FetchCommentTree failed: %v", err)
	}

	t.Logf("Found %d comments for story '%s'", len(comments), storyWithComments.Title)

	// Verify comment structure if we have any
	for _, c := range comments {
		if c.Item == nil {
			t.Error("comment has nil Item")
			continue
		}
		if c.By == "" {
			t.Error("comment has empty author")
		}
		if c.Depth < 0 {
			t.Errorf("invalid comment depth: %d", c.Depth)
		}
	}
}

// TestRedditFetchCommentTreeNoPermalink tests error handling
func TestRedditFetchCommentTreeNoPermalink(t *testing.T) {
	client := NewRedditClient("golang")

	// Item without permalink
	item := &Item{Type: "not_a_permalink"}
	_, err := client.FetchCommentTree(item, 0)
	if err == nil {
		t.Error("expected error for item without valid permalink")
	}
}

// TestRedditFeedConstants tests that feed constants are correct
func TestRedditFeedConstants(t *testing.T) {
	if RedditFeedHot != "hot" {
		t.Errorf("RedditFeedHot should be 'hot', got '%s'", RedditFeedHot)
	}
	if RedditFeedNew != "new" {
		t.Errorf("RedditFeedNew should be 'new', got '%s'", RedditFeedNew)
	}
	if RedditFeedTop != "top" {
		t.Errorf("RedditFeedTop should be 'top', got '%s'", RedditFeedTop)
	}
	if RedditFeedRising != "rising" {
		t.Errorf("RedditFeedRising should be 'rising', got '%s'", RedditFeedRising)
	}
	if RedditFeedBest != "best" {
		t.Errorf("RedditFeedBest should be 'best', got '%s'", RedditFeedBest)
	}
}

// TestRedditThrottle tests that throttling doesn't cause issues
func TestRedditThrottle(t *testing.T) {
	client := NewRedditClient("golang")

	// Record initial time
	start := time.Now()

	// Call throttle twice
	client.throttle()
	client.throttle()

	// Second call should have waited ~1 second
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Errorf("throttle didn't wait long enough: %v", elapsed)
	}
}

// TestRedditCacheClear tests that cache is cleared on new fetch
func TestRedditCacheClear(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewRedditClient("golang")

	// Fetch hot stories
	ids1, err := client.FetchStoryIDs(RedditFeedHot)
	if err != nil {
		t.Fatalf("First FetchStoryIDs failed: %v", err)
	}

	// Note the cache size
	client.cacheMu.RLock()
	cacheSize1 := len(client.storyCache)
	client.cacheMu.RUnlock()

	if cacheSize1 == 0 {
		t.Fatal("Cache should not be empty after fetch")
	}

	// Fetch new stories (different feed)
	_, err = client.FetchStoryIDs(RedditFeedNew)
	if err != nil {
		t.Fatalf("Second FetchStoryIDs failed: %v", err)
	}

	// Cache should still have entries (cleared and repopulated)
	client.cacheMu.RLock()
	cacheSize2 := len(client.storyCache)
	client.cacheMu.RUnlock()

	if cacheSize2 == 0 {
		t.Fatal("Cache should not be empty after second fetch")
	}

	// Old IDs should no longer be valid
	_, err = client.FetchItem(ids1[0])
	if err == nil {
		t.Error("Old cached item should not be accessible after cache clear")
	}
}
