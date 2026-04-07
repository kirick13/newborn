package view

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	card "github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
	"github.com/kirick13/newborn/components/radio_group"
	input "github.com/kirick13/newborn/elements"
	checkbox "github.com/kirick13/newborn/elements/checkbox"
	"github.com/kirick13/newborn/state"
	"github.com/kirick13/newborn/style"
)

type SettingsView struct {
	BaseView
	previous      *HostsView
	firewallRadio *radio_group.RadioGroup
}

func NewSettingsView(previous *HostsView) *SettingsView {
	inputs := []input.Element{}

	swapInput := input.New("100M, 2G etc.")
	inputs = append(inputs, swapInput)

	diskReserveCheckbox := checkbox.New()
	inputs = append(inputs, diskReserveCheckbox)

	firewallRadio := radio_group.New([]radio_group.Option{
		{Title: "Cloudflare", Value: "cloudflare"},
		{Title: "Anywhere", Value: "anywhere"},
	})
	inputs = append(inputs, firewallRadio.Inputs...)

	v := &SettingsView{
		BaseView:      BaseView{},
		previous:      previous,
		firewallRadio: firewallRadio,
	}

	if previous != nil && previous.Display != nil && previous.Display.State() != nil {
		setup := previous.Display.State().Setup
		swapInput.SetValue(setup.Swap)
		diskReserveCheckbox.Value = setup.ReserveFile
		if firewallRadio.HasValue(setup.FirewallHTTP) {
			firewallRadio.SetValue(setup.FirewallHTTP)
		}
	}

	v.SetInputs(inputs)

	return v
}

func (v *SettingsView) Render() string {
	content := []string{
		hostsTitleStyle.Render("Settings"),
		"",
		v.renderField(
			fmt.Sprintf("swap size  %s", style.FormLabelSubStyle.Render("leave empty to turn swap off")),
			0,
		),
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			v.inputs[1].Render(),
			" ",
			fmt.Sprintf(
				"%s\n%s",
				style.FormLabelStyle.Render("add 2G disk reserve file"),
				style.FormLabelSubStyle.Render("delete it to free space in emergency"),
			),
		),
		"",
		style.FormLabelStyle.Render("allow HTTP(S) traffic from"),
		v.firewallRadio.Render(),
		"",
		keys.RenderKeys([]keys.Keys{
			{Key: "esc", Title: "back"},
			{Key: "enter", Title: "next"},
		}),
	}

	return card.New().
		Padding(1, 2).
		Width(64).
		Render(strings.Join(content, "\n"))
}

func (v *SettingsView) renderField(label string, index int) string {
	inputs := v.Inputs()
	if index < 0 || index >= len(inputs) {
		return ""
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.FormLabelStyle.Render(label),
		inputs[index].Render(),
	)
}

func (v *SettingsView) OnEnter() tea.Cmd {
	if v.Display == nil || v.Display.State() == nil {
		return nil
	}

	v.Display.State().Setup = state.SetupOptions{
		Swap:         v.swapValue(),
		ReserveFile:  v.reserveFileValue(),
		FirewallHTTP: v.firewallValue(),
	}
	v.Display.SetCurrentView(NewSoftwareView(v))
	return nil
}

func (v *SettingsView) OnEsc() tea.Cmd {
	if v.Display != nil && v.previous != nil {
		v.Display.SetCurrentView(v.previous)
	}
	return nil
}

func (v *SettingsView) OnMsg(msg tea.Msg) tea.Cmd {
	return nil
}

func (v *SettingsView) swapValue() string {
	if len(v.inputs) == 0 {
		return ""
	}

	swapInput, ok := v.inputs[0].(*input.Model)
	if !ok {
		return ""
	}

	return strings.TrimSpace(swapInput.Value())
}

func (v *SettingsView) reserveFileValue() bool {
	if len(v.inputs) < 2 {
		return false
	}

	box, ok := v.inputs[1].(*checkbox.Model)
	if !ok {
		return false
	}

	return box.Value
}

func (v *SettingsView) firewallValue() string {
	value := strings.TrimSpace(v.firewallRadio.Value())
	if value == "" {
		return "nowhere"
	}

	return value
}
