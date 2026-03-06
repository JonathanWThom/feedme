package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/JonathanWThom/feedme/api"
)

// visibleStoryCount returns how many stories fit on screen
func (m Model) visibleStoryCount() int {
	availableLines := m.height - 3
	count := availableLines / 2
	if count < 1 {
		return 1
	}
	return count
}

func (m Model) renderHeader() string {
	title := HeaderStyle.Render(" " + m.source.Name() + " ")

	if m.view == CommentsView {
		return title
	}

	var tabs []string
	feedLabels := m.source.FeedLabels()
	for i, label := range feedLabels {
		if i == m.feed {
			tabs = append(tabs, ActiveTabStyle.Render(label))
		} else {
			tabs = append(tabs, TabStyle.Render(label))
		}
	}

	tabsStr := strings.Join(tabs, "")
	return title + " " + tabsStr
}

func (m Model) renderStories() string {
	if len(m.stories) == 0 {
		return "\n  No stories to display\n"
	}

	var b strings.Builder
	visibleCount := m.visibleStoryCount()
	start := m.offset
	end := min(start+visibleCount, len(m.stories))

	for i := start; i < end; i++ {
		story := m.stories[i]
		if story == nil {
			continue
		}

		isSelected := i == m.cursor
		b.WriteString(m.renderStory(i, story, isSelected))
	}

	return b.String()
}

func (m Model) renderStory(idx int, story *api.Item, selected bool) string {
	var b strings.Builder
	b.WriteString(m.renderStoryNumber(idx, selected))
	b.WriteString(m.renderStoryTitle(story, selected))
	b.WriteString(renderStoryDomain(story))
	b.WriteString("\n")
	b.WriteString(MetaStyle.Render(storyMeta(story)))
	b.WriteString("\n")
	return b.String()
}

func (m Model) renderStoryNumber(idx int, selected bool) string {
	num := fmt.Sprintf("%3d. ", idx+1)
	if selected {
		return ScoreStyle.Render(num)
	}
	return MetaStyle.Render(num)
}

func (m Model) renderStoryTitle(story *api.Item, selected bool) string {
	title := story.Title
	if len(title) > m.width-20 {
		title = title[:m.width-23] + "..."
	}
	if selected {
		return SelectedTitleStyle.Render(title)
	}
	return TitleStyle.Render(title)
}

func renderStoryDomain(story *api.Item) string {
	if domain := story.Domain(); domain != "" {
		return " " + URLStyle.Render(fmt.Sprintf("(%s)", domain))
	}
	return ""
}

func storyMeta(story *api.Item) string {
	meta := fmt.Sprintf("      %d points by %s %s | %d comments",
		story.Score, story.By, story.TimeAgo(), story.Descendants)
	if story.Text != "" && strings.HasPrefix(story.Text, "[") {
		meta += " " + story.Text
	}
	return meta
}

func (m Model) renderComments() string {
	if m.currentItem == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.renderCommentHeader())
	b.WriteString(MetaStyle.Render(fmt.Sprintf("─── %d comments ───", m.currentItem.Descendants)))
	b.WriteString("\n\n")
	for _, comment := range m.comments {
		b.WriteString(m.renderComment(comment))
	}
	return b.String()
}

func (m Model) renderCommentHeader() string {
	var b strings.Builder
	b.WriteString(SelectedTitleStyle.Render(m.currentItem.Title) + "\n")
	b.WriteString(renderStoryDomain(m.currentItem))
	if domain := m.currentItem.Domain(); domain != "" {
		b.WriteString("\n")
	}
	meta := fmt.Sprintf("%d points by %s %s",
		m.currentItem.Score, m.currentItem.By, m.currentItem.TimeAgo())
	b.WriteString(MetaStyle.Render(meta) + "\n\n")
	if m.currentItem.Text != "" {
		text := cleanHTML(m.currentItem.Text)
		b.WriteString(CommentTextStyle.Render(wrapText(text, m.width-4)) + "\n\n")
	}
	return b.String()
}

func (m Model) renderComment(c *api.Comment) string {
	var b strings.Builder

	indent := strings.Repeat("  ", c.Depth)
	prefix := IndentStyle(c.Depth).Render("│ ")

	author := CommentAuthorStyle.Render(c.By)
	time := CommentMetaStyle.Render(c.TimeAgo())
	b.WriteString(indent + prefix + author + " " + time + "\n")

	text := cleanHTML(c.Text)
	lines := wrapTextLines(text, m.width-len(indent)-4)
	for _, line := range lines {
		b.WriteString(indent + prefix + CommentTextStyle.Render(line) + "\n")
	}
	b.WriteString(indent + prefix + "\n")

	for _, child := range c.Children {
		b.WriteString(m.renderComment(child))
	}

	return b.String()
}

func (m Model) renderStatusBar() string {
	left, right := m.statusBarContent()
	gap := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return StatusBarStyle.Width(m.width).Render(left + strings.Repeat(" ", gap) + right)
}

func (m Model) statusBarContent() (string, string) {
	suffix := m.statusBarSuffix()
	switch m.view {
	case StoriesView:
		return fmt.Sprintf(" %d/%d stories%s", m.cursor+1, len(m.stories), suffix),
			"↑↓:nav  enter:open  c:comments  tab:feed  s:source  ?:help  q:quit "
	case CommentsView:
		return m.commentsStatusLeft(suffix),
			m.commentsStatusRight()
	case SourcePickerView:
		return " Select a source",
			"↑↓:nav  enter:select  esc:cancel  q:quit "
	}
	return "", ""
}

func (m Model) statusBarSuffix() string {
	var s string
	if !m.mouseEnabled {
		s += " [SELECT MODE - m to exit]"
	}
	if m.updateInfo != nil && m.updateInfo.HasUpdate() {
		s += " [" + m.updateInfo.FormatUpdateMessage() + "]"
	}
	return s
}

func (m Model) commentsStatusLeft(suffix string) string {
	if m.visualMode {
		return fmt.Sprintf(" -- VISUAL -- lines %d-%d%s", m.visualStart+1, m.visualEnd+1, suffix)
	}
	return fmt.Sprintf(" %d comments%s", len(m.comments), suffix)
}

func (m Model) commentsStatusRight() string {
	if m.visualMode {
		return "↑↓:select  y:yank  esc:cancel "
	}
	return "↑↓:scroll  v:visual  o:open link  b:back  ?:help  q:quit "
}

func (m Model) renderFullHelp() string {
	return "\n" + m.help.FullHelpView(m.keys.FullHelp()) + "\n\nPress any key to close help."
}
