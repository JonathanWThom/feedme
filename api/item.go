package api

import (
	"fmt"
	"time"
)

// Item represents a news item (story, comment, job, poll)
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

// Comment represents a comment with its depth for rendering
type Comment struct {
	*Item
	Depth    int
	Children []*Comment
}
