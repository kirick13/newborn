package view

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	card "github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
	"github.com/kirick13/newborn/state"
)

type HostsView struct {
	BaseView
}

var (
	hostsTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f5f5f5"))

	hostsMutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#737373"))

	hostsHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d4d4d4"))

	hostsCellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e5e5e5"))

	hostsDividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#525252"))
)

func NewHostsView() *HostsView {
	return &HostsView{
		BaseView: BaseView{},
	}
}

func (v *HostsView) Render() string {
	content := []string{
		hostsTitleStyle.Render("Hosts list"),
		"",
		v.renderBody(),
		"",
		keys.RenderKeys([]keys.Keys{
			{Key: "a", Title: "add"},
			{Key: "e", Title: "edit"},
			{Key: "backspace", Title: "delete"},
			{Key: "enter", Title: "next"},
		}),
	}

	return card.New().
		Padding(1, 2).
		Width(64).
		Render(strings.Join(content, "\n"))
}

func (v *HostsView) renderBody() string {
	if v.Display == nil || v.Display.State() == nil {
		return hostsMutedStyle.Render("loading state...")
	}

	hosts := v.Display.State().Hosts
	if len(hosts) == 0 {
		return hostsMutedStyle.Render("no hosts added yet. press a to add new host")
	}

	nameWidth := len("Name")
	hostWidth := len("Host")
	for _, host := range hosts {
		nameWidth = max(nameWidth, lipgloss.Width(host.Setup.Name))
		hostWidth = max(hostWidth, lipgloss.Width(v.hostLabel(host)))
	}

	rows := []string{
		hostsHeaderStyle.Render(padRight("Name", nameWidth)) + "  " +
			hostsHeaderStyle.Render(padRight("Host", hostWidth)),
		hostsDividerStyle.Render(strings.Repeat("─", nameWidth)) + "  " +
			hostsDividerStyle.Render(strings.Repeat("─", hostWidth)),
	}

	for _, host := range hosts {
		rows = append(
			rows,
			hostsCellStyle.Render(padRight(host.Setup.Name, nameWidth))+"  "+
				hostsCellStyle.Render(padRight(v.hostLabel(host), hostWidth)),
		)
	}

	return strings.Join(rows, "\n")
}

func (v *HostsView) hostLabel(host state.Host) string {
	return fmt.Sprintf("%s:%d", host.Connect.IP, host.Connect.SSHPort)
}

func padRight(value string, width int) string {
	padding := max(width-lipgloss.Width(value), 0)
	return value + strings.Repeat(" ", padding)
}
