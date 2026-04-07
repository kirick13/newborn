package elements

import tea "charm.land/bubbletea/v2"

type Element interface {
	Render() string
	Focus() tea.Cmd
	Blur()
	Focused() bool
	Update(tea.Msg) (Element, tea.Cmd)
}
