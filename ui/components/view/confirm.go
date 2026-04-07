package view

import (
	"strings"

	"charm.land/lipgloss/v2"
	card "github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
)

type ConfirmView struct {
	BaseView
	title     string
	question  string
	previous  View
	onConfirm func()
}

var (
	confirmTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#f5f5f5"))
	confirmBodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d4d4d4"))
)

func NewConfirmView(title, question string, previous View, onConfirm func()) *ConfirmView {
	return &ConfirmView{
		title:     title,
		question:  question,
		previous:  previous,
		onConfirm: onConfirm,
	}
}

func (v *ConfirmView) Render() string {
	content := []string{
		confirmTitleStyle.Render(v.title),
		"",
		confirmBodyStyle.Render(v.question),
		"",
		keys.RenderKeys([]keys.Keys{
			{Key: "enter", Title: "confirm"},
			{Key: "esc", Title: "cancel"},
		}),
	}

	return card.New().
		Padding(1, 2).
		Width(64).
		Render(strings.Join(content, "\n"))
}

func (v *ConfirmView) OnEnter() {
	if v.onConfirm != nil {
		v.onConfirm()
	}

	if v.Display != nil && v.previous != nil {
		v.Display.SetCurrentView(v.previous)
	}
}

func (v *ConfirmView) OnEsc() {
	if v.Display != nil && v.previous != nil {
		v.Display.SetCurrentView(v.previous)
	}
}
