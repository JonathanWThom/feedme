package main

import (
	"testing"
	"time"

	"github.com/JonathanWThom/feedme/api"
)

// Integration tests that hit real APIs
// Run with: go test -v -run Integration
// Skip with: go test -short

// TestIntegrationHackerNewsFullFlow tests the complete HN flow
func TestIntegrationHackerNewsFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := api.NewClient()

	// 1. Verify source info
	if client.Name() != "HN" {
		t.Errorf("expected name 'HN', got '%s'", client.Name())
	}

	// 2. Fetch story IDs from all feeds
	for _, feed := range client.FeedNames() {
		t.Run("Feed_"+feed, func(t *testing.T) {
			ids, err := client.FetchStoryIDs(feed)
			if err != nil {
				t.Fatalf("FetchStoryIDs failed for %s: %v", feed, err)
			}
			if len(ids) == 0 {
				t.Errorf("no stories found for feed %s", feed)
			}
			t.Logf("%s: %d story IDs", feed, len(ids))
		})
	}

	// 3. Fetch stories and verify structure
	ids, _ := client.FetchStoryIDs(api.FeedTop)
	stories, err := client.FetchItems(ids[:min(5, len(ids))])
	if err != nil {
		t.Fatalf("FetchItems failed: %v", err)
	}

	for i, story := range stories {
		if story == nil {
			continue
		}
		t.Logf("Story %d: %s (by %s, %d points)", i+1, story.Title, story.By, story.Score)

		// Verify basic fields
		if story.ID == 0 {
			t.Error("story ID should not be 0")
		}
		if story.Title == "" {
			t.Error("story title should not be empty")
		}
		if story.Time == 0 {
			t.Error("story time should not be 0")
		}

		// Test TimeAgo formatting
		timeAgo := story.TimeAgo()
		if timeAgo == "" {
			t.Error("TimeAgo should not be empty")
		}

		// Test Domain extraction
		if story.URL != "" {
			domain := story.Domain()
			if domain == "" {
				t.Error("Domain should not be empty for story with URL")
			}
		}

		// Test StoryURL generation
		storyURL := client.StoryURL(story)
		if storyURL == "" {
			t.Error("StoryURL should not be empty")
		}
	}

	// 4. Fetch comments for a story with comments
	var storyWithComments *api.Item
	for _, story := range stories {
		if story != nil && story.Descendants > 0 {
			storyWithComments = story
			break
		}
	}

	if storyWithComments != nil {
		comments, err := client.FetchCommentTree(storyWithComments, 2)
		if err != nil {
			t.Logf("FetchCommentTree failed: %v (may be rate limited)", err)
		} else {
			t.Logf("Story '%s' has %d top-level comments (expected ~%d total)",
				storyWithComments.Title, len(comments), storyWithComments.Descendants)

			// Verify comment structure
			for _, comment := range comments[:min(3, len(comments))] {
				if comment.Item == nil {
					t.Error("comment Item should not be nil")
					continue
				}
				if comment.By == "" {
					t.Error("comment author should not be empty")
				}
				t.Logf("  Comment by %s (depth %d): %d children",
					comment.By, comment.Depth, len(comment.Children))
			}
		}
	}
}

// TestIntegrationLobstersFullFlow tests the complete Lobsters flow
func TestIntegrationLobstersFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := api.NewLobstersClient()

	// 1. Verify source info
	if client.Name() != "Lobsters" {
		t.Errorf("expected name 'Lobsters', got '%s'", client.Name())
	}

	// 2. Fetch story IDs from all feeds (with delays to be polite)
	for i, feed := range client.FeedNames() {
		if i > 0 {
			time.Sleep(1 * time.Second) // Extra delay between feeds
		}
		t.Run("Feed_"+feed, func(t *testing.T) {
			ids, err := client.FetchStoryIDs(feed)
			if err != nil {
				t.Fatalf("FetchStoryIDs failed for %s: %v", feed, err)
			}
			if len(ids) == 0 {
				t.Errorf("no stories found for feed %s", feed)
			}
			t.Logf("%s: %d story IDs", feed, len(ids))
		})
	}

	// 3. Fetch stories and verify structure
	ids, _ := client.FetchStoryIDs(api.LobstersFeedHottest)
	stories, err := client.FetchItems(ids[:min(5, len(ids))])
	if err != nil {
		t.Fatalf("FetchItems failed: %v", err)
	}

	for i, story := range stories {
		if story == nil {
			continue
		}
		t.Logf("Story %d: %s (by %s, %d points)", i+1, story.Title, story.By, story.Score)

		// Verify basic fields
		if story.Title == "" {
			t.Error("story title should not be empty")
		}
		if story.By == "" {
			t.Error("story author should not be empty")
		}

		// Test StoryURL generation
		storyURL := client.StoryURL(story)
		if storyURL == "" {
			t.Error("StoryURL should not be empty")
		}

		// Check for tags (stored in Text field)
		if story.Text != "" {
			t.Logf("  Tags: %s", story.Text)
		}
	}

	// 4. Fetch comments for a story with comments
	var storyWithComments *api.Item
	for _, story := range stories {
		if story != nil && story.Descendants > 0 && story.Type != "" && story.Type != "story" {
			storyWithComments = story
			break
		}
	}

	if storyWithComments != nil {
		time.Sleep(1 * time.Second) // Be polite
		comments, err := client.FetchCommentTree(storyWithComments, 0)
		if err != nil {
			t.Logf("FetchCommentTree failed: %v (may be rate limited)", err)
		} else {
			t.Logf("Story '%s' has %d comments", storyWithComments.Title, len(comments))

			for _, comment := range comments[:min(3, len(comments))] {
				if comment.Item == nil {
					t.Error("comment Item should not be nil")
					continue
				}
				t.Logf("  Comment by %s (depth %d)", comment.By, comment.Depth)
			}
		}
	}
}

