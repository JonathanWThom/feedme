package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JonathanWThom/feedme/api"
)

// MockSource implements api.Source for testing
type MockSource struct {
	name        string
	feedNames   []string
	feedLabels  []string
	storyIDs    []int
	items       map[int]*api.Item
	comments    []*api.Comment
	fetchError  error
}

func NewMockSource() *MockSource {
	return &MockSource{
		name:       "Mock",
		feedNames:  []string{"feed1", "feed2"},
		feedLabels: []string{"Feed 1", "Feed 2"},
		storyIDs:   []int{1, 2, 3},
		items: map[int]*api.Item{
			1: {ID: 1, Title: "Test Story 1", By: "author1", Score: 100, URL: "https://example.com/1", Time: 1704067200, Descendants: 5},
			2: {ID: 2, Title: "Test Story 2", By: "author2", Score: 50, URL: "https://example.com/2", Time: 1704067200, Descendants: 0},
			3: {ID: 3, Title: "Test Story 3", By: "author3", Score: 25, URL: "", Time: 1704067200, Descendants: 10, Text: "Ask HN: Test question?"},
		},
		comments: []*api.Comment{
			{Item: &api.Item{By: "commenter1", Text: "Test comment 1", Time: 1704067200}, Depth: 0},
			{Item: &api.Item{By: "commenter2", Text: "Test reply", Time: 1704067200}, Depth: 1},
		},
	}
}

func (m *MockSource) Name() string                                  { return m.name }
func (m *MockSource) FeedNames() []string                           { return m.feedNames }
func (m *MockSource) FeedLabels() []string                          { return m.feedLabels }
func (m *MockSource) StoryURL(item *api.Item) string                { return "https://mock.test/item/" }
func (m *MockSource) FetchStoryIDs(feed string) ([]int, error)      { return m.storyIDs, m.fetchError }
func (m *MockSource) FetchItem(id int) (*api.Item, error)           { return m.items[id], m.fetchError }
func (m *MockSource) FetchItems(ids []int) ([]*api.Item, error) {
	items := make([]*api.Item, len(ids))
	for i, id := range ids {
		items[i] = m.items[id]
	}
	return items, m.fetchError
}
func (m *MockSource) FetchCommentTree(item *api.Item, maxDepth int) ([]*api.Comment, error) {
	return m.comments, m.fetchError
}

// TestNewModel tests model creation
func TestNewModel(t *testing.T) {
	model := New()

	if model.source == nil {
		t.Error("source should not be nil")
	}
	if model.view != StoriesView {
		t.Errorf("expected StoriesView, got %v", model.view)
	}
	if !model.loading {
		t.Error("model should start in loading state")
	}
	if !model.mouseEnabled {
		t.Error("mouse should be enabled by default")
	}
}

// TestNewWithSource tests model creation with custom source
func TestNewWithSource(t *testing.T) {
	mockSource := NewMockSource()
	updateChan := make(chan *api.UpdateInfo, 1)

	model := NewWithSource(mockSource, updateChan)

	if model.source.Name() != "Mock" {
		t.Errorf("expected source name 'Mock', got '%s'", model.source.Name())
	}
	if model.updateChan == nil {
		t.Error("update channel should not be nil")
	}
}

// TestModelInit tests initialization
func TestModelInit(t *testing.T) {
	mockSource := NewMockSource()
	model := NewWithSource(mockSource, nil)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

// TestUpdateKeyQuit tests quit key handling
func TestUpdateKeyQuit(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24

	// Test 'q' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("expected quit command")
	}
}

// TestUpdateKeyNavigation tests navigation keys
func TestUpdateKeyNavigation(t *testing.T) {
	mockSource := NewMockSource()
	model := NewWithSource(mockSource, nil)
	model.width = 80
	model.height = 24
	model.loading = false
	model.stories = []*api.Item{
		mockSource.items[1],
		mockSource.items[2],
		mockSource.items[3],
	}

	// Test down key
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after down, got %d", m.cursor)
	}

	// Test up key
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after up, got %d", m.cursor)
	}
}

// TestUpdateKeyHelp tests help toggle
func TestUpdateKeyHelp(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24

	// Toggle help on
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)
	if !m.showHelp {
		t.Error("help should be shown after pressing ?")
	}

	// Toggle help off
	newModel, _ = m.Update(msg)
	m = newModel.(Model)
	if m.showHelp {
		t.Error("help should be hidden after pressing ? again")
	}
}

