package input

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/rivo/uniseg"
	"github.com/kirick13/newborn/elements"
	"github.com/kirick13/newborn/style"
)

var (
	placeholderColor = lipgloss.Color("#525252")
)

type Model struct {
	textinput.Model
}

func New(placeholder string) *Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = placeholder
	ti.SetVirtualCursor(false)
	ti.SetWidth(45)

	s := ti.Styles()
	s.Cursor.Color = lipgloss.Color("#2563eb")
	s.Cursor.Blink = false
	s.Focused.Text = style.InputFocusedStyle
	s.Focused.Placeholder = style.InputFocusedStyle.Foreground(placeholderColor)
	s.Blurred.Text = style.InputBlurredStyle
	s.Blurred.Placeholder = style.InputBlurredStyle.Foreground(placeholderColor)
	ti.SetStyles(s)

	return &Model{Model: ti}
}

func (m *Model) Update(msg tea.Msg) (elements.Element, tea.Cmd) {
	next, cmd := m.Model.Update(msg)
	m.Model = next
	return m, cmd
}

func (m *Model) Render() string {
	if m.Focused() {
		return m.View()
	}

	styles := m.Styles()
	contentStyle := styles.Blurred.Placeholder
	content := m.Placeholder

	if value := m.Value(); value != "" {
		contentStyle = styles.Blurred.Text
		content = m.maskedValue(value)
	}

	content = trimToWidth(content, m.Width() + 1)
	content = contentStyle.Width(m.Width() + 1).MaxWidth(m.Width() + 1).Render(content)

	return content
}

func (m *Model) maskedValue(value string) string {
	switch m.EchoMode {
	case textinput.EchoPassword:
		return strings.Repeat(string(m.EchoCharacter), len([]rune(value)))
	case textinput.EchoNone:
		return ""
	default:
		return value
	}
}

func trimToWidth(s string, width int) string {
	if width <= 0 || uniseg.StringWidth(s) <= width {
		return s
	}

	runes := []rune(s)
	for len(runes) > 0 && uniseg.StringWidth(string(runes)) > width {
		runes = runes[:len(runes)-1]
	}

	return string(runes)
}
