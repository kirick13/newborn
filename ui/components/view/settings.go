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
	"github.com/kirick13/newborn/style"
)

type SettingsView struct {
	BaseView
	previous  *HostsView
	firewallRadio *radio_group.RadioGroup
	// editIndex int
	// errorText string
	// statusText string
	// checking  bool
	// spinner   spinner.Model
	// pending   *state.Host
	// title     string
}

// previous *HostsView
func NewSettingsView() *SettingsView {
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
		BaseView:  BaseView{},
		firewallRadio: firewallRadio,
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
		// "",
		// v.renderField("path to ssh identity file", 4),
		"",
		keys.RenderKeys([]keys.Keys{
			{Key: "esc", Title: "back"},
			{Key: "enter", Title: "next"},
		}),
	}

	// if v.statusText != "" {
	// 	content = append(content, "", v.statusView())
	// } else if v.errorText != "" {
	// 	content = append(content, "", formErrorStyle.Render(v.errorText))
	// }

	// if !v.checking {
	// 	content = append(content, "", keys.RenderKeys([]keys.Keys{
	// 		{Key: "enter", Title: "save"},
	// 		{Key: "esc", Title: "cancel"},
	// 	}))
	// }

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