// TestUpdateWindowSize tests window resize handling
func TestUpdateWindowSize(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

// TestUpdateStoryIDsLoaded tests story IDs loading
func TestUpdateStoryIDsLoaded(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24
	model.loading = true

	msg := storyIDsLoadedMsg{ids: []int{1, 2, 3}, err: nil}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if len(m.storyIDs) != 3 {
		t.Errorf("expected 3 story IDs, got %d", len(m.storyIDs))
	}
	if cmd == nil {
		t.Error("expected command to load stories")
	}
}

// TestUpdateStoriesLoaded tests stories loading
func TestUpdateStoriesLoaded(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24
	model.loading = true

	stories := []*api.Item{
		{ID: 1, Title: "Story 1"},
		{ID: 2, Title: "Story 2"},
	}
	msg := storiesLoadedMsg{stories: stories, err: nil}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.loading {
		t.Error("loading should be false after stories loaded")
	}
	if len(m.stories) != 2 {
		t.Errorf("expected 2 stories, got %d", len(m.stories))
	}
}

// TestUpdateCommentsLoaded tests comments loading
func TestUpdateCommentsLoaded(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24
	model.loading = true
	model.view = CommentsView
	model.currentItem = &api.Item{ID: 1, Title: "Test", Descendants: 5}

	comments := []*api.Comment{
		{Item: &api.Item{By: "user1", Text: "comment"}, Depth: 0},
	}
	msg := commentsLoadedMsg{comments: comments, err: nil}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.loading {
		t.Error("loading should be false after comments loaded")
	}
	if len(m.comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(m.comments))
	}
}

// TestUpdateUpdateCheck tests update notification
func TestUpdateUpdateCheck(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24

	info := &api.UpdateInfo{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.1.0",
		UpdateURL:      "https://example.com",
	}
	msg := updateCheckMsg{info: info}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.updateInfo == nil {
		t.Error("update info should be set")
	}
	if m.updateInfo.LatestVersion != "v1.1.0" {
		t.Errorf("expected version v1.1.0, got %s", m.updateInfo.LatestVersion)
	}
}

// TestViewLoading tests loading view
func TestViewLoading(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24
	model.loading = true

	view := model.View()
	if !strings.Contains(view, "Loading") {
		t.Error("loading view should contain 'Loading'")
	}
}

// TestViewStories tests stories view
func TestViewStories(t *testing.T) {
	mockSource := NewMockSource()
	model := NewWithSource(mockSource, nil)
	model.width = 80
	model.height = 24
	model.loading = false
	model.stories = []*api.Item{mockSource.items[1]}

	view := model.View()
	if !strings.Contains(view, "Test Story 1") {
		t.Error("stories view should contain story title")
	}
}

// TestViewSourcePicker tests source picker view
func TestViewSourcePicker(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24
	model.view = SourcePickerView
	model.loading = false // Must be false to show source picker

	view := model.View()
	if !strings.Contains(view, "Hacker News") {
		t.Error("source picker should show Hacker News option")
	}
	if !strings.Contains(view, "Lobste.rs") {
		t.Error("source picker should show Lobste.rs option")
	}
	if !strings.Contains(view, "Reddit") {
		t.Error("source picker should show Reddit option")
	}
}

// TestVisibleStoryCount tests visible story count calculation
func TestVisibleStoryCount(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)

	testCases := []struct {
		height   int
		expected int
	}{
		{24, 10},  // (24 - 3) / 2 = 10
		{10, 3},   // (10 - 3) / 2 = 3
		{5, 1},    // minimum 1
	}

	for _, tc := range testCases {
		model.height = tc.height
		result := model.visibleStoryCount()
		if result != tc.expected {
			t.Errorf("height %d: expected %d, got %d", tc.height, tc.expected, result)
		}
	}
}

// TestAdjustOffset tests offset adjustment for cursor visibility
func TestAdjustOffset(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.height = 10 // 3 visible stories

	// Cursor within view
	model.cursor = 1
	model.offset = 0
	model.adjustOffset()
	if model.offset != 0 {
		t.Errorf("offset should stay 0, got %d", model.offset)
	}

	// Cursor below view
	model.cursor = 5
	model.adjustOffset()
	if model.offset == 0 {
		t.Error("offset should increase when cursor is below view")
	}

	// Cursor above view
	model.cursor = 0
	model.offset = 5
	model.adjustOffset()
	if model.offset != 0 {
		t.Errorf("offset should be 0 when cursor is 0, got %d", model.offset)
	}
}

// TestRenderHeader tests header rendering
func TestRenderHeader(t *testing.T) {
	mockSource := NewMockSource()
	model := NewWithSource(mockSource, nil)
	model.width = 80
	model.feed = 0

	header := model.renderHeader()
	if !strings.Contains(header, "Mock") {
		t.Error("header should contain source name")
	}
	if !strings.Contains(header, "Feed 1") {
		t.Error("header should contain feed label")
	}
}

// TestRenderStatusBar tests status bar rendering
func TestRenderStatusBar(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.stories = []*api.Item{{}, {}, {}}
	model.cursor = 1

	statusBar := model.renderStatusBar()
	if !strings.Contains(statusBar, "2/3") {
		t.Error("status bar should show cursor position")
	}
}

