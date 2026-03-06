package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/JonathanWThom/feedme/api"
	"github.com/pkg/browser"
)

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

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

// handleKeyMsg handles all keyboard input
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m.handleUp()

	case key.Matches(msg, m.keys.Down):
		return m.handleDown()

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
		m.openCurrentURL()

	case key.Matches(msg, m.keys.Comments):
		return m.openComments()

	case key.Matches(msg, m.keys.Back):
		return m.handleBack()

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

	return m, nil
}

func (m Model) handleUp() (tea.Model, tea.Cmd) {
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
	return m, nil
}

func (m Model) handleDown() (tea.Model, tea.Cmd) {
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
	return m, nil
}

func (m *Model) openCurrentURL() {
	var story = m.currentStory()
	if story == nil {
		return
	}
	if story.URL != "" {
		_ = browser.OpenURL(story.URL)
	} else {
		_ = browser.OpenURL(m.source.StoryURL(story))
	}
}

func (m Model) openComments() (tea.Model, tea.Cmd) {
	if m.view != StoriesView || len(m.stories) == 0 {
		return m, nil
	}
	story := m.stories[m.cursor]
	if story == nil || story.Descendants == 0 {
		return m, nil
	}
	m.currentItem = story
	m.view = CommentsView
	m.loading = true
	m.comments = nil
	return m, tea.Batch(m.spinner.Tick, m.loadComments(story))
}

func (m Model) handleBack() (tea.Model, tea.Cmd) {
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
	return m, nil
}

func (m Model) currentStory() *api.Item {
	if m.view == StoriesView && len(m.stories) > 0 {
		return m.stories[m.cursor]
	}
	if m.view == CommentsView {
		return m.currentItem
	}
	return nil
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
