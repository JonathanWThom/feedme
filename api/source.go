package api

// Source represents a news source (HN, Lobste.rs, etc.)
type Source interface {
	// Name returns the display name of the source
	Name() string

	// FeedNames returns the available feed names for this source
	FeedNames() []string

	// FeedLabels returns the display labels for feeds (for UI tabs)
	FeedLabels() []string

	// FetchStoryIDs fetches the list of story IDs for a given feed
	// For sources without IDs (like Lobste.rs), this returns page-based pseudo-IDs
	FetchStoryIDs(feed string) ([]int, error)

	// FetchItem fetches a single item by ID
	FetchItem(id int) (*Item, error)

	// FetchItems fetches multiple items by ID
	FetchItems(ids []int) ([]*Item, error)

	// FetchCommentTree fetches the comment tree for a story
	FetchCommentTree(item *Item, maxDepth int) ([]*Comment, error)

	// StoryURL returns the URL for viewing a story on the source's website
	StoryURL(item *Item) string
}
