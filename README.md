
# newborn

Init your newly created Ubuntu webservers in a minutes.

## What does it do?

- Updates the system;
- installs Docker or Podman;
- installs Docker Compose or Podman Compose;
- creates a new user:
    - with random password (**same** on each host);
    - with SSH key (**same** on each host);
    - adds (or not) sudo access to it;
- secures SSH:
    - moves it to random port (**different** on each host);
    - disables password authentication;
    - disables root login;
- hides server's hostname;
- creates pretty Bash prompt.

## Usage

### Step 1: Clone this repository.

```
git clone https://github.com/kirick13/newborn.git
```

### Step 2: build docker image.

```
docker build -t newborn .
```

### Step 3: Create an Ansible inventory

Create file by `ansible/inventory` path. Fill this file with your servers — Ansible will connect to them as `root` with password.

Optionally, you can add a webserver name to each host using `newborn_server_name` option. That name will only be visible in Bash prompt. For example:

```
10.0.13.1 ansible_password=so0AIkp50XS3JsA3 newborn_server_name=server-1
10.0.13.2 ansible_password=1HKEmF7sSKimp68M newborn_server_name=server-2
```

### Step 4: Run!

```
docker run --rm \
           -v "$PWD/ansible/inventory:/app/ansible/inventory" \
           -v "$PWD/output:/data/newborn" \
           git.kirick.me/kirick/newborn \
           --user iambread \
           --user-sudo \
           --docker \
           --compose
```

It's important to mount following directories to the container:

- `/app/ansible/inventory` — your Ansible inventory file;
- `/data/newborn` — empty directory that will be filled with some important files:
    - `inventory` — ansible inventory with SSH ports appended;
    - `password.txt` — new user's password;
    - `ssh.private.key` — new user's SSH private key.

There are some options you can pass to the script itself:

| Option | Description |
| - | - |
| `-n`, `--name` | Server name if you don't want to specify server names individually using `newborn_server_name` option. |
| `-u`, `--user` | Create a new user with this name. |
| `--user-sudo` | Add the user to sudoers. |
| `--docker` | Install Docker. |
| `--podman` | Install Podman. |
| `--compose` | Install Compose for Docker or Podman. |

## TODO:

- Connect to servers via root's SSH key.
