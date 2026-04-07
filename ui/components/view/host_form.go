package view

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	card "github.com/kirick13/newborn/components/card"
	"github.com/kirick13/newborn/components/keys"
	"github.com/kirick13/newborn/config"
	"github.com/kirick13/newborn/elements"
	"github.com/kirick13/newborn/elements/input"
	"github.com/kirick13/newborn/state"
	"github.com/kirick13/newborn/style"
	"golang.org/x/crypto/ssh"
)

type HostFormView struct {
	BaseView
	previous  *HostsView
	editIndex int
	errorText string
	statusText string
	checking  bool
	spinner   spinner.Model
	pending   *state.Host
	title     string
}

type sshCheckResultMsg struct {
	host state.Host
	err  error
}

const rootSSHUser = "root"

func NewHostFormView(previous *HostsView, editIndex int) *HostFormView {
	v := &HostFormView{
		BaseView:  BaseView{},
		previous:  previous,
		editIndex: editIndex,
		title:     "New host",
	}
	v.spinner = spinner.New()
	v.spinner.Spinner = spinner.Dot
	v.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa"))

	if editIndex >= 0 {
		v.title = "Edit host"
	}

	nameInput := input.New("")

	ipInput := input.New("")
	ipInput.SetWidth(32)

	portInput := input.New("")
	portInput.CharLimit = 5
	portInput.SetWidth(10)
	portInput.SetValue("22")

	passwordInput := input.New("")
	passwordInput.EchoMode = textinput.EchoPassword

	identityInput := input.New("")

	if previous != nil && previous.Display != nil && previous.Display.State() != nil &&
		editIndex >= 0 && editIndex < len(previous.Display.State().Hosts) {
		host := previous.Display.State().Hosts[editIndex]
		nameInput.SetValue(host.Setup.Name)
		ipInput.SetValue(host.Connect.IP)
		portInput.SetValue(strconv.Itoa(host.Connect.SSHPort))
		passwordInput.SetValue(host.Connect.Password)
		identityInput.SetValue(host.Connect.SSHKeyPath)
	} else {
		defaults := config.LoadDefaults()
		nameInput.SetValue(defaults.Name)
		ipInput.SetValue(defaults.IP)
		identityInput.SetValue(defaults.SSHKeyPath)
	}

	v.SetInputs([]elements.Element{
		nameInput,
		ipInput,
		portInput,
		passwordInput,
		identityInput,
	})
	return v
}

func (v *HostFormView) Render() string {
	content := []string{
		hostsTitleStyle.Render(v.title),
		"",
		v.renderField("host display name", 0),
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			v.renderField("IP", 1),
			"  ",
			v.renderField("port", 2),
		),
		"",
		v.renderField("root password", 3),
		"",
		v.renderField("path to ssh identity file", 4),
	}

	if v.statusText != "" {
		content = append(content, "", v.statusView())
	} else if v.errorText != "" {
		content = append(content, "", style.FormErrorStyle.Render(v.errorText))
	}

	if !v.checking {
		content = append(content, "", keys.RenderKeys([]keys.Keys{
			{Key: "enter", Title: "save"},
			{Key: "esc", Title: "cancel"},
		}))
	}

	return card.New().
		Padding(1, 2).
		Width(64).
		Render(strings.Join(content, "\n"))
}

