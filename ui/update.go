package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		m.handlePageDown()

	case key.Matches(msg, m.keys.PageUp):
		m.handlePageUp()

	case key.Matches(msg, m.keys.Home):
		m.handleHome()

	case key.Matches(msg, m.keys.End):
		m.handleEnd()

	case key.Matches(msg, m.keys.Enter), key.Matches(msg, m.keys.Open):
		m.openCurrentURL()

	case key.Matches(msg, m.keys.Comments):
		return m.openComments()

	case key.Matches(msg, m.keys.Back):
		return m.handleBack()

	case key.Matches(msg, m.keys.Visual):
		m.startVisualMode()

	case key.Matches(msg, m.keys.Yank):
		m.handleYank()

	case key.Matches(msg, m.keys.NextTab):
		return m.switchFeed(1)

	case key.Matches(msg, m.keys.PrevTab):
		return m.switchFeed(-1)

	case key.Matches(msg, m.keys.Refresh):
		m.resetForNewFeed()
		return m, tea.Batch(m.spinner.Tick, m.loadStoryIDs())

	case key.Matches(msg, m.keys.ToggleMouse):
		return m.toggleMouse()

	case key.Matches(msg, m.keys.SwitchSource):
		return m.openSourcePicker()
	}

	return m, nil
}
