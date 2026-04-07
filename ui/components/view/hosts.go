package view

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	card "github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
	"github.com/kirick13/newborn/state"
)

type HostsView struct {
	BaseView
	table table.Model
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
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 24},
			{Title: "Host", Width: 28},
		}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(8),
		table.WithWidth(56),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return &HostsView{
		BaseView: BaseView{},
		table:    t,
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

	v.syncTable()
	return v.table.View()
}

func (v *HostsView) hostLabel(host state.Host) string {
	return fmt.Sprintf("%s:%d", host.Connect.IP, host.Connect.SSHPort)
}

func (v *HostsView) OnKey(key string) {
	if v.Display == nil || v.Display.State() == nil {
		return
	}

	switch key {
	case "a":
		v.Display.State().AddRandomHost()
		v.syncTable()
	case "backspace":
		v.confirmDelete()
	}
}

func (v *HostsView) OnMsg(msg tea.Msg) tea.Cmd {
	if v.Display == nil || v.Display.State() == nil || len(v.Display.State().Hosts) == 0 {
		return nil
	}

	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return cmd
}

func (v *HostsView) syncTable() {
	if v.Display == nil || v.Display.State() == nil {
		return
	}

	hosts := v.Display.State().Hosts
	rows := make([]table.Row, 0, len(hosts))
	for _, host := range hosts {
		rows = append(rows, table.Row{
			host.Setup.Name,
			v.hostLabel(host),
		})
	}

	v.table.SetRows(rows)
}

func (v *HostsView) confirmDelete() {
	hosts := v.Display.State().Hosts
	if len(hosts) == 0 {
		return
	}

	index := v.table.Cursor()
	if index < 0 || index >= len(hosts) {
		return
	}

	host := hosts[index]
	v.Display.SetCurrentView(NewConfirmView(
		"Delete host",
		fmt.Sprintf("Delete host %q (%s)?", host.Setup.Name, v.hostLabel(host)),
		v,
		func() {
			if v.Display == nil || v.Display.State() == nil {
				return
			}

			if !v.Display.State().DeleteHost(index) {
				return
			}

			v.syncTable()
			if len(v.Display.State().Hosts) == 0 {
				return
			}

			v.table.SetCursor(min(index, len(v.Display.State().Hosts)-1))
		},
	))
}