func (v *HostFormView) renderField(label string, index int) string {
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

func (v *HostFormView) OnEnter() tea.Cmd {
	if v.Display == nil || v.Display.State() == nil {
		return nil
	}
	if v.checking {
		return nil
	}

	host, err := v.validateAndBuildHost()
	if err != nil {
		v.statusText = ""
		v.errorText = err.Error()
		return nil
	}

	v.pending = &host
	v.checking = true
	v.errorText = ""
	v.statusText = "checking SSH connection..."
	return tea.Batch(v.spinner.Tick, runSSHCheck(host))
}

func (v *HostFormView) OnEsc() tea.Cmd {
	if v.checking {
		return nil
	}
	if v.Display != nil && v.previous != nil {
		v.Display.SetCurrentView(v.previous)
	}
	return nil
}

func (v *HostFormView) validateAndBuildHost() (state.Host, error) {
	values := v.values()
	displayName := strings.TrimSpace(values[0])
	ip := strings.TrimSpace(values[1])
	portText := strings.TrimSpace(values[2])
	password := values[3]
	identityPath := strings.TrimSpace(values[4])

	if displayName == "" {
		return state.Host{}, errors.New("host display name is required")
	}

	if ip == "" {
		return state.Host{}, errors.New("IP is required")
	}

	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return state.Host{}, errors.New("port must be a number between 1 and 65535")
	}

	if password != "" && identityPath != "" {
		return state.Host{}, errors.New("use either password or identity file, not both")
	}

	if password == "" && identityPath == "" {
		return state.Host{}, errors.New("password or identity file is required")
	}

	resolvedIdentity := ""
	if identityPath != "" {
		resolvedIdentity = resolvePath(identityPath)
		if _, err := os.Stat(resolvedIdentity); err != nil {
			if os.IsNotExist(err) {
				return state.Host{}, errors.New("identity file does not exist")
			}
			return state.Host{}, fmt.Errorf("could not read identity file: %w", err)
		}
	}

	host := state.Host{
		Connect: state.ConnectOptions{
			IP:         ip,
			Password:   password,
			SSHPort:    port,
			SSHKeyPath: resolvedIdentity,
		},
	}

	if v.editIndex >= 0 && v.previous != nil && v.Display != nil && v.Display.State() != nil &&
		v.editIndex < len(v.Display.State().Hosts) {
		host.Setup = v.Display.State().Hosts[v.editIndex].Setup
	} else {
		setup, err := generateHostSetup(displayName)
		if err != nil {
			return state.Host{}, err
		}
		host.Setup = setup
	}

	host.Setup.Name = displayName
	return host, nil
}

func (v *HostFormView) OnMsg(msg tea.Msg) tea.Cmd {
	if v.checking {
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			return cmd
		case sshCheckResultMsg:
			v.checking = false
			v.statusText = ""
			if msg.err != nil {
				v.pending = nil
				v.errorText = msg.err.Error()
				return nil
			}

			if v.Display == nil || v.Display.State() == nil {
				return nil
			}

			index := v.Display.State().UpsertHost(v.editIndex, msg.host)
			v.pending = nil
			if v.previous != nil {
				v.previous.syncTable()
				v.previous.table.SetCursor(index)
			}
			v.Display.SetCurrentView(v.previous)
		}
	}

	return nil
}

func (v *HostFormView) Inputs() []elements.Element {
	if v.checking {
		return nil
	}

	return v.BaseView.Inputs()
}

func (v *HostFormView) statusView() string {
	if !v.checking {
		return style.FormErrorStyle.Render(v.statusText)
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d4d4d4")).
		Render(v.spinner.View() + " " + v.statusText)
}

func (v *HostFormView) values() []string {
	inputs := v.Inputs()
	values := make([]string, len(inputs))
	for i := range inputs {
		textInput, ok := inputs[i].(*input.Model)
		if !ok {
			continue
		}
		values[i] = textInput.Value()
	}
	return values
}

func resolvePath(value string) string {
	if value == "" {
		return ""
	}

	if strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			value = filepath.Join(home, value[2:])
		}
	}

	return filepath.Clean(value)
}

