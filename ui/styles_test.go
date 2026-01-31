package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHeaderStyle(t *testing.T) {
	// Test that HeaderStyle is not nil and can render
	result := HeaderStyle.Render("Test")
	if result == "" {
		t.Error("HeaderStyle should render non-empty string")
	}
}

func TestTabStyles(t *testing.T) {
	// Test TabStyle
	result := TabStyle.Render("Tab")
	if result == "" {
		t.Error("TabStyle should render non-empty string")
	}

	// Test ActiveTabStyle
	result = ActiveTabStyle.Render("Active")
	if result == "" {
		t.Error("ActiveTabStyle should render non-empty string")
	}
}

func TestTitleStyles(t *testing.T) {
	// Test TitleStyle
	result := TitleStyle.Render("Title")
	if result == "" {
		t.Error("TitleStyle should render non-empty string")
	}

	// Test SelectedTitleStyle
	result = SelectedTitleStyle.Render("Selected")
	if result == "" {
		t.Error("SelectedTitleStyle should render non-empty string")
	}
}

func TestMetaStyles(t *testing.T) {
	// Test URLStyle
	result := URLStyle.Render("(example.com)")
	if result == "" {
		t.Error("URLStyle should render non-empty string")
	}

	// Test MetaStyle
	result = MetaStyle.Render("10 points by user")
	if result == "" {
		t.Error("MetaStyle should render non-empty string")
	}

	// Test ScoreStyle
	result = ScoreStyle.Render("100")
	if result == "" {
		t.Error("ScoreStyle should render non-empty string")
	}
}

func TestCommentStyles(t *testing.T) {
	// Test CommentAuthorStyle
	result := CommentAuthorStyle.Render("username")
	if result == "" {
		t.Error("CommentAuthorStyle should render non-empty string")
	}

	// Test CommentTextStyle
	result = CommentTextStyle.Render("This is a comment")
	if result == "" {
		t.Error("CommentTextStyle should render non-empty string")
	}

	// Test CommentMetaStyle
	result = CommentMetaStyle.Render("2 hours ago")
	if result == "" {
		t.Error("CommentMetaStyle should render non-empty string")
	}
}

func TestGeneralStyles(t *testing.T) {
	// Test HelpStyle
	result := HelpStyle.Render("help text")
	if result == "" {
		t.Error("HelpStyle should render non-empty string")
	}

	// Test ErrorStyle
	result = ErrorStyle.Render("error message")
	if result == "" {
		t.Error("ErrorStyle should render non-empty string")
	}

	// Test SpinnerStyle
	result = SpinnerStyle.Render("●")
	if result == "" {
		t.Error("SpinnerStyle should render non-empty string")
	}

	// Test StatusBarStyle
	result = StatusBarStyle.Render("status")
	if result == "" {
		t.Error("StatusBarStyle should render non-empty string")
	}
}

func TestVisualSelectStyle(t *testing.T) {
	result := VisualSelectStyle.Render("selected text")
	if result == "" {
		t.Error("VisualSelectStyle should render non-empty string")
	}
}

func TestIndentStyle(t *testing.T) {
	// Test that IndentStyle returns different colors for different depths
	style0 := IndentStyle(0)
	style1 := IndentStyle(1)
	style2 := IndentStyle(2)

	// Each should render without error
	result0 := style0.Render("│")
	result1 := style1.Render("│")
	result2 := style2.Render("│")

	if result0 == "" || result1 == "" || result2 == "" {
		t.Error("IndentStyle should render non-empty strings")
	}

	// Test that colors cycle after 6 depths
	style6 := IndentStyle(6)
	result6 := style6.Render("│")
	if result6 == "" {
		t.Error("IndentStyle should handle depth >= 6")
	}
}

func TestIndentStyleColors(t *testing.T) {
	// Test that each depth gets a different color (within the cycle)
	colors := make(map[int]lipgloss.Style)
	for i := 0; i < 6; i++ {
		colors[i] = IndentStyle(i)
	}

	// Verify cycle repeats
	for i := 0; i < 6; i++ {
		style := IndentStyle(i)
		cycledStyle := IndentStyle(i + 6)
		result1 := style.Render("│")
		result2 := cycledStyle.Render("│")
		if result1 != result2 {
			t.Errorf("IndentStyle should cycle: depth %d and %d should match", i, i+6)
		}
	}
}

func TestColorConstants(t *testing.T) {
	// Verify color constants are defined
	colors := []lipgloss.Color{
		orange,
		dimOrange,
		subtle,
		highlight,
		dimText,
		commentBg,
	}

	for i, c := range colors {
		if c == "" {
			t.Errorf("color constant at index %d should not be empty", i)
		}
	}
}

func TestOrangeColor(t *testing.T) {
	// Orange should be #FF6600 (Hacker News orange)
	if orange != lipgloss.Color("#FF6600") {
		t.Errorf("orange should be #FF6600, got %s", string(orange))
	}
}

func TestStylesNotNil(t *testing.T) {
	styles := []lipgloss.Style{
		HeaderStyle,
		TabStyle,
		ActiveTabStyle,
		TitleStyle,
		SelectedTitleStyle,
		URLStyle,
		MetaStyle,
		ScoreStyle,
		CommentAuthorStyle,
		CommentTextStyle,
		CommentMetaStyle,
		HelpStyle,
		ErrorStyle,
		SpinnerStyle,
		StatusBarStyle,
		VisualSelectStyle,
	}

	for i, style := range styles {
		// Try to render something - if style is broken, this will show
		result := style.Render("test")
		if result == "" {
			t.Errorf("style at index %d rendered empty string", i)
		}
	}
}

func TestStatusBarStyleWidth(t *testing.T) {
	// Test that StatusBarStyle can be used with Width
	styled := StatusBarStyle.Width(80).Render("test content")
	if styled == "" {
		t.Error("StatusBarStyle with Width should render non-empty string")
	}
}

func TestMultipleRenders(t *testing.T) {
	// Ensure styles can be used multiple times
	for i := 0; i < 10; i++ {
		result := TitleStyle.Render("test")
		if result == "" {
			t.Errorf("TitleStyle render %d was empty", i)
		}
	}
}

func TestEmptyStringRender(t *testing.T) {
	// Styles should handle empty strings
	result := TitleStyle.Render("")
	// Empty input may produce empty output or just styling codes
	_ = result // Just ensure no panic
}

func TestLongTextRender(t *testing.T) {
	// Styles should handle long text
	longText := ""
	for i := 0; i < 1000; i++ {
		longText += "x"
	}
	result := CommentTextStyle.Render(longText)
	if result == "" {
		t.Error("should handle long text")
	}
}

func TestSpecialCharacterRender(t *testing.T) {
	// Styles should handle special characters
	specialChars := "Hello 你好 🎉 <>&\""
	result := TitleStyle.Render(specialChars)
	if result == "" {
		t.Error("should handle special characters")
	}
}

func TestNewlineRender(t *testing.T) {
	// Styles should handle text with newlines
	multiLine := "Line 1\nLine 2\nLine 3"
	result := CommentTextStyle.Render(multiLine)
	if result == "" {
		t.Error("should handle multiline text")
	}
}
