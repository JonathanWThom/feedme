package api

import (
	"regexp"
	"strconv"
	"strings"

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
	item := &Item{Type: "story"}

	parseLobstersTitle(s, item)
	parseLobstersShortID(s, item)
	parseLobstersScore(s, item)
	parseLobstersAuthor(s, item)
	parseLobstersTime(s.Find(".byline time"), item)
	parseLobstersCommentCount(s, item)
	parseLobstersTags(s, item)

	if item.Title == "" {
		return nil
	}
	return item
}

func parseLobstersTitle(s *goquery.Selection, item *Item) {
	linkSel := s.Find("a.u-url")
	if linkSel.Length() == 0 {
		linkSel = s.Find(".link a").First()
	}
	if linkSel.Length() == 0 {
		return
	}
	item.Title = strings.TrimSpace(linkSel.Text())
	item.URL, _ = linkSel.Attr("href")
	if strings.HasPrefix(item.URL, "/") {
		item.URL = lobstersBaseURL + item.URL
	}
}

func parseLobstersShortID(s *goquery.Selection, item *Item) {
	if shortID, exists := s.Attr("data-shortid"); exists {
		item.Type = shortID
		item.ID = hashShortID(shortID)
	}
}

func parseLobstersScore(s *goquery.Selection, item *Item) {
	scoreSel := s.Find(".voters a.upvoter")
	if scoreSel.Length() == 0 {
		return
	}
	scoreText := strings.TrimSpace(scoreSel.Text())
	if score, err := strconv.Atoi(scoreText); err == nil {
		item.Score = score
	}
}

func parseLobstersAuthor(s *goquery.Selection, item *Item) {
	authorSel := s.Find(".byline a.u-author")
	if authorSel.Length() == 0 {
		authorSel = s.Find(".byline a[href^='/~']")
	}
	if authorSel.Length() > 0 {
		item.By = strings.TrimSpace(authorSel.Text())
	}
}

func parseLobstersCommentCount(s *goquery.Selection, item *Item) {
	commentSel := s.Find(".comments_label")
	if commentSel.Length() == 0 {
		return
	}
	commentText := strings.TrimSpace(commentSel.Text())
	re := regexp.MustCompile(`(\d+)`)
	if matches := re.FindStringSubmatch(commentText); len(matches) > 1 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			item.Descendants = count
		}
	}
}

func parseLobstersTags(s *goquery.Selection, item *Item) {
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
	item := &Item{Type: "comment"}

	parseLobstersAuthor(s, item)
	parseLobstersCommentText(s, item)
	parseLobstersTime(s.Find(".byline time"), item)
	parseLobstersScore(s, item)

	if item.By == "" && item.Text == "" {
		return nil
	}
	return &Comment{
		Item:  item,
		Depth: parseLobstersDepth(s),
	}
}

func parseLobstersCommentText(s *goquery.Selection, item *Item) {
	textSel := s.Find(".comment_text")
	if textSel.Length() > 0 {
		html, _ := textSel.Html()
		item.Text = html
	}
}

func parseLobstersDepth(s *goquery.Selection) int {
	depth := 0
	s.Parents().Each(func(i int, parent *goquery.Selection) {
		if parent.Is("ol.comments") {
			depth++
		}
	})
	if depth > 0 {
		depth--
	}
	return depth
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
