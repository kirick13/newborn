package view

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	card "github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
	"github.com/kirick13/newborn/style"
)

type TextView struct {
	BaseView
	previous View
	title    string
	text     string
}

func NewTextView(previous View, title, text string) *TextView {
	return &TextView{
		BaseView: BaseView{},
		previous: previous,
		title:    title,
		text:     text,
	}
}

func (v *TextView) Render() string {
	content := []string{
		hostsTitleStyle.Render(v.title),
		"",
		style.FormLabelStyle.Render(v.text),
		"",
		keys.RenderKeys([]keys.Keys{
			{Key: "esc", Title: "back"},
		}),
	}

	return card.New().
		Padding(1, 2).
		Width(80).
		Render(strings.Join(content, "\n"))
}

func (v *TextView) OnEnter() tea.Cmd {
	return nil
}

func (v *TextView) OnEsc() tea.Cmd {
	if v.Display != nil && v.previous != nil {
		v.Display.SetCurrentView(v.previous)
	}
	return nil
}
