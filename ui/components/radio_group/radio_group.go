package radio_group

import (
	"strings"

	"charm.land/lipgloss/v2"
	checkbox "github.com/kirick13/newborn/elements/checkbox"
	input "github.com/kirick13/newborn/elements"
	"github.com/kirick13/newborn/style"
)

type Option struct {
	Title string
	Value string
}

type RadioGroup struct {
	Options    []Option
	Inputs     []input.Element
	value      string
	checkboxes []*checkbox.Model
	lastValues []bool
}

func New(options []Option) *RadioGroup {
	g := &RadioGroup{
		Options:    append([]Option(nil), options...),
		checkboxes: make([]*checkbox.Model, len(options)),
		Inputs:     make([]input.Element, len(options)),
		lastValues: make([]bool, len(options)),
		value:      "",
	}

	for i := range options {
		c := checkbox.New()
		c.Symbol = "o"
		g.checkboxes[i] = c
		g.Inputs[i] = c
	}

	return g
}

func (g *RadioGroup) Render() string {
	g.sync()

	parts := make([]string, 0, len(g.checkboxes))
	for index, c := range g.checkboxes {
		parts = append(parts, lipgloss.JoinHorizontal(
			lipgloss.Left,
			c.Render(),
			" ",
			style.FormLabelStyle.Render(g.Options[index].Title),
		))
	}

	return strings.Join(parts, "  ")
}

func (g *RadioGroup) Value() string {
	g.sync()
	return g.value
}

func (g *RadioGroup) SetValue(value string) {
	if g == nil {
		return
	}

	selectedIndex := -1
	for i, option := range g.Options {
		if option.Value == value {
			selectedIndex = i
			break
		}
	}

	for i, box := range g.checkboxes {
		box.Value = i == selectedIndex && selectedIndex != -1
		g.lastValues[i] = box.Value
	}

	if selectedIndex == -1 {
		g.value = ""
		return
	}

	g.value = g.Options[selectedIndex].Value
}

func (g *RadioGroup) HasValue(value string) bool {
	for _, option := range g.Options {
		if option.Value == value {
			return true
		}
	}

	return false
}

func (g *RadioGroup) sync() {
	if g == nil {
		return
	}

	changedIndex := -1
	for i, box := range g.checkboxes {
		if box.Value != g.lastValues[i] {
			changedIndex = i
			break
		}
	}

	if changedIndex >= 0 {
		if g.checkboxes[changedIndex].Value {
			for i, box := range g.checkboxes {
				box.Value = i == changedIndex
				g.lastValues[i] = box.Value
			}
			g.value = g.Options[changedIndex].Value
			return
		}

		for i, box := range g.checkboxes {
			if i == changedIndex {
				box.Value = false
			}
			g.lastValues[i] = box.Value
		}
		g.value = ""
		return
	}

	selectedIndex := -1
	for i, box := range g.checkboxes {
		if box.Value {
			selectedIndex = i
			break
		}
	}

	if selectedIndex == -1 {
		g.value = ""
	} else {
		g.value = g.Options[selectedIndex].Value
	}

	for i, box := range g.checkboxes {
		box.Value = i == selectedIndex && selectedIndex != -1
		g.lastValues[i] = box.Value
	}
}
