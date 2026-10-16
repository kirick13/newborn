# Project knowledge

## Product scope

- Newborn bootstraps fresh Ubuntu servers. The current target range is Ubuntu 22–26.
- This is strictly a first-setup tool. Existing installations, repeat provisioning,
  migrations, and general configuration management are outside the current scope.
- The application is Go-only. Keep implementation code under `ui/` and a thin
  `main.go` at the repository root. Root `go.mod` and `go.sum` allow `go run .`.
- Bitwarden CLI is mandatory for now. Preserve the current Bitwarden flow unless
  a task explicitly changes it. Other password managers and local-directory export
  are future possibilities; do not implement them speculatively.

## Architecture

- `main.go` calls `ui.Run()`; `ui/app.go` owns the Bubble Tea model and event loop.
- `ui/components/view/` contains host management, setup, software, and Bitwarden
  screens. Shared controls live under `ui/components/` and `ui/elements/`.
- `ui/state/` holds host credentials and setup/software selections in memory.
- `ui/config/env.go` reads optional `.env` defaults: `NAME`, `IP`, `SSH_KEY_PATH`.
- `ui/provision/` creates the Ansible inventory, builds Docker commands, and calls
  the `bw` CLI. The Go module path is `github.com/kirick13/newborn`; internal imports
  include `/ui/`.
- `container/Dockerfile` builds the Ansible runner as `local/newborn`.
- `container/playbook.yaml` performs the remote setup; `container/scripts/` and
  `container/configs/` contain its supporting assets.
- Run the application from the repository root so the container build context and
  local defaults are available.

## Provisioning contract

- Initial access is root SSH with either a password or an identity file. The host
  form checks connectivity before accepting a host.
- Host-specific inventory variables include generated hostname, username, password,
  password salt, SSH port, and public key. Keep their `newborn_*` names aligned with
  the playbook when changing either side.
- Setup/software flags are passed as JSON extra variables; boolean flags currently
  use `"y"` or an empty string.
- Local Docker is required even when the target does not install Docker.
- Ansible changes root credentials, creates a sudo user, disables root SSH and
  password authentication, changes the SSH port, and replaces firewall rules.
- After provisioning succeeds, the UI unlocks Bitwarden and saves an SSH-key item
  per host. The account must already be logged in through `bw`.
- Generated private keys currently remain in memory until Bitwarden saves them.
  Inventory files contain passwords, use mode `0600`, and are not currently cleaned
  up automatically. Do not expose credentials in logs, committed fixtures, or docs.
- k0s installation does not start or join a cluster. Rust CLI utility installation
  is currently commented out in the playbook.

## Development and validation

Run checks from the repository root:

```sh
go build ./...
go test ./...
go vet ./...
bash -n container/scripts/swap.sh
sh -n container/scripts/iptables.sh
sh -n container/scripts/remove_snap.sh
```

- Format changed Go files with `gofmt`.
- There are currently no automated test cases or CI configuration. A successful
  `go test ./...` alone does not establish that provisioning works.
- For meaningful Go tests, prioritize inventory serialization, command quoting,
  option validation, and Bitwarden payload/error handling. Stub external commands
  in unit tests rather than contacting servers or a real vault.
- Provisioning tests require a fresh disposable VM for each scenario. Do not use
  an existing server or treat rerunning the playbook as an idempotence requirement.
- Use full VMs to exercise the kernel, swap, systemd, SSH, and firewall behavior.
  See `docs/testing.md` for the proposed Multipass workflow and its limitations.
- The initial proposed VM matrix is Ubuntu 22.04, 24.04, and 26.04. This is a test
  proposal, not evidence that these releases have passed. Interim-release and CPU
  architecture coverage still need to be agreed.
- Distinguish local static/build checks, VM provisioning verification, and real
  Bitwarden integration checks when reporting results.
