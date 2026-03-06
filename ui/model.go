package ui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/JonathanWThom/feedme/api"
	"github.com/pkg/browser"
)

// View represents the current view
type View int

const (
	StoriesView View = iota
	CommentsView
	SourcePickerView
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

type updateCheckMsg struct {
	info *api.UpdateInfo
}

// Model is the main application model
type Model struct {
	source   api.Source
	keys     KeyMap
	help     help.Model
	spinner  spinner.Model
	viewport viewport.Model

	// State
	view         View
	feed         int
	storyIDs     []int
	stories      []*api.Item
	comments     []*api.Comment
	cursor       int
	offset       int
	loading      bool
	err          error
	showHelp     bool
	mouseEnabled bool
	width        int
	height       int
	currentItem  *api.Item

	// Source picker state
	sourcePickerCursor int
	subredditInput     string
	editingSubreddit   bool

	// Visual mode state
	visualMode   bool
	visualStart  int
	visualEnd    int
	commentLines []string

	// Update notification
	updateInfo *api.UpdateInfo
	updateChan <-chan *api.UpdateInfo
}

// New creates a new Model with the default HN source
func New() Model {
	return NewWithSource(api.NewClient(), nil)
}

// NewWithSource creates a new Model with a specific source
func NewWithSource(source api.Source, updateChan <-chan *api.UpdateInfo) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	h := help.New()
	h.Styles.ShortKey = HelpStyle
	h.Styles.ShortDesc = HelpStyle

	return Model{
		source:       source,
		keys:         DefaultKeyMap(),
		help:         h,
		spinner:      s,
		view:         StoriesView,
		feed:         0,
		loading:      true,
		mouseEnabled: true,
		updateChan:   updateChan,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
		m.loadStoryIDs(),
	}
	if m.updateChan != nil {
		cmds = append(cmds, m.waitForUpdate())
	}
	return tea.Batch(cmds...)
}

func (m Model) waitForUpdate() tea.Cmd {
	return func() tea.Msg {
		info := <-m.updateChan
		return updateCheckMsg{info: info}
	}
}

func (m Model) loadStoryIDs() tea.Cmd {
	feedNames := m.source.FeedNames()
	feed := feedNames[m.feed]
	return func() tea.Msg {
		ids, err := m.source.FetchStoryIDs(feed)
		return storyIDsLoadedMsg{ids: ids, err: err}
	}
}

func (m Model) loadStories(ids []int) tea.Cmd {
	return func() tea.Msg {
		stories, err := m.source.FetchItems(ids)
		return storiesLoadedMsg{stories: stories, err: err}
	}
}

func (m Model) loadComments(item *api.Item) tea.Cmd {
	return func() tea.Msg {
		comments, err := m.source.FetchCommentTree(item, 0)
		return commentsLoadedMsg{comments: comments, err: err}
	}
}

// resetForNewSource resets state when switching sources or feeds
func (m *Model) resetForNewSource() {
	m.view = StoriesView
	m.feed = 0
	m.stories = nil
	m.storyIDs = nil
	m.cursor = 0
	m.offset = 0
	m.err = nil
	m.loading = true
}

