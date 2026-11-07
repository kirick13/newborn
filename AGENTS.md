# Project knowledge

## Product scope

- Newborn bootstraps fresh Ubuntu servers. The current targets are Ubuntu 22–26
  on AMD64 (`x86_64`) and ARM64 (`aarch64`), matching the CLI installation pipeline.
- This is strictly a first-setup tool. Existing installations, repeat provisioning,
  migrations, and general configuration management are outside the current scope.
- The terminal application is Go-only. Keep implementation code under `ui/` and
  a thin `main.go` at the repository root. Root `go.mod` and `go.sum` allow `go run .`.
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
- `container/scripts/apt_sources.py` checks active HTTP/HTTPS repository URLs in
  `/etc/apt/sources.list` and `/etc/apt/sources.list.d/ubuntu.sources` concurrently
  from the target before updating. Only files containing a failing URL are replaced
  with official sources, in their existing format, without backups. Healthy files
  and other source files (including `docker.list`) are preserved. AMD64 uses
  archive.ubuntu.com and security.ubuntu.com; ARM64 uses ports.ubuntu.com. Sources
  use the host codename.
- Keep the APT checker self-contained using Python 3's standard library and APIs
  available on Ubuntu 22–26. It builds both source formats internally; no Jinja
  templates or pip packages are needed. Invoke it with Ansible's `script` module,
  like `remove_snap.sh`, so it is staged temporarily instead of installed remotely.
- Keep APT operations as direct `apt` tasks. The initial update/upgrade uses native
  `lock_timeout: 7200`, with two task retries and one cache refresh attempt per run.
  Ansible 2.17's timeout handles python-apt lock exceptions, not every lock failure
  from its apt-get subprocesses. Do not reintroduce a generic retry wrapper.
- Use `ansible_architecture` as the architecture source. Translate it through
  `newborn_apt_arch` for the Ubuntu source script and Docker's APT repository, and
  through `newborn_compose_arch` for Compose download filenames. Keep these two
  mappings separate; for example, `x86_64` becomes `amd64` for APT and stays
  `x86_64` for Compose. Do not invoke `dpkg --print-architecture` for architecture
  detection.
- Validate Ubuntu version and architecture immediately after fact gathering,
  before any target changes. All architectures other than `x86_64` and `aarch64`
  must fail early. The workflow assumes initial APT contention clears once and no
  competing package manager starts afterward.
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
- `iptables.sh` writes the `NEWBORN` chain to temporary files and loads each family
  with one `iptables-restore --noflush`, so the firewall is never briefly open and a
  failure keeps the current rules. A failed Cloudflare download reapplies the ranges
  already loaded, so the SSH rule still updates; only the first run, before SSH changes,
  can fail on it. The playbook applies the new SSH port before restarting sshd. Jumps
  are inserted only when missing and never moved back to the top; conflicts with other
  firewall software are the user's responsibility. Match ports with
  `-m conntrack --ctorigdstport`, not `--dport`: Docker's DNAT rewrites ports before
  `FORWARD`/`DOCKER-USER`. Allow
  rules use `RETURN`, not `ACCEPT`. The playbook sets `newborn_firewall_http` as the
  argument of the `allow_http '<mode>'` call; `nowhere` falls to the default case
  and opens no HTTP ports.
- After provisioning succeeds, the UI unlocks Bitwarden and saves an SSH-key item
  per host. The account must already be logged in through `bw`.
- Generated private keys currently remain in memory until Bitwarden saves them.
  Inventory files contain passwords, use mode `0600`, and are not currently cleaned
  up automatically. Do not expose credentials in logs, committed fixtures, or docs.
- k0s installation does not start or join a cluster.
- CLI tools are installed through Mise's GitHub backend: bat, bottom (`btm`), dust,
  eza, fd, and ripgrep (`rg`). Install `extrepo` and Mise with native `apt` tasks;
  enable the Mise repository with `extrepo enable mise`. Provisioning runs as root,
  so these tasks do not use `sudo`.
- The CLI installation shell must stop on any download, executable lookup, or copy
  failure. Keep `mise which` in a separate assignment so its exit status is checked.
  Copy binaries into `/usr/local/bin` with mode `0755`; the new user needs neither
  Mise activation nor `/opt/rust-utils` on PATH. Tool versions currently resolve to
  latest, and no Rust/Cargo toolchain is installed by this workflow.

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
- Offline platform/source regression tests live in `tests/test_apt.py`; CLI shell
  failure handling is covered by `tests/test_cli_tools.py` and firewall rule loading
  by `tests/test_iptables.py`, both using stubbed commands. Run them in the Ansible
  Docker image; see `docs/testing.md` for the command.
  There are no Go test cases or CI configuration. `go test ./...` alone does not
  establish that provisioning works.
- For meaningful Go tests, prioritize inventory serialization, command quoting,
  option validation, and Bitwarden payload/error handling. Stub external commands
  in unit tests rather than contacting servers or a real vault.
- Provisioning tests require a fresh disposable VM for each scenario. Do not use
  an existing server or treat rerunning the playbook as an idempotence requirement.
- Use full VMs to exercise the kernel, swap, systemd, SSH, and firewall behavior.
  See `docs/testing.md` for the proposed Multipass workflow and its limitations.
- The initial proposed VM matrix is Ubuntu 22.04, 24.04, and 26.04. This is a test
  proposal, not evidence that these releases have passed. Cover both AMD64 and
  ARM64; interim-release coverage still needs to be agreed.
- Distinguish local static/build checks, VM provisioning verification, and real
  Bitwarden integration checks when reporting results.
