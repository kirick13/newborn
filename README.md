
# newborn

Init your newly created Ubuntu webservers in a minutes.

## What does it do?

- updates the system;
- adds swap;
- installs Docker or Podman;
- installs Docker Compose or Podman Compose;
- creates a new user:
    - with random password (**same** on each host);
    - with SSH key (**same** on each host);
    - adds sudo access to it;
- secures SSH:
    - moves it to random port (**different** on each host);
    - disables password authentication;
    - disables root login;
- hides server's hostname;
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
| `--password-stdin` | Enter root user's password via stdin. <br> Incompatible with `--password`, `--ssh-connect-key`. <br> Alias: `-p`. |
| `--ssh-connect-key <path>` | Path to SSH private key to connect to server. <br> Incompatible with `--password`, `--password-stdin`. |

#### Setup options

| Option | Description |
| - | - |
| `--name <name>` | Server name to use in Bash prompt. <br> Default: `server`. <br> Alias: `-n`. |
| `--swap <size>` | Swap to add (e.g. "500M", "1G", "2G", "4G", etc.) |
| `--user <name>` | New user name. <br> Default: `user`. <br> Alias: `-u`. |
| `--user-sudo` | Add the user to sudoers. |
| `--ask-new-password` | Ask for new user password; otherwise random password will be generated. |
| `--ssh-key <path>` | Path to new SSH key; otherwise it will be generated. |

#### Sowtware options

| Option | Description |
| - | - |
| `--docker` | Install Docker. <br> Incompatible with `--podman`. |
| `--podman` | Install Podman. <br> Incompatible with `--docker`. |
| `--compose` | Install Compose for Docker or Podman. |

#### Output options

| Option | Description |
| - | - |
| `--append-inventory <path>` | Append processed host to Ansible inventory. |
| `--print-password` | Print new user password to stdout. |
| `--copy-ssh-key <path>` | Copy SSH key to file. |

## TODO:

- Support different distros.
