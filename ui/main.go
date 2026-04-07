package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/kirick13/newborn/components"
	view "github.com/kirick13/newborn/components/view"
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	display           *components.Display
	quitting          bool
	// clearOnNextRender bool
}

func initialModel() model {
	return model{
		display:           components.NewDisplay(),
		// clearOnNextRender: true,
	}
}

func (m model) Init() tea.Cmd {
	if spinnerView, ok := m.display.CurrentView.(*view.SpinnerView); ok {
		return spinnerView.Init()
	}

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var viewCmd tea.Cmd
	var keyCmd tea.Cmd
	initialView := m.display.CurrentView
	if m.display != nil && m.display.CurrentView != nil {
		viewCmd = m.display.CurrentView.OnMsg(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.display.UpdateDocumentSize(msg.Width, msg.Height)
		return m, viewCmd
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			m.display.MoveFocus(1)
			return m, nil
		case "shift+tab":
			m.display.MoveFocus(-1)
			return m, nil
		case "enter":
			keyCmd = m.display.CurrentView.OnEnter()
		case "esc":
			keyCmd = m.display.CurrentView.OnEsc()
		default:
			keyCmd = m.display.CurrentView.OnKey(msg.String())
		}
	}

	if m.display == nil || m.display.CurrentView == nil {
		return m, viewCmd
	}

	if initialView != nil && m.display.CurrentView != initialView {
		return m, tea.Batch(viewCmd, keyCmd)
	}

	inputs := m.display.CurrentView.Inputs()
	cmds := make([]tea.Cmd, len(inputs))
	for i := range inputs {
		inputs[i], cmds[i] = inputs[i].Update(msg)
	}

	cmds = append(cmds, viewCmd, keyCmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	content := m.display.Render()
	// if m.clearOnNextRender {
	// 	content = clearTerminalSequence + content
	// }

	return tea.NewView(content)
}
