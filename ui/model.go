package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JonathanWThom/feedme/api"
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

// resetForNewSource resets state when switching sources
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

// resetForNewFeed resets state when switching feeds
func (m *Model) resetForNewFeed() {
	m.stories = nil
	m.storyIDs = nil
	m.cursor = 0
	m.offset = 0
	m.err = nil
	m.loading = true
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
		fmt.Fprintf(&b, "\n  %s Loading...\n", m.spinner.View())
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

