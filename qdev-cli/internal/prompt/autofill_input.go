package prompt

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AutoFillInput is an input field that supports auto-filling suggestions on Tab
type AutoFillInput struct {
	accessor    huh.Accessor[string]
	key         string
	id          int
	title       string
	desc        string
	placeholder string
	suggestion  string
	validate    func(string) error
	err         error
	focused     bool
	theme       *huh.Theme
	width       int
	height      int
	position    huh.FieldPosition
	keymap      huh.InputKeyMap

	// internal textinput
	ti textinput.Model
}

// NewAutoFillInput creates a new AutoFillInput
func NewAutoFillInput() *AutoFillInput {
	ti := textinput.New()
	ti.KeyMap.AcceptSuggestion = key.NewBinding(key.WithKeys("tab", "right"))
	return &AutoFillInput{
		accessor:   &huh.EmbeddedAccessor[string]{},
		ti:         ti,
		validate:   func(string) error { return nil },
		id:        nextID(),
	}
}

var _idCounter int

func nextID() int {
	_idCounter++
	return _idCounter
}

// Suggestion sets a single suggestion for Tab auto-fill
func (i *AutoFillInput) Suggestion(s string) *AutoFillInput {
	i.suggestion = s
	i.ti.ShowSuggestions = true
	i.ti.KeyMap.AcceptSuggestion.SetEnabled(true)
	i.ti.SetSuggestions([]string{s})
	return i
}

// Suggestions sets suggestions list (compatible with huh interface)
func (i *AutoFillInput) Suggestions(suggestions []string) *AutoFillInput {
	if len(suggestions) > 0 {
		i.Suggestion(suggestions[0])
	}
	return i
}

// Value sets value pointer
func (i *AutoFillInput) Value(value *string) *AutoFillInput {
	i.accessor = huh.NewPointerAccessor(value)
	i.ti.SetValue(i.accessor.Get())
	return i
}

// Title sets title
func (i *AutoFillInput) Title(title string) *AutoFillInput {
	i.title = title
	return i
}

// Description sets description
func (i *AutoFillInput) Description(desc string) *AutoFillInput {
	i.desc = desc
	return i
}

// Placeholder sets placeholder
func (i *AutoFillInput) Placeholder(placeholder string) *AutoFillInput {
	i.placeholder = placeholder
	i.ti.Placeholder = placeholder
	return i
}

// Validate sets validation function
func (i *AutoFillInput) Validate(validate func(string) error) *AutoFillInput {
	i.validate = validate
	return i
}

// Inline sets inline mode
func (i *AutoFillInput) Inline(inline bool) *AutoFillInput {
	return i
}

// Key sets key
func (i *AutoFillInput) Key(key string) *AutoFillInput {
	i.key = key
	return i
}

// CharLimit sets char limit
func (i *AutoFillInput) CharLimit(limit int) *AutoFillInput {
	i.ti.CharLimit = limit
	return i
}

// EchoMode sets echo mode
func (i *AutoFillInput) EchoMode(mode huh.EchoMode) *AutoFillInput {
	i.ti.EchoMode = textinput.EchoMode(mode)
	return i
}

// Password sets password mode
func (i *AutoFillInput) Password(password bool) *AutoFillInput {
	if password {
		i.ti.EchoMode = textinput.EchoPassword
	} else {
		i.ti.EchoMode = textinput.EchoNormal
	}
	return i
}

// Prompt sets prompt
func (i *AutoFillInput) Prompt(prompt string) *AutoFillInput {
	i.ti.Prompt = prompt
	return i
}

// Focus focuses the field
func (i *AutoFillInput) Focus() tea.Cmd {
	i.focused = true
	return i.ti.Focus()
}

// Blur blurs the field
func (i *AutoFillInput) Blur() tea.Cmd {
	i.focused = false
	i.accessor.Set(i.ti.Value())
	i.ti.Blur()
	i.err = i.validate(i.accessor.Get())
	return nil
}

// KeyBinds returns key binding help info
func (i *AutoFillInput) KeyBinds() []key.Binding {
	if i.ti.ShowSuggestions {
		return []key.Binding{i.keymap.AcceptSuggestion, i.keymap.Prev, i.keymap.Submit, i.keymap.Next}
	}
	return []key.Binding{i.keymap.Prev, i.keymap.Submit, i.keymap.Next}
}

// Skip returns whether to skip
func (i *AutoFillInput) Skip() bool {
	return false
}

// Zoom returns whether to zoom
func (i *AutoFillInput) Zoom() bool {
	return false
}

// GetKey returns the key
func (i *AutoFillInput) GetKey() string {
	return i.key
}

// GetValue returns the value
func (i *AutoFillInput) GetValue() any {
	return i.accessor.Get()
}

// Error returns error
func (i *AutoFillInput) Error() error {
	return i.err
}

// Run runs the field
func (i *AutoFillInput) Run() error {
	return huh.Run(i)
}

