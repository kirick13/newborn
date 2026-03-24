# newborn

Init your newly created Ubuntu webservers in minutes.

## What does it do?

- updates the system;
- adds (or removes) swap;
- installs Docker/Podman (and compose), k0s/microk8s;
- creates a new user:
    - with random password;
    - with SSH key;
    - adds sudo access to it;
- secures SSH:
    - moves it to random port;
    - disables password authentication;
    - disables root login;
- adds firewall rules;
- creates unique server's hostname;
- creates pretty Bash prompt;
- and many more stuff like UTC timezone and Unicode locale.

## Usage

### Install Bun

```sh
curl -fsSL https://bun.sh/install | bash
```

### Clone this repository

```sh
git clone https://github.com/kirick13/newborn.git
cd newborn
```

### Run it!

```sh
bun run newborn.ts
```

CLI will interactively ask you to provide connection options and setup/software options.