// TestIntegrationRedditFullFlow tests the complete Reddit flow
func TestIntegrationRedditFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test with r/golang
	client := api.NewRedditClient("golang")

	// 1. Verify source info
	expectedName := "r/golang"
	if client.Name() != expectedName {
		t.Errorf("expected name '%s', got '%s'", expectedName, client.Name())
	}

	// 2. Fetch story IDs from multiple feeds
	feeds := []string{api.RedditFeedHot, api.RedditFeedNew, api.RedditFeedTop}
	for i, feed := range feeds {
		if i > 0 {
			time.Sleep(2 * time.Second) // Reddit rate limit
		}
		t.Run("Feed_"+feed, func(t *testing.T) {
			ids, err := client.FetchStoryIDs(feed)
			if err != nil {
				t.Fatalf("FetchStoryIDs failed for %s: %v", feed, err)
			}
			if len(ids) == 0 {
				t.Errorf("no stories found for feed %s", feed)
			}
			t.Logf("r/golang/%s: %d posts", feed, len(ids))
		})
	}

	// 3. Fetch stories and verify structure
	ids, _ := client.FetchStoryIDs(api.RedditFeedHot)
	stories, err := client.FetchItems(ids[:min(5, len(ids))])
	if err != nil {
		t.Fatalf("FetchItems failed: %v", err)
	}

	for i, story := range stories {
		if story == nil {
			continue
		}
		t.Logf("Post %d: %s (by %s, %d upvotes)", i+1, story.Title, story.By, story.Score)

		// Verify basic fields
		if story.Title == "" {
			t.Error("post title should not be empty")
		}
		if story.By == "" {
			t.Error("post author should not be empty")
		}
		if story.Type == "" {
			t.Error("post Type (permalink) should not be empty")
		}

		// Test StoryURL generation
		storyURL := client.StoryURL(story)
		if storyURL == "" {
			t.Error("StoryURL should not be empty")
		}

		// Check for flair
		if story.Text != "" {
			t.Logf("  Flair: %s", story.Text)
		}
	}

	// 4. Fetch comments for a post with comments
	var postWithComments *api.Item
	for _, story := range stories {
		if story != nil && story.Descendants > 0 {
			postWithComments = story
			break
		}
	}

	if postWithComments != nil {
		time.Sleep(2 * time.Second) // Reddit rate limit
		comments, err := client.FetchCommentTree(postWithComments, 3)
		if err != nil {
			t.Logf("FetchCommentTree failed: %v (may be rate limited)", err)
		} else {
			t.Logf("Post '%s' has %d top-level comments", postWithComments.Title, len(comments))

			for _, comment := range comments[:min(3, len(comments))] {
				if comment.Item == nil {
					t.Error("comment Item should not be nil")
					continue
				}
				t.Logf("  Comment by %s (depth %d, %d points): %d children",
					comment.By, comment.Depth, comment.Score, len(comment.Children))
			}
		}
	}
}

// TestIntegrationMultipleSubreddits tests Reddit with different subreddits
func TestIntegrationMultipleSubreddits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	subreddits := []string{"programming", "technology", "news", "worldnews"}

	for i, sub := range subreddits {
		if i > 0 {
			time.Sleep(2 * time.Second) // Reddit rate limit
		}
		t.Run("r/"+sub, func(t *testing.T) {
			client := api.NewRedditClient(sub)

			if client.Name() != "r/"+sub {
				t.Errorf("expected name 'r/%s', got '%s'", sub, client.Name())
			}

			ids, err := client.FetchStoryIDs(api.RedditFeedHot)
			if err != nil {
				t.Fatalf("FetchStoryIDs failed for r/%s: %v", sub, err)
			}

			t.Logf("r/%s: %d posts", sub, len(ids))

			if len(ids) == 0 {
				t.Errorf("expected posts for r/%s", sub)
			}
		})
	}
}

