package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// parseLobstersStories extracts stories from the HTML document
func parseLobstersStories(doc *goquery.Document) ([]*Item, error) {
	var stories []*Item

	doc.Find("ol.stories > li.story").Each(func(i int, s *goquery.Selection) {
		story := parseLobstersStory(s)
		if story != nil {
			stories = append(stories, story)
		}
	})

	return stories, nil
}

// parseLobstersStory extracts a single story from an HTML element
func parseLobstersStory(s *goquery.Selection) *Item {
	item := &Item{
		Type: "story",
	}

	// Get story link and title
	linkSel := s.Find("a.u-url")
	if linkSel.Length() == 0 {
		linkSel = s.Find(".link a").First()
	}
	if linkSel.Length() > 0 {
		item.Title = strings.TrimSpace(linkSel.Text())
		item.URL, _ = linkSel.Attr("href")
		if strings.HasPrefix(item.URL, "/") {
			item.URL = lobstersBaseURL + item.URL
		}
	}

	// Get short ID from the data attribute
	if shortID, exists := s.Attr("data-shortid"); exists {
		item.Type = shortID
		item.ID = hashShortID(shortID)
	}

	// Get score
	scoreSel := s.Find(".voters a.upvoter")
	if scoreSel.Length() > 0 {
		scoreText := strings.TrimSpace(scoreSel.Text())
		if score, err := strconv.Atoi(scoreText); err == nil {
			item.Score = score
		}
	}

	// Get author
	authorSel := s.Find(".byline a.u-author")
	if authorSel.Length() == 0 {
		authorSel = s.Find(".byline a[href^='/~']")
	}
	if authorSel.Length() > 0 {
		item.By = strings.TrimSpace(authorSel.Text())
	}

	// Get time
	parseLobstersTime(s.Find(".byline time"), item)

	// Get comment count
	commentSel := s.Find(".comments_label")
	if commentSel.Length() > 0 {
		commentText := strings.TrimSpace(commentSel.Text())
		re := regexp.MustCompile(`(\d+)`)
		if matches := re.FindStringSubmatch(commentText); len(matches) > 1 {
			if count, err := strconv.Atoi(matches[1]); err == nil {
				item.Descendants = count
			}
		}
	}

	// Get tags
	var tags []string
	s.Find(".tags a.tag").Each(func(i int, tagSel *goquery.Selection) {
		tag := strings.TrimSpace(tagSel.Text())
		if tag != "" {
			tags = append(tags, tag)
		}
	})
	if len(tags) > 0 {
		item.Text = "[" + strings.Join(tags, ", ") + "]"
	}

	if item.Title == "" {
		return nil
	}

	return item
}

// parseLobstersComments extracts comments from a story page
func parseLobstersComments(doc *goquery.Document) ([]*Comment, error) {
	var comments []*Comment

	doc.Find("div.comment[data-shortid]").Each(func(i int, s *goquery.Selection) {
		comment := parseLobstersComment(s)
		if comment != nil {
			comments = append(comments, comment)
		}
	})

	return comments, nil
}

// parseLobstersComment extracts a single comment
func parseLobstersComment(s *goquery.Selection) *Comment {
	item := &Item{
		Type: "comment",
	}

	// Get author
	authorSel := s.Find(".byline a[href^='/~']")
	if authorSel.Length() > 0 {
		item.By = strings.TrimSpace(authorSel.Text())
	}

	// Get comment text
	textSel := s.Find(".comment_text")
	if textSel.Length() > 0 {
		html, _ := textSel.Html()
		item.Text = html
	}

	// Get time
	parseLobstersTime(s.Find(".byline time"), item)

	// Get depth by counting parent ol.comments elements
	depth := 0
	s.Parents().Each(func(i int, parent *goquery.Selection) {
		if parent.Is("ol.comments") {
			depth++
		}
	})
	if depth > 0 {
		depth--
	}

	// Get score
	scoreSel := s.Find(".voters a.upvoter")
	if scoreSel.Length() > 0 {
		scoreText := strings.TrimSpace(scoreSel.Text())
		if score, err := strconv.Atoi(scoreText); err == nil {
			item.Score = score
		}
	}

	if item.By == "" && item.Text == "" {
		return nil
	}

	return &Comment{
		Item:  item,
		Depth: depth,
	}
}

// parseLobstersTime extracts time from a Lobsters time element into an Item
func parseLobstersTime(timeSel *goquery.Selection, item *Item) {
	if timeSel.Length() == 0 {
		return
	}
	if title, exists := timeSel.Attr("title"); exists {
		if t, err := parseTime(title); err == nil {
			item.Time = t.Unix()
		}
	}
	if item.Time == 0 {
		timeText := strings.TrimSpace(timeSel.Text())
		item.Time = parseRelativeTime(timeText)
	}
}

// hashShortID creates a pseudo-numeric ID from a short ID string
func hashShortID(shortID string) int {
	hash := 0
	for i, c := range shortID {
		hash += int(c) * (i + 1) * 31
	}
	return hash
}

// parseTime attempts to parse a timestamp string
func parseTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05 -0700",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// parseRelativeTime converts relative time strings to Unix timestamp
func parseRelativeTime(s string) int64 {
	now := time.Now()
	s = strings.ToLower(strings.TrimSpace(s))

	re := regexp.MustCompile(`(\d+)\s*(second|minute|hour|day|week|month|year)s?\s*ago`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 3 {
		return now.Unix()
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return now.Unix()
	}

	var duration time.Duration
	switch matches[2] {
	case "second":
		duration = time.Duration(num) * time.Second
	case "minute":
		duration = time.Duration(num) * time.Minute
	case "hour":
		duration = time.Duration(num) * time.Hour
	case "day":
		duration = time.Duration(num) * 24 * time.Hour
	case "week":
		duration = time.Duration(num) * 7 * 24 * time.Hour
	case "month":
		duration = time.Duration(num) * 30 * 24 * time.Hour
	case "year":
		duration = time.Duration(num) * 365 * 24 * time.Hour
	}

	return now.Add(-duration).Unix()
}
