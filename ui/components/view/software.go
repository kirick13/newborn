package view

import (
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
	"github.com/kirick13/newborn/components/radio_group"
	"github.com/kirick13/newborn/elements"
	"github.com/kirick13/newborn/elements/checkbox"
	"github.com/kirick13/newborn/provision"
	"github.com/kirick13/newborn/state"
	"github.com/kirick13/newborn/style"
)

type SoftwareView struct {
	BaseView
	previous                *SettingsView
	containerRuntimeRadio  *radio_group.RadioGroup
	composeCheckbox        *checkbox.Model
	kubernetesRuntimeRadio *radio_group.RadioGroup
	removeSnapCheckbox     *checkbox.Model
	errorText              string
}

type dockerFinishedMsg struct {
	err error
}

func NewSoftwareView(previous *SettingsView) *SoftwareView {
	v := &SoftwareView{
		BaseView: BaseView{},
		previous: previous,
	}

	inputs := []elements.Element{}

	v.containerRuntimeRadio = radio_group.New([]radio_group.Option{
		{Title: "Docker", Value: "docker"},
		{Title: "Podman", Value: "podman"},
	})
	inputs = append(inputs, v.containerRuntimeRadio.Inputs...)

	v.composeCheckbox = checkbox.New()
	inputs = append(inputs, v.composeCheckbox)

	v.kubernetesRuntimeRadio = radio_group.New([]radio_group.Option{
		{Title: "k0s", Value: "k0s"},
		{Title: "Microk8s", Value: "microk8s"},
	})
	inputs = append(inputs, v.kubernetesRuntimeRadio.Inputs...)

	v.removeSnapCheckbox = checkbox.New()
	inputs = append(inputs, v.removeSnapCheckbox)

	if previous != nil && previous.Display != nil && previous.Display.State() != nil {
		software := previous.Display.State().Software
		if v.containerRuntimeRadio.HasValue(software.OCIRuntime) {
			v.containerRuntimeRadio.SetValue(software.OCIRuntime)
		}
		v.composeCheckbox.Value = software.OCICompose
		if v.kubernetesRuntimeRadio.HasValue(software.K8sRuntime) {
			v.kubernetesRuntimeRadio.SetValue(software.K8sRuntime)
		}
		v.removeSnapCheckbox.Value = software.RemoveSnap
	}

	v.SetInputs(inputs)

	return v
}

func (v *SoftwareView) Render() string {
	content := []string{
		hostsTitleStyle.Render("Software"),
		"",
		style.FormLabelStyle.Render("container runtime"),
		v.containerRuntimeRadio.Render(),
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			v.composeCheckbox.Render(),
			" ",
			style.FormLabelStyle.Render("install compose"),
		),
		"",
		style.FormLabelStyle.Render("kubernetes runtime"),
		v.kubernetesRuntimeRadio.Render(),
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			v.removeSnapCheckbox.Render(),
			" ",
			style.FormLabelStyle.Render("remove snap"),
		),
	}

	if v.errorText != "" {
		content = append(content, "", style.FormErrorStyle.Render(v.errorText))
	}

	content = append(content, "",
		keys.RenderKeys([]keys.Keys{
			{Key: "esc", Title: "back"},
			{Key: "enter", Title: "next"},
		}),
	)

	return card.New().
		Padding(1, 2).
		Width(64).
		Render(strings.Join(content, "\n"))
}

func (v *SoftwareView) OnEnter() tea.Cmd {
	if v.Display == nil || v.Display.State() == nil {
		return nil
	}

	software, err := v.softwareState()
	if err != nil {
		v.errorText = err.Error()
		return nil
	}

	v.Display.State().Software = software
	v.errorText = ""
	return tea.ExecProcess(
		buildProvisionCommand(v.Display.State()),
		func(err error) tea.Msg {
			return dockerFinishedMsg{err: err}
		},
	)
}

func (v *SoftwareView) OnEsc() tea.Cmd {
	if v.Display != nil && v.previous != nil {
		v.Display.SetCurrentView(v.previous)
	}
	return nil
}

func (v *SoftwareView) OnMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case dockerFinishedMsg:
		if msg.err != nil {
			return func() tea.Msg {
				return QuitNowMsg{}
			}
		}

		provision.ClearTerminal()
		if v.Display != nil {
			v.Display.SetCurrentView(NewBwPasswordView())
		}
		return tea.ClearScreen
	}

	return nil
}

func (v *SoftwareView) softwareState() (state.SoftwareOptions, error) {
	software := state.SoftwareOptions{
		OCIRuntime: v.containerRuntimeRadio.Value(),
		OCICompose: v.composeCheckbox.Value,
		K8sRuntime: v.kubernetesRuntimeRadio.Value(),
		RemoveSnap: v.removeSnapCheckbox.Value,
	}

	if software.OCICompose && software.OCIRuntime == "" {
		return state.SoftwareOptions{}, errSoftware("compose can not be installed without container runtime")
	}

	if software.RemoveSnap && software.K8sRuntime == "microk8s" {
		return state.SoftwareOptions{}, errSoftware("microk8s requires snap, so it can not be removed")
	}

	return software, nil
}

type errSoftware string

func (e errSoftware) Error() string {
	return string(e)
}

func buildProvisionCommand(app *state.Newborn) *exec.Cmd {
	cmd := exec.Command("sh", "-lc", provision.BuildProvisionShellScript(app))
	return cmd
}
