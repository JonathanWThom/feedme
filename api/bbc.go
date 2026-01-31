package api

// BBC News RSS feeds
// BBC provides excellent RSS coverage across different topics

var bbcFeeds = map[string]string{
	"top":       "https://feeds.bbci.co.uk/news/rss.xml",
	"world":     "https://feeds.bbci.co.uk/news/world/rss.xml",
	"uk":        "https://feeds.bbci.co.uk/news/uk/rss.xml",
	"business":  "https://feeds.bbci.co.uk/news/business/rss.xml",
	"tech":      "https://feeds.bbci.co.uk/news/technology/rss.xml",
	"science":   "https://feeds.bbci.co.uk/news/science_and_environment/rss.xml",
	"health":    "https://feeds.bbci.co.uk/news/health/rss.xml",
	"entertain": "https://feeds.bbci.co.uk/news/entertainment_and_arts/rss.xml",
}

var bbcFeedNames = []string{"top", "world", "uk", "business", "tech", "science", "health", "entertain"}
var bbcFeedLabels = []string{"Top", "World", "UK", "Business", "Tech", "Science", "Health", "Arts"}

// BBCClient is a client for BBC News RSS feeds
type BBCClient struct {
	*RSSClient
}

// NewBBCClient creates a new BBC News client
func NewBBCClient() *BBCClient {
	return &BBCClient{
		RSSClient: NewRSSClient("BBC", "https://www.bbc.com", bbcFeeds, bbcFeedNames, bbcFeedLabels),
	}
}

// StoryURL returns the BBC article URL
func (c *BBCClient) StoryURL(item *Item) string {
	return item.URL
}
