package view

import (
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	card "github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
	"github.com/kirick13/newborn/elements"
	"github.com/kirick13/newborn/elements/input"
	"github.com/kirick13/newborn/provision"
	"github.com/kirick13/newborn/state"
	"github.com/kirick13/newborn/style"
)

type BwPasswordView struct {
	BaseView
	errorText  string
	statusText string
	checking   bool
	spinner    spinner.Model
}

type bwFinishedMsg struct {
	err error
}

func NewBwPasswordView() *BwPasswordView {
	passwordInput := input.New("")
	passwordInput.EchoMode = textinput.EchoPassword

	v := &BwPasswordView{
		BaseView: BaseView{},
		spinner:  spinner.New(),
	}
	v.spinner.Spinner = spinner.Dot
	v.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa"))
	v.SetInputs([]elements.Element{passwordInput})

	return v
}

func (v *BwPasswordView) Render() string {
	baseInputs := v.BaseView.Inputs()
	passwordField := ""
	if len(baseInputs) > 0 {
		passwordField = baseInputs[0].Render()
	}

	content := []string{
		hostsTitleStyle.Render("Unlock Bitwarden"),
		"",
		style.FormLabelStyle.Render("Enter your Bitwarden master password to save hosts."),
		"",
		passwordField,
	}

	if v.statusText != "" {
		content = append(content, "", lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d4d4d4")).
			Render(v.spinner.View()+" "+v.statusText))
	} else if v.errorText != "" {
		content = append(content, "", style.FormErrorStyle.Render(v.errorText))
	}

	if !v.checking {
		content = append(content, "",
			keys.RenderKeys([]keys.Keys{
				{Key: "enter", Title: "unlock"},
			}),
		)
	}

	return card.New().
		Padding(1, 2).
		Width(64).
		Render(strings.Join(content, "\n"))
}

func (v *BwPasswordView) Inputs() []elements.Element {
	if v.checking {
		return nil
	}
	return v.BaseView.Inputs()
}

func (v *BwPasswordView) OnEnter() tea.Cmd {
	if v.Display == nil || v.Display.State() == nil || v.checking {
		return nil
	}

	passwordInput, ok := v.BaseView.Inputs()[0].(*input.Model)
	if !ok {
		return nil
	}

	password := passwordInput.Value()
	if strings.TrimSpace(password) == "" {
		v.errorText = "bitwarden password is required"
		return nil
	}

	v.errorText = ""
	v.checking = true
	v.statusText = "unlocking Bitwarden..."
	return tea.Batch(v.spinner.Tick, runBitwardenSync(password, v.Display.State().Hosts))
}

func (v *BwPasswordView) OnMsg(msg tea.Msg) tea.Cmd {
	if !v.checking {
		return nil
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return cmd
	case bwFinishedMsg:
		v.checking = false
		v.statusText = ""
		if msg.err != nil {
			v.errorText = msg.err.Error()
			return nil
		}

		provision.ClearTerminal()
		return tea.Quit
	}

	return nil
}

func runBitwardenSync(password string, hosts []state.Host) tea.Cmd {
	return func() tea.Msg {
		session, err := provision.UnlockBitwarden(password)
		if err != nil {
			return bwFinishedMsg{err: err}
		}

		if err := provision.CreateSSHKeyItems(session, hosts); err != nil {
			return bwFinishedMsg{err: err}
		}

		return bwFinishedMsg{}
	}
}
