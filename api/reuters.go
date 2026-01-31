package api

// Reuters RSS feeds
// Reuters wire news - straight news, less opinion

var reutersFeeds = map[string]string{
	"world":    "https://www.reutersagency.com/feed/?best-topics=world&post_type=best",
	"business": "https://www.reutersagency.com/feed/?best-topics=business-finance&post_type=best",
	"tech":     "https://www.reutersagency.com/feed/?best-topics=tech&post_type=best",
	"sports":   "https://www.reutersagency.com/feed/?best-topics=sports&post_type=best",
	"life":     "https://www.reutersagency.com/feed/?best-topics=lifestyle&post_type=best",
}

var reutersFeedNames = []string{"world", "business", "tech", "sports", "life"}
var reutersFeedLabels = []string{"World", "Business", "Tech", "Sports", "Life"}

// ReutersClient is a client for Reuters RSS feeds
type ReutersClient struct {
	*RSSClient
}

// NewReutersClient creates a new Reuters client
func NewReutersClient() *ReutersClient {
	return &ReutersClient{
		RSSClient: NewRSSClient("Reuters", "https://www.reuters.com", reutersFeeds, reutersFeedNames, reutersFeedLabels),
	}
}

// StoryURL returns the Reuters article URL
func (c *ReutersClient) StoryURL(item *Item) string {
	return item.URL
}
