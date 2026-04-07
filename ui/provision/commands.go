package provision

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirick13/newborn/state"
)

func BuildDockerRunCommand(app *state.Newborn) string {
	payload := struct {
		NewbornSwap        string `json:"newborn_swap"`
		NewbornReserveFile string `json:"newborn_reserve_file"`
		NewbornFirewall    string `json:"newborn_firewall_http"`
		NewbornOCIRuntime  string `json:"newborn_oci_runtime"`
		NewbornOCICompose  string `json:"newborn_oci_compose"`
		NewbornK8sRuntime  string `json:"newborn_k8s_runtime"`
		NewbornRemoveSnap  string `json:"newborn_remove_snap"`
	}{
		NewbornSwap:        app.Setup.Swap,
		NewbornReserveFile: boolToFlag(app.Setup.ReserveFile),
		NewbornFirewall:    app.Setup.FirewallHTTP,
		NewbornOCIRuntime:  app.Software.OCIRuntime,
		NewbornOCICompose:  boolToFlag(app.Software.OCICompose),
		NewbornK8sRuntime:  app.Software.K8sRuntime,
		NewbornRemoveSnap:  boolToFlag(app.Software.RemoveSnap),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "docker run -t --rm local/newborn -e '<could not build payload>'"
	}

	args := []string{
		// "echo",
		"docker",
		"run",
		"-t",
		"--rm",
	}

	if inventoryPath := strings.TrimSpace(app.InventoryPath); inventoryPath != "" {
		args = append(args, "-v", shellQuote(inventoryPath+":/app/inventory.yaml:ro"))
	}

	for _, host := range app.Hosts {
		if strings.TrimSpace(host.Connect.SSHKeyPath) == "" {
			continue
		}

		args = append(args,
			"-v",
			shellQuote(host.Connect.SSHKeyPath+":/opt/bind/ssh/"+host.Setup.Name+".key:ro"),
		)
	}

	args = append(args,
		"local/newborn",
		"-e",
		shellQuote(string(jsonPayload)),
	)

	return strings.Join(args, " ")
}

func BuildDockerBuildCommand() string {
	return "docker build -t local/newborn " + shellQuote(filepath.Join(repoRoot(), "container"))
}

func BuildProvisionShellScript(app *state.Newborn) string {
	return strings.Join([]string{
		"printf '\\033[3J\\033[2J\\033[H'",
		"set -e",
		BuildDockerBuildCommand(),
		BuildDockerRunCommand(app),
	}, "\n")
}

func BuildDebugText(app *state.Newborn) string {
	parts := []string{}
	if content := strings.TrimSpace(app.InventoryContent); content != "" {
		parts = append(parts, content)
	}
	parts = append(parts, BuildDockerRunCommand(app))
	return strings.Join(parts, "\n\n")
}

func repoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	candidates := []string{cwd, filepath.Dir(cwd)}
	for _, candidate := range candidates {
		if _, err := os.Stat(filepath.Join(candidate, "container", "Dockerfile")); err == nil {
			return candidate
		}
	}

	return cwd
}

func boolToFlag(value bool) string {
	if value {
		return "y"
	}

	return ""
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func ClearTerminal() {
	fmt.Print("\033[3J\033[2J\033[H")
}
