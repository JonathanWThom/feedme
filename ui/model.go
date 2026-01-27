package ui

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jonathanthom/hn/api"
	"github.com/pkg/browser"
)

// View represents the current view
type View int

const (
	StoriesView View = iota
	CommentsView
)

// Messages
type storiesLoadedMsg struct {
	stories []*api.Item
	err     error
}

type commentsLoadedMsg struct {
	comments []*api.Comment
	err      error
}

type storyIDsLoadedMsg struct {
	ids []int
	err error
}

// Model is the main application model
type Model struct {
	client   *api.Client
	keys     KeyMap
	help     help.Model
	spinner  spinner.Model
	viewport viewport.Model

	// State
	view        View
	feed        int
	storyIDs    []int
	stories     []*api.Item
	comments    []*api.Comment
	cursor      int
	offset      int
	loading     bool
	err         error
	showHelp    bool
	width       int
	height      int
	currentItem *api.Item
}

// New creates a new Model
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	h := help.New()
	h.Styles.ShortKey = HelpStyle
	h.Styles.ShortDesc = HelpStyle

	return Model{
		client:  api.NewClient(),
		keys:    DefaultKeyMap(),
		help:    h,
		spinner: s,
		view:    StoriesView,
		feed:    0,
		loading: true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadStoryIDs(),
	)
}

func (m Model) loadStoryIDs() tea.Cmd {
	feed := api.FeedNames[m.feed]
	return func() tea.Msg {
		ids, err := m.client.FetchStoryIDs(feed)
		return storyIDsLoadedMsg{ids: ids, err: err}
	}
}

func (m Model) loadStories(ids []int) tea.Cmd {
	return func() tea.Msg {
		stories, err := m.client.FetchItems(ids)
		return storiesLoadedMsg{stories: stories, err: err}
	}
}

func (m Model) loadComments(item *api.Item) tea.Cmd {
	return func() tea.Msg {
		comments, err := m.client.FetchCommentTree(item, 0)
		return commentsLoadedMsg{comments: comments, err: err}
	}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle help toggle first
		if key.Matches(msg, m.keys.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}

		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Up):
			if m.view == StoriesView && m.cursor > 0 {
				m.cursor--
				m.adjustOffset()
			} else if m.view == CommentsView {
				m.viewport.LineUp(1)
			}

		case key.Matches(msg, m.keys.Down):
			if m.view == StoriesView && m.cursor < len(m.stories)-1 {
				m.cursor++
				m.adjustOffset()
				// Load more stories if near the end
				if m.cursor >= len(m.stories)-5 && len(m.storyIDs) > len(m.stories) {
					nextBatch := m.storyIDs[len(m.stories):min(len(m.stories)+30, len(m.storyIDs))]
					if len(nextBatch) > 0 {
						m.loading = true
						return m, m.loadStories(nextBatch)
					}
				}
			} else if m.view == CommentsView {
				m.viewport.LineDown(1)
			}

		case key.Matches(msg, m.keys.PageDown):
			if m.view == StoriesView {
				m.cursor = min(m.cursor+10, len(m.stories)-1)
				m.adjustOffset()
			} else {
				m.viewport.HalfViewDown()
			}

		case key.Matches(msg, m.keys.PageUp):
			if m.view == StoriesView {
				m.cursor = max(m.cursor-10, 0)
				m.adjustOffset()
			} else {
				m.viewport.HalfViewUp()
			}

		case key.Matches(msg, m.keys.Home):
			if m.view == StoriesView {
				m.cursor = 0
				m.offset = 0
			} else {
				m.viewport.GotoTop()
			}

		case key.Matches(msg, m.keys.End):
			if m.view == StoriesView {
				m.cursor = len(m.stories) - 1
				m.adjustOffset()
			} else {
				m.viewport.GotoBottom()
			}

		case key.Matches(msg, m.keys.Enter), key.Matches(msg, m.keys.Open):
			if m.view == StoriesView && len(m.stories) > 0 {
				story := m.stories[m.cursor]
				if story != nil && story.URL != "" {
					_ = browser.OpenURL(story.URL)
				} else if story != nil {
					// For Ask HN, Show HN, etc. - open the HN page
					url := fmt.Sprintf("https://news.ycombinator.com/item?id=%d", story.ID)
					_ = browser.OpenURL(url)
				}
			}

		case key.Matches(msg, m.keys.Comments):
			if m.view == StoriesView && len(m.stories) > 0 {
				story := m.stories[m.cursor]
				if story != nil && story.Descendants > 0 {
					m.currentItem = story
					m.view = CommentsView
					m.loading = true
					m.comments = nil
					return m, tea.Batch(m.spinner.Tick, m.loadComments(story))
				}
			}

		case key.Matches(msg, m.keys.Back):
			if m.view == CommentsView {
				m.view = StoriesView
				m.comments = nil
			}

		case key.Matches(msg, m.keys.NextTab):
			if m.view == StoriesView {
				m.feed = (m.feed + 1) % len(api.FeedNames)
				m.stories = nil
				m.storyIDs = nil
				m.cursor = 0
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
			}

		case key.Matches(msg, m.keys.PrevTab):
			if m.view == StoriesView {
				m.feed = (m.feed - 1 + len(api.FeedNames)) % len(api.FeedNames)
				m.stories = nil
				m.storyIDs = nil
				m.cursor = 0
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
			}

		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			m.stories = nil
			m.storyIDs = nil
			m.cursor = 0
			return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport = viewport.New(msg.Width, msg.Height-4)
		m.viewport.Style = lipgloss.NewStyle()
		m.help.Width = msg.Width

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case storyIDsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.loading = false
		} else {
			m.storyIDs = msg.ids
			// Load first batch of stories
			batchSize := min(30, len(msg.ids))
			return m, m.loadStories(msg.ids[:batchSize])
		}

	case storiesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			// Append to existing stories
			m.stories = append(m.stories, msg.stories...)
		}

	case commentsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.comments = msg.comments
			m.viewport.SetContent(m.renderComments())
			m.viewport.GotoTop()
		}
	}

	return m, tea.Batch(cmds...)
}

