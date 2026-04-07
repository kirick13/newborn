package view

import (
	"encoding/json"
	"strings"

	tea "charm.land/bubbletea/v2"
	card "github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
	"github.com/kirick13/newborn/style"
)

type SoftwareView struct {
	BaseView
	previous *SettingsView
}

func NewSoftwareView(previous *SettingsView) *SoftwareView {
	return &SoftwareView{
		BaseView: BaseView{},
		previous: previous,
	}
}

func (v *SoftwareView) Render() string {
	content := []string{
		hostsTitleStyle.Render("Software"),
		"",
		style.FormLabelSubStyle.Render("software form is empty for now"),
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

func (v *SoftwareView) OnEnter() tea.Cmd {
	if v.Display == nil || v.Display.State() == nil {
		return nil
	}

	v.Display.SetCurrentView(NewTextView(v, "Debug", buildDockerRunCommand(v)))
	return nil
}

func (v *SoftwareView) OnEsc() tea.Cmd {
	if v.Display != nil && v.previous != nil {
		v.Display.SetCurrentView(v.previous)
	}
	return nil
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
		"local/newborn",
		"-e",
		shellQuoteCommand(string(jsonPayload)),
	}

	return strings.Join(args, " ")
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
