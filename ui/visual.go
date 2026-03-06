package ui

import (
	"strings"

	"github.com/atotto/clipboard"
)

// updateViewportWithHighlight re-renders the viewport with visual selection
func (m *Model) updateViewportWithHighlight() {
	if len(m.commentLines) == 0 {
		return
	}

	yOffset := m.viewport.YOffset

	var lines []string
	start, end := m.normalizedSelection()

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

	start, end := m.normalizedSelection()
	start = max(start, 0)
	end = min(end, len(m.commentLines)-1)

	selected := m.commentLines[start : end+1]
	text := strings.Join(selected, "\n")
	text = stripAnsi(text)

	clipboard.WriteAll(text)
}

// normalizedSelection returns start/end with start <= end
func (m *Model) normalizedSelection() (int, int) {
	start, end := m.visualStart, m.visualEnd
	if start > end {
		start, end = end, start
	}
	return start, end
}