// resetForNewFeed resets state when switching feeds within same source
func (m *Model) resetForNewFeed() {
	m.stories = nil
	m.storyIDs = nil
	m.cursor = 0
	m.offset = 0
	m.err = nil
	m.loading = true
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.view == SourcePickerView {
			return m.handleSourcePickerInput(msg)
		}

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
				if m.visualMode {
					m.visualEnd = m.viewport.YOffset
					m.updateViewportWithHighlight()
				}
			}

		case key.Matches(msg, m.keys.Down):
			if m.view == StoriesView && m.cursor < len(m.stories)-1 {
				m.cursor++
				m.adjustOffset()
				if m.cursor >= len(m.stories)-5 && len(m.storyIDs) > len(m.stories) {
					nextBatch := m.storyIDs[len(m.stories):min(len(m.stories)+30, len(m.storyIDs))]
					if len(nextBatch) > 0 {
						m.loading = true
						return m, m.loadStories(nextBatch)
					}
				}
			} else if m.view == CommentsView {
				m.viewport.LineDown(1)
				if m.visualMode {
					m.visualEnd = m.viewport.YOffset
					m.updateViewportWithHighlight()
				}
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
			var story *api.Item
			if m.view == StoriesView && len(m.stories) > 0 {
				story = m.stories[m.cursor]
			} else if m.view == CommentsView {
				story = m.currentItem
			}
			if story != nil && story.URL != "" {
				_ = browser.OpenURL(story.URL)
			} else if story != nil {
				url := m.source.StoryURL(story)
				_ = browser.OpenURL(url)
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
			if m.visualMode {
				m.visualMode = false
				m.updateViewportWithHighlight()
				return m, nil
			}
			if m.view == CommentsView {
				m.view = StoriesView
				m.comments = nil
				m.commentLines = nil
			}

		case key.Matches(msg, m.keys.Visual):
			if m.view == CommentsView && !m.visualMode {
				m.visualMode = true
				m.visualStart = m.viewport.YOffset
				m.visualEnd = m.viewport.YOffset
				m.updateViewportWithHighlight()
			}

		case key.Matches(msg, m.keys.Yank):
			if m.view == CommentsView && m.visualMode {
				m.yankSelection()
				m.visualMode = false
				m.updateViewportWithHighlight()
			}

		case key.Matches(msg, m.keys.NextTab):
			if m.view == StoriesView {
				feedNames := m.source.FeedNames()
				m.feed = (m.feed + 1) % len(feedNames)
				m.resetForNewFeed()
				return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
			}

		case key.Matches(msg, m.keys.PrevTab):
			if m.view == StoriesView {
				feedNames := m.source.FeedNames()
				m.feed = (m.feed - 1 + len(feedNames)) % len(feedNames)
				m.resetForNewFeed()
				return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
			}

		case key.Matches(msg, m.keys.Refresh):
			m.resetForNewFeed()
			return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())

		case key.Matches(msg, m.keys.ToggleMouse):
			m.mouseEnabled = !m.mouseEnabled
			if m.mouseEnabled {
				return m, tea.EnableMouseCellMotion
			}
			return m, tea.DisableMouse

		case key.Matches(msg, m.keys.SwitchSource):
			if m.view == StoriesView {
				m.view = SourcePickerView
				m.sourcePickerCursor = 0
				m.subredditInput = ""
				m.editingSubreddit = false
				return m, nil
			}
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
			batchSize := min(30, len(msg.ids))
			return m, m.loadStories(msg.ids[:batchSize])
		}

	case storiesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			for _, s := range msg.stories {
				if s != nil {
					m.stories = append(m.stories, s)
				}
			}
		}

	case commentsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.comments = msg.comments
			content := m.renderComments()
			m.commentLines = strings.Split(content, "\n")
			m.viewport.SetContent(content)
			m.viewport.GotoTop()
		}

	case updateCheckMsg:
		if msg.info != nil && msg.info.HasUpdate() {
			m.updateInfo = msg.info
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

// View renders the model
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

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
		case SourcePickerView:
			b.WriteString(m.renderSourcePicker())
		}
	}

	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// updateViewportWithHighlight re-renders the viewport with visual selection highlighting
func (m *Model) updateViewportWithHighlight() {
	if len(m.commentLines) == 0 {
		return
	}

	yOffset := m.viewport.YOffset

	var lines []string
	start, end := m.visualStart, m.visualEnd
	if start > end {
		start, end = end, start
	}

	for i, line := range m.commentLines {
		if m.visualMode && i >= start && i <= end {
			lines = append(lines, VisualSelectStyle.Render(line))
		} else {
			lines = append(lines, line)
		}
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.viewport.SetYOffset(yOffset)
}

// yankSelection copies the selected text to the clipboard
func (m *Model) yankSelection() {
	if len(m.commentLines) == 0 {
		return
	}

	start, end := m.visualStart, m.visualEnd
	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if end >= len(m.commentLines) {
		end = len(m.commentLines) - 1
	}

	selected := m.commentLines[start : end+1]
	text := strings.Join(selected, "\n")
	text = stripAnsi(text)

	clipboard.WriteAll(text)
}
