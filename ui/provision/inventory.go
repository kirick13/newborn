package provision

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kirick13/newborn/state"
)

func BuildInventory(hosts []state.Host) string {
	lines := []string{
		"all:",
		"  hosts:",
	}

	for _, host := range hosts {
		lines = append(lines, fmt.Sprintf("    %s:", yamlScalar(host.Connect.IP)))

		if host.Connect.SSHPort != 22 {
			lines = append(lines, fmt.Sprintf("      ansible_ssh_port: %d", host.Connect.SSHPort))
		}
		if host.Connect.Password != "" {
			lines = append(lines, fmt.Sprintf("      ansible_ssh_pass: %s", yamlScalar(host.Connect.Password)))
		}
		if host.Connect.SSHKeyPath != "" {
			lines = append(lines, fmt.Sprintf(
				"      ansible_ssh_private_key_file: %s",
				yamlScalar(fmt.Sprintf("/opt/bind/ssh/%s.key", host.Setup.Name)),
			))
		}

		lines = append(lines,
			fmt.Sprintf("      newborn_hostname: %s", yamlScalar(host.Setup.Hostname)),
			fmt.Sprintf("      newborn_name: %s", yamlScalar(host.Setup.Name)),
			fmt.Sprintf("      newborn_user: %s", yamlScalar(host.Setup.Username)),
			fmt.Sprintf("      newborn_password: %s", yamlScalar(host.Setup.Password)),
			fmt.Sprintf("      newborn_password_salt: %s", yamlScalar(host.Setup.PasswordSalt)),
			fmt.Sprintf("      newborn_ssh_port: %d", host.Setup.SSHPort),
			fmt.Sprintf("      newborn_ssh_key_public: %s", yamlScalar(host.Setup.SSHPublicKey)),
		)
	}

	return strings.Join(lines, "\n") + "\n"
}

func WriteInventoryFile(content string) (string, error) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("inventory.%d.yaml", time.Now().UnixMilli()))
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", err
	}

	return path, nil
}

func yamlScalar(value string) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return `""`
	}

	return string(bytes)
}