// adjustOffset ensures the cursor is visible within the viewport
func (m *Model) adjustOffset() {
	visibleCount := m.visibleStoryCount()
	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+visibleCount {
		m.offset = m.cursor - visibleCount + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

// visibleStoryCount returns how many stories fit on screen
func (m Model) visibleStoryCount() int {
	// Each story takes 2 lines, account for header (2 lines) and status bar (1 line)
	availableLines := m.height - 3
	count := availableLines / 2
	if count < 1 {
		return 1
	}
	return count
}

// View renders the model
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Header with tabs
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Main content
	if m.loading {
		b.WriteString(fmt.Sprintf("\n  %s Loading...\n", m.spinner.View()))
	} else if m.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("\n  Error: %v\n", m.err)))
	} else if m.showHelp {
		b.WriteString(m.renderFullHelp())
	} else {
		switch m.view {
		case StoriesView:
			b.WriteString(m.renderStories())
		case CommentsView:
			b.WriteString(m.viewport.View())
		}
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m Model) renderHeader() string {
	title := HeaderStyle.Render(" HN ")

	var tabs []string
	feedLabels := []string{"Top", "New", "Best", "Ask", "Show"}
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

	// Number
	num := fmt.Sprintf("%3d. ", idx+1)
	if selected {
		num = ScoreStyle.Render(num)
	} else {
		num = MetaStyle.Render(num)
	}
	b.WriteString(num)

	// Title
	title := story.Title
	if len(title) > m.width-20 {
		title = title[:m.width-23] + "..."
	}
	if selected {
		b.WriteString(SelectedTitleStyle.Render(title))
	} else {
		b.WriteString(TitleStyle.Render(title))
	}

	// Domain
	if domain := story.Domain(); domain != "" {
		b.WriteString(" ")
		b.WriteString(URLStyle.Render(fmt.Sprintf("(%s)", domain)))
	}
	b.WriteString("\n")

	// Meta line
	meta := fmt.Sprintf("      %d points by %s %s | %d comments",
		story.Score, story.By, story.TimeAgo(), story.Descendants)
	b.WriteString(MetaStyle.Render(meta))
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderComments() string {
	if m.currentItem == nil {
		return ""
	}

	var b strings.Builder

	// Story header
	b.WriteString(SelectedTitleStyle.Render(m.currentItem.Title))
	b.WriteString("\n")
	if domain := m.currentItem.Domain(); domain != "" {
		b.WriteString(URLStyle.Render(fmt.Sprintf("(%s)", domain)))
		b.WriteString("\n")
	}
	meta := fmt.Sprintf("%d points by %s %s",
		m.currentItem.Score, m.currentItem.By, m.currentItem.TimeAgo())
	b.WriteString(MetaStyle.Render(meta))
	b.WriteString("\n\n")

	// Story text (for Ask HN, etc.)
	if m.currentItem.Text != "" {
		text := cleanHTML(m.currentItem.Text)
		b.WriteString(CommentTextStyle.Render(wrapText(text, m.width-4)))
		b.WriteString("\n\n")
	}

	b.WriteString(MetaStyle.Render(fmt.Sprintf("─── %d comments ───", m.currentItem.Descendants)))
	b.WriteString("\n\n")

	// Comments
	for _, comment := range m.comments {
		b.WriteString(m.renderComment(comment))
	}

	return b.String()
}

func (m Model) renderComment(c *api.Comment) string {
	var b strings.Builder

	indent := strings.Repeat("  ", c.Depth)
	prefix := IndentStyle(c.Depth).Render("│ ")

	// Author line
	author := CommentAuthorStyle.Render(c.By)
	time := CommentMetaStyle.Render(c.TimeAgo())
	b.WriteString(indent + prefix + author + " " + time + "\n")

	// Comment text
	text := cleanHTML(c.Text)
	lines := wrapTextLines(text, m.width-len(indent)-4)
	for _, line := range lines {
		b.WriteString(indent + prefix + CommentTextStyle.Render(line) + "\n")
	}
	b.WriteString(indent + prefix + "\n")

	// Children
	for _, child := range c.Children {
		b.WriteString(m.renderComment(child))
	}

	return b.String()
}

func (m Model) renderStatusBar() string {
	var left, right string

	switch m.view {
	case StoriesView:
		left = fmt.Sprintf(" %d/%d stories", m.cursor+1, len(m.stories))
		right = "↑↓:nav  enter:open  c:comments  tab:feed  ?:help  q:quit "
	case CommentsView:
		left = fmt.Sprintf(" %d comments", len(m.comments))
		right = "↑↓:scroll  b:back  ?:help  q:quit "
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return StatusBarStyle.Width(m.width).Render(left + strings.Repeat(" ", gap) + right)
}

func (m Model) renderFullHelp() string {
	return "\n" + m.help.FullHelpView(m.keys.FullHelp()) + "\n\nPress any key to close help."
}

// Helper functions

func cleanHTML(s string) string {
	// Decode HTML entities
	s = html.UnescapeString(s)

	// Replace <p> tags with newlines
	s = regexp.MustCompile(`<p>`).ReplaceAllString(s, "\n\n")

	// Replace <br> tags
	s = regexp.MustCompile(`<br\s*/?\s*>`).ReplaceAllString(s, "\n")

	// Replace links with just the text
	s = regexp.MustCompile(`<a\s+href="([^"]*)"[^>]*>([^<]*)</a>`).ReplaceAllString(s, "$2 [$1]")

	// Remove remaining HTML tags
	s = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")

	// Clean up whitespace
	s = strings.TrimSpace(s)

	return s
}

func wrapText(s string, width int) string {
	return strings.Join(wrapTextLines(s, width), "\n")
}

func wrapTextLines(s string, width int) []string {
	if width <= 0 {
		width = 80
	}

	var lines []string
	paragraphs := strings.Split(s, "\n")

	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
		lines = append(lines, currentLine)
	}

	return lines
}
