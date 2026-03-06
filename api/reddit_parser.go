package api

import (
	"encoding/json"
	"fmt"
)

// redditListing represents the top-level JSON structure
type redditListing struct {
	Data struct {
		Children []struct {
			Data redditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// redditPost represents a Reddit post
type redditPost struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	Score         int     `json:"score"`
	URL           string  `json:"url"`
	Permalink     string  `json:"permalink"`
	NumComments   int     `json:"num_comments"`
	CreatedUTC    float64 `json:"created_utc"`
	Selftext      string  `json:"selftext"`
	IsSelf        bool    `json:"is_self"`
	Subreddit     string  `json:"subreddit"`
	Domain        string  `json:"domain"`
	LinkFlairText string  `json:"link_flair_text"`
}

// redditCommentListing represents the comments JSON structure
type redditCommentListing struct {
	Data struct {
		Children []struct {
			Kind string          `json:"kind"`
			Data json.RawMessage `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// redditComment represents a Reddit comment
type redditComment struct {
	ID         string  `json:"id"`
	Author     string  `json:"author"`
	Body       string  `json:"body"`
	Score      int     `json:"score"`
	CreatedUTC float64 `json:"created_utc"`
	Depth      int     `json:"depth"`
	Replies    any     `json:"replies"` // Can be "" or a listing
}

// parseRedditStories converts a listing of Reddit posts to Items
func parseRedditStories(listing redditListing) []*Item {
	var stories []*Item
	for _, child := range listing.Data.Children {
		stories = append(stories, redditPostToItem(child.Data))
	}
	return stories
}

func redditPostToItem(post redditPost) *Item {
	item := &Item{
		ID:          hashShortID(post.ID),
		Type:        post.Permalink,
		Title:       post.Title,
		By:          post.Author,
		Score:       post.Score,
		URL:         post.URL,
		Time:        int64(post.CreatedUTC),
		Descendants: post.NumComments,
	}
	if post.LinkFlairText != "" {
		item.Text = "[" + post.LinkFlairText + "]"
	}
	if post.IsSelf {
		item.URL = fmt.Sprintf("https://www.reddit.com%s", post.Permalink)
	}
	return item
}

// parseRedditComments extracts comments from the listing
func parseRedditComments(listing redditCommentListing, maxDepth int) ([]*Comment, error) {
	var comments []*Comment

	for _, child := range listing.Data.Children {
		if child.Kind != "t1" {
			continue
		}

		var rc redditComment
		if err := json.Unmarshal(child.Data, &rc); err != nil {
			continue
		}

		comment := parseRedditComment(rc, maxDepth)
		if comment != nil {
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

// parseRedditComment converts a Reddit comment to our Comment type
func parseRedditComment(rc redditComment, maxDepth int) *Comment {
	if rc.Author == "" || rc.Author == "[deleted]" {
		return nil
	}

	item := &Item{
		ID:    hashShortID(rc.ID),
		Type:  "comment",
		By:    rc.Author,
		Text:  rc.Body,
		Score: rc.Score,
		Time:  int64(rc.CreatedUTC),
	}

	comment := &Comment{
		Item:  item,
		Depth: rc.Depth,
	}

	if maxDepth <= 0 || rc.Depth < maxDepth {
		comment.Children = parseRedditReplies(rc.Replies, maxDepth)
	}

	return comment
}

// parseRedditReplies extracts child comments from the replies field
func parseRedditReplies(replies any, maxDepth int) []*Comment {
	repliesMap, ok := replies.(map[string]any)
	if !ok {
		return nil
	}
	data, ok := repliesMap["data"].(map[string]any)
	if !ok {
		return nil
	}
	children, ok := data["children"].([]any)
	if !ok {
		return nil
	}

	var result []*Comment
	for _, child := range children {
		childMap, ok := child.(map[string]any)
		if !ok || childMap["kind"] != "t1" {
			continue
		}
		childData, ok := childMap["data"].(map[string]any)
		if !ok {
			continue
		}
		childJSON, _ := json.Marshal(childData)
		var childRC redditComment
		if err := json.Unmarshal(childJSON, &childRC); err != nil {
			continue
		}
		if c := parseRedditComment(childRC, maxDepth); c != nil {
			result = append(result, c)
		}
	}
	return result
}