// TestCleanHTML tests HTML cleaning
func TestCleanHTML(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "paragraph tags",
			input:    "<p>Hello</p><p>World</p>",
			expected: "Hello\n\nWorld",
		},
		{
			name:     "br tags",
			input:    "Line 1<br>Line 2<br/>Line 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "links",
			input:    `<a href="https://example.com">Example</a>`,
			expected: "Example [https://example.com]",
		},
		{
			name:     "HTML entities",
			input:    "Hello &amp; World &quot;test&quot;",
			expected: "Hello & World \"test\"",
		},
		{
			name:     "strip other tags",
			input:    "<div><span>Text</span></div>",
			expected: "Text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cleanHTML(tc.input)
			if result != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestWrapText tests text wrapping
func TestWrapText(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "no wrap needed",
			input:    "short text",
			width:    80,
			expected: "short text",
		},
		{
			name:     "wrap long line",
			input:    "this is a long line that should wrap",
			width:    20,
			expected: "this is a long line\nthat should wrap",
		},
		{
			name:     "preserve paragraphs",
			input:    "para 1\n\npara 2",
			width:    80,
			expected: "para 1\n\npara 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := wrapText(tc.input, tc.width)
			if result != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestWrapTextLines tests text wrapping to lines
func TestWrapTextLines(t *testing.T) {
	// "word1 word2 word3" at width 10:
	// "word1" (5 chars) + space + "word2" = 11 chars > 10, so wrap
	// Line 1: "word1", Line 2: "word2", Line 3: "word3"
	result := wrapTextLines("word1 word2 word3", 10)
	if len(result) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(result), result)
	}
}

// TestWrapTextLinesZeroWidth tests zero width handling
func TestWrapTextLinesZeroWidth(t *testing.T) {
	result := wrapTextLines("test", 0)
	// Should use default width of 80
	if len(result) != 1 {
		t.Errorf("expected 1 line, got %d", len(result))
	}
}

// TestStripAnsi tests ANSI code stripping
func TestStripAnsi(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no ansi",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "with color codes",
			input:    "\x1b[31mred\x1b[0m text",
			expected: "red text",
		},
		{
			name:     "multiple codes",
			input:    "\x1b[1;31;40mbold red on black\x1b[0m",
			expected: "bold red on black",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripAnsi(tc.input)
			if result != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestSourcePickerNavigation tests source picker navigation
func TestSourcePickerNavigation(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24
	model.view = SourcePickerView
	model.sourcePickerCursor = 0

	// Navigate down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)
	if m.sourcePickerCursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.sourcePickerCursor)
	}

	// Navigate up
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)
	if m.sourcePickerCursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.sourcePickerCursor)
	}
}

// TestToggleMouse tests mouse toggle
func TestToggleMouse(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24
	model.mouseEnabled = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.mouseEnabled {
		t.Error("mouse should be disabled after toggle")
	}
	if cmd == nil {
		t.Error("should return disable mouse command")
	}
}

// TestFeedNavigation tests feed tab navigation
func TestFeedNavigation(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80
	model.height = 24
	model.loading = false
	model.feed = 0

	// Next tab
	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)
	if m.feed != 1 {
		t.Errorf("expected feed 1, got %d", m.feed)
	}

	// Prev tab (wraps around)
	msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)
	if m.feed != 0 {
		t.Errorf("expected feed 0, got %d", m.feed)
	}
}

// TestViewConstants tests View type constants
func TestViewConstants(t *testing.T) {
	if StoriesView != 0 {
		t.Errorf("StoriesView should be 0, got %d", StoriesView)
	}
	if CommentsView != 1 {
		t.Errorf("CommentsView should be 1, got %d", CommentsView)
	}
	if SourcePickerView != 2 {
		t.Errorf("SourcePickerView should be 2, got %d", SourcePickerView)
	}
}

// TestRenderFullHelp tests full help rendering
func TestRenderFullHelp(t *testing.T) {
	model := NewWithSource(NewMockSource(), nil)
	model.width = 80

	help := model.renderFullHelp()
	if !strings.Contains(help, "help") || !strings.Contains(help, "quit") {
		t.Error("full help should contain key descriptions")
	}
}

// TestSourceOptions tests that source options are defined
func TestSourceOptions(t *testing.T) {
	if len(sourceOptions) != 3 {
		t.Errorf("expected 3 source options, got %d", len(sourceOptions))
	}
	if sourceOptions[0] != "Hacker News" {
		t.Errorf("first option should be 'Hacker News', got '%s'", sourceOptions[0])
	}
	if sourceOptions[1] != "Lobste.rs" {
		t.Errorf("second option should be 'Lobste.rs', got '%s'", sourceOptions[1])
	}
	if sourceOptions[2] != "Reddit" {
		t.Errorf("third option should be 'Reddit', got '%s'", sourceOptions[2])
	}
}
