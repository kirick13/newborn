package provision

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kirick13/newborn/state"
	"golang.org/x/crypto/ssh"
)

func UnlockBitwarden(password string) (string, error) {
	if _, err := exec.LookPath("bw"); err != nil {
		return "", fmt.Errorf("bw command is required: %w", err)
	}

	cmd := exec.Command("bw", "unlock", "--raw", "--passwordenv", "BW_PASSWORD")
	cmd.Env = append(os.Environ(), "BW_PASSWORD="+password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", formatCommandError("bw unlock failed", output, err)
	}

	session := strings.TrimSpace(string(output))
	if session == "" {
		return "", fmt.Errorf("bw unlock failed: empty session returned")
	}

	return session, nil
}

func CreateSSHKeyItems(session string, hosts []state.Host) error {
	for _, host := range hosts {
		if err := createSSHKeyItem(session, host); err != nil {
			return err
		}
	}

	return nil
}

func createSSHKeyItem(session string, host state.Host) error {
	fingerprint, err := sshFingerprint(host.Setup.SSHPublicKey)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"organizationId": nil,
		"folderId":       nil,
		"type":           5,
		"name":           host.Setup.Name,
		"notes":          "",
		"favorite":       false,
		"fields": []map[string]any{
			{"name": "IP", "value": host.Connect.IP, "type": 0},
			{"name": "IPv6", "value": "", "type": 0},
			{"name": "SSH port", "value": fmt.Sprintf("%d", host.Setup.SSHPort), "type": 0},
			{"name": "hostname", "value": host.Setup.Hostname, "type": 0},
			{"name": "username", "value": host.Setup.Username, "type": 0},
			{"name": "password", "value": host.Setup.Password, "type": 1},
		},
		"sshKey": map[string]any{
			"privateKey":     host.Setup.SSHPrivateKey,
			"publicKey":      host.Setup.SSHPublicKey,
			"keyFingerprint": fingerprint,
		},
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(bytes)
	cmd := exec.Command("bw", "create", "item", encoded)
	cmd.Env = append(os.Environ(),
		"BW_SESSION="+session,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return formatCommandError("bw create item failed", output, err)
	}

	return nil
}

func sshFingerprint(publicKey string) (string, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return "", fmt.Errorf("could not parse ssh public key: %w", err)
	}

	sum := sha256.Sum256(key.Marshal())
	return "SHA256:" + base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:]), nil
}

func formatCommandError(prefix string, output []byte, err error) error {
	text := strings.TrimSpace(string(output))
	if text == "" {
		return fmt.Errorf("%s: %w", prefix, err)
	}

	return fmt.Errorf("%s: %s", prefix, text)
}
