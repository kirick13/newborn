
# newborn

Init your newly created Ubuntu webservers in a minutes.

## What does it do?

- updates the system;
- adds (or removes) swap;
- installs Docker, Podman, Compose, microk8s;
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
- creates pretty Bash prompt.

## Usage

### Step 1: Clone this repository

```
git clone https://github.com/kirick13/newborn.git
```

### Step 2: Run it!

```
./newborn.sh --ip 10.0.13.10 \
             --password 123456 \
             --name myhost \
             --user catloaf \
             --user-sudo \
             --docker \
             --compose \
             --print-password \
             --copy-ssh-key ~/ssh_keys/myhost.key
```

#### Connection options

| Option | Description |
| - | - |
| `--ip <ip>` | IP address of the server. <br> Alias: `-h`. |
| `--password <password>` | Root user's password. <br> Incompatible with `--password-stdin`, `--ssh-connect-key`. <br> Alias: `-p`. |
| `--password-stdin` | Enter root user's password via stdin. <br> Incompatible with `--password`, `--ssh-connect-key`. |
| `--ssh-connect-key <path>` | Path to SSH private key to connect to server. <br> Incompatible with `--password`, `--password-stdin`. |

#### Setup options

| Option | Description |
| - | - |
| `--name <name>` | Server name to use in Bash prompt. <br> Default: `server`. <br> Alias: `-n`. |
| `--swap <size>` | Swap to add (e.g. "500M", "1G", "2G", "4G", etc.). <br> By default, swap will be disabled. |
| `--user <name>` | New user name. <br> Default: `user`. <br> Alias: `-u`. |
| `--user-sudo` | Add the user to sudoers. |
| `--ask-new-password` | Ask for new user password; otherwise random password will be generated. |
| `--ssh-key <path>` | Path to new SSH key; otherwise it will be generated. |
| `--firewall` | Enable iptables rules. <br> That will disable all incoming connections except current SSH port and ports 80 and 443 from Cloudflare. To change rules, edit `/root/iptables.sh` script on your server. Cron will re-apply iptables rules every day at 3 AM to maintain actual Cloudflare IPs. |

#### Software options

| Option | Description |
| - | - |
| `--docker` | Install Docker. <br> Incompatible with `--podman`. |
| `--podman` | Install Podman. <br> Incompatible with `--docker`. |
| `--compose` | Install Compose for Docker or Podman. |
| `--microk8s` | Install MicroK8S. |

#### Output options

| Option | Description |
| - | - |
| `--append-inventory <path>` | Append processed host to Ansible inventory. |
| `--copy-ssh-key <path>` | Copy SSH key to file. |

## TODO:

- Support different distros.
