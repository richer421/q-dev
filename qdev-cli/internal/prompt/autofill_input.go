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

// AutoFillInput 是一个支持在空输入时按 Tab 键自动填充的输入字段
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

	// 内部 textinput
	ti textinput.Model
}

// NewAutoFillInput 创建一个新的 AutoFillInput
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

// Suggestion 设置单个建议值（用于 Tab 自动填充）
func (i *AutoFillInput) Suggestion(s string) *AutoFillInput {
	i.suggestion = s
	i.ti.ShowSuggestions = true
	i.ti.KeyMap.AcceptSuggestion.SetEnabled(true)
	i.ti.SetSuggestions([]string{s})
	return i
}

// Suggestions 设置建议列表（兼容 huh 接口）
func (i *AutoFillInput) Suggestions(suggestions []string) *AutoFillInput {
	if len(suggestions) > 0 {
		i.Suggestion(suggestions[0])
	}
	return i
}

// Value 设置值指针
func (i *AutoFillInput) Value(value *string) *AutoFillInput {
	i.accessor = huh.NewPointerAccessor(value)
	i.ti.SetValue(i.accessor.Get())
	return i
}

// Title 设置标题
func (i *AutoFillInput) Title(title string) *AutoFillInput {
	i.title = title
	return i
}

// Description 设置描述
func (i *AutoFillInput) Description(desc string) *AutoFillInput {
	i.desc = desc
	return i
}

// Placeholder 设置占位符
func (i *AutoFillInput) Placeholder(placeholder string) *AutoFillInput {
	i.placeholder = placeholder
	i.ti.Placeholder = placeholder
	return i
}

// Validate 设置验证函数
func (i *AutoFillInput) Validate(validate func(string) error) *AutoFillInput {
	i.validate = validate
	return i
}

// Inline 设置内联模式
func (i *AutoFillInput) Inline(inline bool) *AutoFillInput {
	return i
}

// Key 设置键
func (i *AutoFillInput) Key(key string) *AutoFillInput {
	i.key = key
	return i
}

// CharLimit 设置字符限制
func (i *AutoFillInput) CharLimit(limit int) *AutoFillInput {
	i.ti.CharLimit = limit
	return i
}

// EchoMode 设置回显模式
func (i *AutoFillInput) EchoMode(mode huh.EchoMode) *AutoFillInput {
	i.ti.EchoMode = textinput.EchoMode(mode)
	return i
}

// Password 设置密码模式
func (i *AutoFillInput) Password(password bool) *AutoFillInput {
	if password {
		i.ti.EchoMode = textinput.EchoPassword
	} else {
		i.ti.EchoMode = textinput.EchoNormal
	}
	return i
}

// Prompt 设置提示符
func (i *AutoFillInput) Prompt(prompt string) *AutoFillInput {
	i.ti.Prompt = prompt
	return i
}

// Focus 聚焦
func (i *AutoFillInput) Focus() tea.Cmd {
	i.focused = true
	return i.ti.Focus()
}

// Blur 失焦
func (i *AutoFillInput) Blur() tea.Cmd {
	i.focused = false
	i.accessor.Set(i.ti.Value())
	i.ti.Blur()
	i.err = i.validate(i.accessor.Get())
	return nil
}

// KeyBinds 返回键绑定帮助信息
func (i *AutoFillInput) KeyBinds() []key.Binding {
	if i.ti.ShowSuggestions {
		return []key.Binding{i.keymap.AcceptSuggestion, i.keymap.Prev, i.keymap.Submit, i.keymap.Next}
	}
	return []key.Binding{i.keymap.Prev, i.keymap.Submit, i.keymap.Next}
}

// Skip 是否跳过
func (i *AutoFillInput) Skip() bool {
	return false
}

// Zoom 是否缩放
func (i *AutoFillInput) Zoom() bool {
	return false
}

// GetKey 获取键
func (i *AutoFillInput) GetKey() string {
	return i.key
}

