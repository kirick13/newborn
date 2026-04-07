package view

import (
	"encoding/json"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
	"github.com/kirick13/newborn/components/radio_group"
	"github.com/kirick13/newborn/elements"
	"github.com/kirick13/newborn/elements/checkbox"
	"github.com/kirick13/newborn/state"
	"github.com/kirick13/newborn/style"
)

type SoftwareView struct {
	BaseView
	previous                *SettingsView
	containerRuntimeRadio *radio_group.RadioGroup
	composeCheckbox       *checkbox.Model
	kubernetesRuntimeRadio *radio_group.RadioGroup
	removeSnapCheckbox    *checkbox.Model
	errorText             string
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
	v.Display.SetCurrentView(NewTextView(v, "Debug", buildDebugText(v)))
	return nil
}

func (v *SoftwareView) OnEsc() tea.Cmd {
	if v.Display != nil && v.previous != nil {
		v.Display.SetCurrentView(v.previous)
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

func buildDockerRunCommand(v *SoftwareView) string {
	payload := struct {
		NewbornSwap        string `json:"newborn_swap"`
		NewbornReserveFile string `json:"newborn_reserve_file"`
		NewbornFirewall    string `json:"newborn_firewall_http"`
		NewbornOCIRuntime  string `json:"newborn_oci_runtime"`
		NewbornOCICompose  string `json:"newborn_oci_compose"`
		NewbornK8sRuntime  string `json:"newborn_k8s_runtime"`
		NewbornRemoveSnap  string `json:"newborn_remove_snap"`
	}{
		NewbornSwap:        v.Display.State().Setup.Swap,
		NewbornReserveFile: boolToFlag(v.Display.State().Setup.ReserveFile),
		NewbornFirewall:    v.Display.State().Setup.FirewallHTTP,
		NewbornOCIRuntime:  v.Display.State().Software.OCIRuntime,
		NewbornOCICompose:  boolToFlag(v.Display.State().Software.OCICompose),
		NewbornK8sRuntime:  v.Display.State().Software.K8sRuntime,
		NewbornRemoveSnap:  boolToFlag(v.Display.State().Software.RemoveSnap),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "docker run -t --rm local/newborn -e '<could not build payload>'"
	}

	args := []string{
		"docker",
		"run",
		"-t",
		"--rm",
	}

	if inventoryPath := strings.TrimSpace(v.Display.State().InventoryPath); inventoryPath != "" {
		args = append(args, "-v", shellQuoteCommand(inventoryPath+":/app/inventory.yaml:ro"))
	}

	for _, host := range v.Display.State().Hosts {
		if strings.TrimSpace(host.Connect.SSHKeyPath) == "" {
			continue
		}

		args = append(args,
			"-v",
			shellQuoteCommand(host.Connect.SSHKeyPath+":/opt/bind/ssh/"+host.Setup.Name+".key:ro"),
		)
	}

	args = append(args,
		"local/newborn",
		"-e",
		shellQuoteCommand(string(jsonPayload)),
	)

	return strings.Join(args, " ")
}

func buildDebugText(v *SoftwareView) string {
	parts := []string{}

	if content := strings.TrimSpace(v.Display.State().InventoryContent); content != "" {
		parts = append(parts, content)
	}

	parts = append(parts, buildDockerRunCommand(v))
	return strings.Join(parts, "\n\n")
}

func boolToFlag(value bool) string {
	if value {
		return "y"
	}

	return ""
}

func shellQuoteCommand(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
