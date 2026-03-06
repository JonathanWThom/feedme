package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/JonathanWThom/feedme/api"
	"github.com/pkg/browser"
)

func (m *Model) handlePageDown() {
	if m.view == StoriesView {
		m.cursor = min(m.cursor+10, len(m.stories)-1)
		m.adjustOffset()
	} else {
		m.viewport.HalfViewDown()
	}
}

func (m *Model) handlePageUp() {
	if m.view == StoriesView {
		m.cursor = max(m.cursor-10, 0)
		m.adjustOffset()
	} else {
		m.viewport.HalfViewUp()
	}
}

func (m *Model) handleHome() {
	if m.view == StoriesView {
		m.cursor = 0
		m.offset = 0
	} else {
		m.viewport.GotoTop()
	}
}

func (m *Model) handleEnd() {
	if m.view == StoriesView {
		m.cursor = len(m.stories) - 1
		m.adjustOffset()
	} else {
		m.viewport.GotoBottom()
	}
}

func (m *Model) startVisualMode() {
	if m.view != CommentsView || m.visualMode {
		return
	}
	m.visualMode = true
	m.visualStart = m.viewport.YOffset
	m.visualEnd = m.viewport.YOffset
	m.updateViewportWithHighlight()
}

func (m *Model) handleYank() {
	if m.view != CommentsView || !m.visualMode {
		return
	}
	m.yankSelection()
	m.visualMode = false
	m.updateViewportWithHighlight()
}

func (m Model) switchFeed(delta int) (tea.Model, tea.Cmd) {
	if m.view != StoriesView {
		return m, nil
	}
	feedNames := m.source.FeedNames()
	m.feed = (m.feed + delta + len(feedNames)) % len(feedNames)
	m.resetForNewFeed()
	return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())
}

func (m Model) toggleMouse() (tea.Model, tea.Cmd) {
	m.mouseEnabled = !m.mouseEnabled
	if m.mouseEnabled {
		return m, tea.EnableMouseCellMotion
	}
	return m, tea.DisableMouse
}

func (m Model) openSourcePicker() (tea.Model, tea.Cmd) {
	if m.view != StoriesView {
		return m, nil
	}
	m.view = SourcePickerView
	m.sourcePickerCursor = 0
	m.subredditInput = ""
	m.editingSubreddit = false
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
		return m.maybeLoadNextBatch()
	}
	if m.view == CommentsView {
		m.viewport.LineDown(1)
		if m.visualMode {
			m.visualEnd = m.viewport.YOffset
			m.updateViewportWithHighlight()
		}
	}
	return m, nil
}

func (m Model) maybeLoadNextBatch() (tea.Model, tea.Cmd) {
	if m.cursor < len(m.stories)-5 || len(m.storyIDs) <= len(m.stories) {
		return m, nil
	}
	end := min(len(m.stories)+30, len(m.storyIDs))
	nextBatch := m.storyIDs[len(m.stories):end]
	if len(nextBatch) == 0 {
		return m, nil
	}
	m.loading = true
	return m, m.loadStories(nextBatch)
}

func (m *Model) openCurrentURL() {
	story := m.currentStory()
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
