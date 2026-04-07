package checkbox

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	input "github.com/kirick13/newborn/elements"
	"github.com/kirick13/newborn/style"
)

type Model struct {
	Value bool
	Symbol string
	focus bool
}

func New() *Model {
	return &Model{
		Value: false,
		Symbol: "v",
	}
}

func (m *Model) Update(msg tea.Msg) (input.Element, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ", "space":
			m.Value = !m.Value
		}
	}

	return m, nil
}

func (m *Model) Render() string {
	value := "   "
	if m.Value {
		value = fmt.Sprintf(" %s ", m.Symbol)
	}

	if m.focus {
		return style.InputFocusedStyle.Render(value)
	}

	return style.InputBlurredStyle.Render(value)
}

func (m *Model) Focus() tea.Cmd {
	m.focus = true
	return nil
}

func (m *Model) Blur() {
	m.focus = false
}

func (m *Model) Focused() bool {
	return m.focus
}