// TestIntegrationSourceInterface tests that all sources implement the interface correctly
func TestIntegrationSourceInterface(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	sources := []api.Source{
		api.NewClient(),
		api.NewLobstersClient(),
		api.NewRedditClient("golang"),
	}

	for _, source := range sources {
		t.Run(source.Name(), func(t *testing.T) {
			// Test Name
			name := source.Name()
			if name == "" {
				t.Error("Name() should not be empty")
			}

			// Test FeedNames
			feedNames := source.FeedNames()
			if len(feedNames) == 0 {
				t.Error("FeedNames() should not be empty")
			}

			// Test FeedLabels
			feedLabels := source.FeedLabels()
			if len(feedLabels) == 0 {
				t.Error("FeedLabels() should not be empty")
			}
			if len(feedLabels) != len(feedNames) {
				t.Error("FeedLabels and FeedNames should have same length")
			}

			// Test FetchStoryIDs
			ids, err := source.FetchStoryIDs(feedNames[0])
			if err != nil {
				t.Fatalf("FetchStoryIDs failed: %v", err)
			}
			if len(ids) == 0 {
				t.Error("FetchStoryIDs should return some IDs")
			}

			time.Sleep(1 * time.Second) // Rate limiting

			// Test FetchItem
			item, err := source.FetchItem(ids[0])
			if err != nil {
				t.Fatalf("FetchItem failed: %v", err)
			}
			if item == nil {
				t.Error("FetchItem should not return nil")
			}

			// Test FetchItems
			items, err := source.FetchItems(ids[:min(3, len(ids))])
			if err != nil {
				t.Fatalf("FetchItems failed: %v", err)
			}
			if len(items) == 0 {
				t.Error("FetchItems should return some items")
			}

			// Test StoryURL
			if item != nil {
				url := source.StoryURL(item)
				if url == "" {
					t.Error("StoryURL should not be empty")
				}
			}

			t.Logf("%s: %d feeds, %d stories fetched", name, len(feedNames), len(ids))
		})

		time.Sleep(2 * time.Second) // Rate limiting between sources
	}
}

// TestIntegrationVersionCheck tests the update check against GitHub
func TestIntegrationVersionCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This should hit GitHub API
	info := api.CheckForUpdate("v0.0.1")

	// If nil, might be rate limited or no releases
	if info == nil {
		t.Log("CheckForUpdate returned nil (might be rate limited or no releases)")
		return
	}

	t.Logf("Current: %s, Latest: %s", info.CurrentVersion, info.LatestVersion)

	if info.CurrentVersion != "v0.0.1" {
		t.Errorf("CurrentVersion should be v0.0.1, got %s", info.CurrentVersion)
	}

	if info.LatestVersion == "" {
		t.Error("LatestVersion should not be empty")
	}

	// Check HasUpdate logic
	if info.HasUpdate() {
		t.Logf("Update available: %s", info.FormatUpdateMessage())
		if info.UpdateURL == "" {
			t.Error("UpdateURL should not be empty when update available")
		}
	} else {
		t.Log("No update available (or same version)")
	}
}

// TestIntegrationConcurrentRequests tests that concurrent requests work
func TestIntegrationConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := api.NewClient()

	// Fetch story IDs
	ids, err := client.FetchStoryIDs(api.FeedTop)
	if err != nil {
		t.Fatalf("FetchStoryIDs failed: %v", err)
	}

	// FetchItems uses concurrent fetching internally
	start := time.Now()
	items, err := client.FetchItems(ids[:min(10, len(ids))])
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("FetchItems failed: %v", err)
	}

	nonNil := 0
	for _, item := range items {
		if item != nil {
			nonNil++
		}
	}

	t.Logf("Fetched %d/%d items in %v", nonNil, len(items), elapsed)

	// Should be reasonably fast due to concurrency
	if elapsed > 30*time.Second {
		t.Errorf("Fetching 10 items should not take more than 30s, took %v", elapsed)
	}
}

// BenchmarkHNFetchStoryIDs benchmarks HN story ID fetching
func BenchmarkHNFetchStoryIDs(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	client := api.NewClient()

	for i := 0; i < b.N; i++ {
		_, err := client.FetchStoryIDs(api.FeedTop)
		if err != nil {
			b.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond) // Rate limiting
	}
}

// BenchmarkHNFetchItem benchmarks HN item fetching
func BenchmarkHNFetchItem(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	client := api.NewClient()
	ids, _ := client.FetchStoryIDs(api.FeedTop)
	if len(ids) == 0 {
		b.Fatal("no story IDs")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.FetchItem(ids[i%len(ids)])
		if err != nil {
			b.Fatal(err)
		}
		time.Sleep(50 * time.Millisecond) // Rate limiting
	}
}
