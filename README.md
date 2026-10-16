# newborn

Bootstrap fresh Ubuntu servers through an interactive Go terminal UI and Ansible.
The target range is Ubuntu 22–26. Newborn is for first setup only; running it on
existing installations or rerunning it on an already configured server is outside
the supported workflow.

## What it does

- Updates the system and installs basic packages.
- Replaces or removes swap, with an optional 2 GB emergency disk reserve.
- Generates a hostname, user, password, Ed25519 SSH key, and SSH port.
- Grants sudo access and disables root SSH login and password authentication.
- Configures IPv4/IPv6 firewall rules, with optional public or Cloudflare-only HTTP(S).
- Optionally installs Docker or Podman, Compose, k0s or MicroK8s, and removes Snap.
- Sets UTC time, a UTF-8 locale, a custom Bash prompt, and Ghostty terminfo.
- Saves each server's generated SSH key and credentials to Bitwarden.

The k0s option installs the binary and generates configuration; controller/worker
startup and cluster joining remain manual.

## Requirements

- Go 1.25 or newer, subject to dependency requirements in `go.mod`.
- Docker CLI and a running Docker engine, reachable from the local shell.
- OpenSSH client (`ssh`).
- Bitwarden CLI (`bw`), already logged in to your account.
- A fresh Ubuntu server reachable as root using a password or SSH identity file.

Bitwarden is mandatory. Other password managers and saving credentials to a local
directory are possible future options, not implemented modes.

## Run

From the repository root:

```sh
go run .
```

Add hosts, choose setup and software options, and continue to provisioning.
The app builds `local/newborn` from `container/`, then runs Ansible in that image.
After successful provisioning, it asks for your Bitwarden master password and
saves one SSH-key item per host.

Use Tab/Shift+Tab to move between controls, Space to toggle options, Enter to
continue, and Esc to go back where available. Ctrl+C quits the application.

An optional root `.env` file supplies host-form defaults. Use plain, unquoted
values:

```dotenv
NAME=test-server
IP=192.0.2.10
SSH_KEY_PATH=/absolute/path/to/bootstrap-key
```

Generated private keys currently remain in memory until the Bitwarden save
succeeds. Finish that step before exiting; there is no local key-file fallback.

## Layout

- `main.go`, `go.mod`, `go.sum`: entry point and Go module configuration.
- `ui/`: application, terminal components, state, defaults, and provisioning helpers.
- `container/`: Ansible image, playbook, target configuration, and shell scripts.
- `AGENTS.md`: project decisions and guidance for future changes.
- `docs/testing.md`: local checks and the proposed fresh-VM testing workflow.

## Development checks

Run from the repository root:

```sh
go build ./...
go test ./...
go vet ./...
bash -n container/scripts/swap.sh
sh -n container/scripts/iptables.sh
sh -n container/scripts/remove_snap.sh
```

Use `gofmt` on changed Go files. There are currently no automated test cases;
`go test ./...` checks package compilation but does not validate server setup.
See [the testing plan](docs/testing.md) for how to test real provisioning.
