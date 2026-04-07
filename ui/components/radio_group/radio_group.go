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
	Value      string
	checkboxes []*checkbox.Model
	lastValues []bool
}

func New(options []Option) *RadioGroup {
	g := &RadioGroup{
		Options:    append([]Option(nil), options...),
		checkboxes: make([]*checkbox.Model, len(options)),
		Inputs:     make([]input.Element, len(options)),
		lastValues: make([]bool, len(options)),
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

func (g *RadioGroup) CheckedValue() string {
	g.sync()
	return g.Value
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
			g.Value = g.Options[changedIndex].Value
			return
		}

		for i, box := range g.checkboxes {
			if i == changedIndex {
				box.Value = false
			}
			g.lastValues[i] = box.Value
		}
		g.Value = ""
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
		g.Value = ""
	} else {
		g.Value = g.Options[selectedIndex].Value
	}

	for i, box := range g.checkboxes {
		box.Value = i == selectedIndex && selectedIndex != -1
		g.lastValues[i] = box.Value
	}
}