// RunAccessible runs in accessible mode
func (i *AutoFillInput) RunAccessible(w io.Writer, r io.Reader) error {
	if i.suggestion != "" && i.accessor.Get() == "" {
		i.accessor.Set(i.suggestion)
		fmt.Fprintf(w, "%s: %s\n", i.title, i.suggestion)
		return nil
	}
	return nil
}

// Init initializes
func (i *AutoFillInput) Init() tea.Cmd {
	i.ti.Blur()
	return nil
}

// Update handles updates
func (i *AutoFillInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		i.err = nil

		// Handle Tab key - auto-fill suggestion when input is empty (don't jump)
		if msg.String() == "tab" {
			currentValue := i.ti.Value()
			if currentValue == "" && i.suggestion != "" {
				// Auto-fill suggestion
				i.ti.SetValue(i.suggestion)
				i.accessor.Set(i.suggestion)
				// Don't jump, let user continue editing
				return i, nil
			}
		}

		switch {
		case key.Matches(msg, i.keymap.Prev):
			cmds = append(cmds, huh.PrevField)
		case key.Matches(msg, i.keymap.Next, i.keymap.Submit):
			value := i.ti.Value()
			i.err = i.validate(value)
			if i.err != nil {
				return i, nil
			}
			cmds = append(cmds, huh.NextField)
		}
	}

	// Update textinput
	var cmd tea.Cmd
	i.ti, cmd = i.ti.Update(msg)
	cmds = append(cmds, cmd)

	// Sync value to accessor
	i.accessor.Set(i.ti.Value())

	return i, tea.Batch(cmds...)
}


// View renders the field
func (i *AutoFillInput) View() string {
	styles := i.activeStyles()
	maxWidth := i.width - styles.Base.GetHorizontalFrameSize()

	// Set textinput styles
	i.ti.PlaceholderStyle = styles.TextInput.Placeholder
	i.ti.PromptStyle = styles.TextInput.Prompt
	i.ti.Cursor.Style = styles.TextInput.Cursor
	i.ti.Cursor.TextStyle = styles.TextInput.CursorText
	i.ti.TextStyle = styles.TextInput.Text

	// Adjust width
	if i.ti.CharLimit > 0 {
		i.ti.Width = max3(min2(i.ti.CharLimit, i.ti.Width), maxWidth, 0)
	}

	var sb strings.Builder

	if i.title != "" {
		sb.WriteString(styles.Title.Render(wrap(i.title, maxWidth)))
		sb.WriteString("\n")
	}

	if i.desc != "" {
		sb.WriteString(styles.Description.Render(wrap(i.desc, maxWidth)))
		sb.WriteString("\n")
	}

	sb.WriteString(i.ti.View())

	return styles.Base.
		Width(i.width).
		Height(i.height).
		Render(sb.String())
}


// wrap wraps text
func wrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	return text
}

func max3(a, b, c int) int {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// activeStyles returns active styles
func (i *AutoFillInput) activeStyles() *huh.FieldStyles {
	theme := i.theme
	if theme == nil {
		theme = huh.ThemeCharm()
	}
	if i.focused {
		return &theme.Focused
	}
	return &theme.Blurred
}

// WithKeyMap sets keymap
func (i *AutoFillInput) WithKeyMap(k *huh.KeyMap) huh.Field {
	i.keymap = k.Input
	i.ti.KeyMap.AcceptSuggestion = i.keymap.AcceptSuggestion
	return i
}

// WithTheme sets theme
func (i *AutoFillInput) WithTheme(theme *huh.Theme) huh.Field {
	if i.theme != nil {
		return i
	}
	i.theme = theme
	return i
}

// WithWidth sets width
func (i *AutoFillInput) WithWidth(width int) huh.Field {
	i.width = width
	frameSize := i.activeStyles().Base.GetHorizontalFrameSize()
	promptWidth := lipgloss.Width(i.ti.PromptStyle.Render(i.ti.Prompt))
	i.ti.Width = width - frameSize - promptWidth - 1
	return i
}

// WithHeight sets height
func (i *AutoFillInput) WithHeight(height int) huh.Field {
	i.height = height
	return i
}

// WithPosition sets position
func (i *AutoFillInput) WithPosition(p huh.FieldPosition) huh.Field {
	i.position = p
	i.keymap.Prev.SetEnabled(!p.IsFirst())
	i.keymap.Next.SetEnabled(!p.IsLast())
	i.keymap.Submit.SetEnabled(p.IsLast())
	return i
}

// WithAccessible sets accessible mode
func (i *AutoFillInput) WithAccessible(accessible bool) huh.Field {
	return i
}

// Accessor sets accessor
func (i *AutoFillInput) Accessor(accessor huh.Accessor[string]) *AutoFillInput {
	i.accessor = accessor
	i.ti.SetValue(i.accessor.Get())
	return i
}