// GetValue 获取值
func (i *AutoFillInput) GetValue() any {
	return i.accessor.Get()
}

// Error 返回错误
func (i *AutoFillInput) Error() error {
	return i.err
}

// Run 运行字段
func (i *AutoFillInput) Run() error {
	return huh.Run(i)
}

// RunAccessible 以可访问模式运行
func (i *AutoFillInput) RunAccessible(w io.Writer, r io.Reader) error {
	// 简化实现 - 直接设置默认值
	if i.suggestion != "" && i.accessor.Get() == "" {
		i.accessor.Set(i.suggestion)
		fmt.Fprintf(w, "%s: %s\n", i.title, i.suggestion)
		return nil
	}
	return nil
}

// Init 初始化
func (i *AutoFillInput) Init() tea.Cmd {
	i.ti.Blur()
	return nil
}

// Update 更新
func (i *AutoFillInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		i.err = nil

		// 处理 Tab 键 - 在输入为空时自动填充建议（不跳转)
		if msg.String() == "tab" {
			currentValue := i.ti.Value()
			if currentValue == "" && i.suggestion != "" {
				// 自动填充建议
				i.ti.SetValue(i.suggestion)
				i.accessor.Set(i.suggestion)
				// 不跳转，让用户继续编辑或 return i, nil
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

	// 更新 textinput
	var cmd tea.Cmd
	i.ti, cmd = i.ti.Update(msg)
	cmds = append(cmds, cmd)

	// 同步值到 accessor
	i.accessor.Set(i.ti.Value())

	return i, tea.Batch(cmds...)
}

// View 渲染
func (i *AutoFillInput) View() string {
	styles := i.activeStyles()
	maxWidth := i.width - styles.Base.GetHorizontalFrameSize()

	// 设置 textinput 样式
	i.ti.PlaceholderStyle = styles.TextInput.Placeholder
	i.ti.PromptStyle = styles.TextInput.Prompt
	i.ti.Cursor.Style = styles.TextInput.Cursor
	i.ti.Cursor.TextStyle = styles.TextInput.CursorText
	i.ti.TextStyle = styles.TextInput.Text

	// 调整宽度
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

// wrap 简单的文本换行
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

// activeStyles 获取当前样式
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

// WithKeyMap 设置键映射
func (i *AutoFillInput) WithKeyMap(k *huh.KeyMap) huh.Field {
	i.keymap = k.Input
	i.ti.KeyMap.AcceptSuggestion = i.keymap.AcceptSuggestion
	return i
}

// WithTheme 设置主题
func (i *AutoFillInput) WithTheme(theme *huh.Theme) huh.Field {
	if i.theme != nil {
		return i
	}
	i.theme = theme
	return i
}

// WithWidth 设置宽度
func (i *AutoFillInput) WithWidth(width int) huh.Field {
	i.width = width
	frameSize := i.activeStyles().Base.GetHorizontalFrameSize()
	promptWidth := lipgloss.Width(i.ti.PromptStyle.Render(i.ti.Prompt))
	i.ti.Width = width - frameSize - promptWidth - 1
	return i
}

// WithHeight 设置高度
func (i *AutoFillInput) WithHeight(height int) huh.Field {
	i.height = height
	return i
}

// WithPosition 设置位置
func (i *AutoFillInput) WithPosition(p huh.FieldPosition) huh.Field {
	i.position = p
	i.keymap.Prev.SetEnabled(!p.IsFirst())
	i.keymap.Next.SetEnabled(!p.IsLast())
	i.keymap.Submit.SetEnabled(p.IsLast())
	return i
}

// WithAccessible 设置可访问模式
func (i *AutoFillInput) WithAccessible(accessible bool) huh.Field {
	return i
}

// Accessor 设置 accessor
func (i *AutoFillInput) Accessor(accessor huh.Accessor[string]) *AutoFillInput {
	i.accessor = accessor
	i.ti.SetValue(i.accessor.Get())
	return i
}