func verifySSH(ip string, port int, password, identityPath string) error {
	if _, err := exec.LookPath("ssh"); err != nil {
		return errors.New("ssh command is required to validate host connection")
	}

	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=10",
		"-o", "NumberOfPasswordPrompts=1",
		"-p", strconv.Itoa(port),
	}

	if identityPath != "" {
		args = append(args,
			"-o", "BatchMode=yes",
			"-i", identityPath,
		)
	} else {
		args = append(args,
			"-o", "PreferredAuthentications=password",
			"-o", "PubkeyAuthentication=no",
			"-o", "KbdInteractiveAuthentication=no",
		)
	}

	args = append(args, fmt.Sprintf("%s@%s", rootSSHUser, ip), "true")

	cmd := exec.Command("ssh", args...)
	if password != "" {
		scriptPath, cleanup, err := createAskpassScript(password)
		if err != nil {
			return fmt.Errorf("could not prepare password auth check: %w", err)
		}
		defer cleanup()

		cmd.Env = append(os.Environ(),
			"DISPLAY=codex",
			"SSH_ASKPASS="+scriptPath,
			"SSH_ASKPASS_REQUIRE=force",
		)
	}

	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	text := strings.TrimSpace(string(output))
	lower := strings.ToLower(text)
	if strings.Contains(lower, "change your password") ||
		strings.Contains(lower, "password has expired") ||
		strings.Contains(lower, "password expired") ||
		strings.Contains(lower, "administrator enforced") {
		return errors.New("ssh login worked, but this server requires a password change before commands can run. change the password first, then save the host")
	}

	if text == "" {
		return errors.New("ssh connection check failed")
	}

	return fmt.Errorf("ssh connection check failed: %s", text)
}

func runSSHCheck(host state.Host) tea.Cmd {
	return func() tea.Msg {
		err := verifySSH(
			host.Connect.IP,
			host.Connect.SSHPort,
			host.Connect.Password,
			host.Connect.SSHKeyPath,
		)
		return sshCheckResultMsg{
			host: host,
			err:  err,
		}
	}
}

func createAskpassScript(password string) (string, func(), error) {
	file, err := os.CreateTemp("", "newborn-ssh-askpass-*")
	if err != nil {
		return "", nil, err
	}

	script := "#!/bin/sh\nprintf '%s\\n' " + shellQuote(password) + "\n"
	if _, err := file.WriteString(script); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return "", nil, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(file.Name())
		return "", nil, err
	}
	if err := os.Chmod(file.Name(), 0o700); err != nil {
		_ = os.Remove(file.Name())
		return "", nil, err
	}

	return file.Name(), func() {
		_ = os.Remove(file.Name())
	}, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func generateHostSetup(displayName string) (state.HostSetup, error) {
	setup := state.HostSetup{
		Hostname:      "host-" + randomString(8, lowerAlphaNumeric),
		Name:          displayName,
		Username:      randomString(7, lowerAlphaNumeric),
		Password:      randomString(100, alphaNumeric),
		PasswordSalt:  randomString(16, alphaNumeric),
		SSHPort:       randomInt(1025, 65535),
	}

	privateKey, publicKey, err := generateSSHKeyPair()
	if err != nil {
		return state.HostSetup{}, err
	}
	setup.SSHPrivateKey = privateKey
	setup.SSHPublicKey = publicKey

	return setup, nil
}

func generateSSHKeyPair() (string, string, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("could not generate ssh keypair: %w", err)
	}

	block, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return "", "", fmt.Errorf("could not marshal private key: %w", err)
	}

	privatePEM := string(pem.EncodeToMemory(block))

	publicKey, err := ssh.NewPublicKey(privateKey.Public())
	if err != nil {
		return "", "", fmt.Errorf("could not derive public key: %w", err)
	}

	return privatePEM, strings.TrimSpace(string(ssh.MarshalAuthorizedKey(publicKey))), nil
}

const (
	alphaNumeric      = "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	lowerAlphaNumeric = "1234567890abcdefghijklmnopqrstuvwxyz"
)

func randomString(length int, alphabet string) string {
	if length <= 0 || len(alphabet) == 0 {
		return ""
	}

	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = alphabet[randomInt(0, len(alphabet)-1)]
	}

	return string(bytes)
}

func randomInt(minValue, maxValue int) int {
	if maxValue <= minValue {
		return minValue
	}

	size := big.NewInt(int64(maxValue - minValue + 1))
	value, err := rand.Int(rand.Reader, size)
	if err != nil {
		return minValue
	}

	return minValue + int(value.Int64())
}
