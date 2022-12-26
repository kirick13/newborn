#!/bin/bash

newborn_say () {
    echo '[NEWBORN] '$1
}

normalpath () {
    python3 -c 'import os,sys; print(os.path.abspath(sys.argv[1]))' $1
}

# connection
INVENTORY_FILE=''
IP=''
PASSWORD=''
# setup
SERVER_NAME='server'
NEW_USER_NAME='user'
NEW_USER_SUDO='n'
NEW_USER_PASSWORD=''
SSH_KEY_PATH=''
# packages
OCI_PLATFORM='none'
OCI_COMPOSE='n'
# output
OUT_INVENTORY_PATH=''
OUT_PRINT_PASSWORD='n'
OUT_SSH_KEY_PATH=''

while [[ $# -gt 0 ]]; do
    case $1 in
        # connection
        -i|--inventory)
            INVENTORY_FILE=$2
            shift
            shift
            ;;
        --ip)
            IP=$2
            shift
            shift
            read -s -p '[NEWBORN] Enter current password for '$IP': ' PASSWORD
            echo
            ;;
        # setup
        -n|--name)
            SERVER_NAME=$2
            shift
            shift
            ;;
        -u|--user)
            NEW_USER_NAME=$2
            shift
            shift
            ;;
        --user-sudo)
            NEW_USER_SUDO='y'
            shift
            ;;
        --ask-new-password)
            read -s -p '[NEWBORN] Enter new password for '$IP': ' NEW_USER_PASSWORD
            echo
            shift
            ;;
        -k|--ssh-key)
            SSH_KEY_PATH=$2
            shift
            shift
            ;;
        # packages
        --docker)
            OCI_PLATFORM='docker'
            shift
            ;;
        --podman)
            OCI_PLATFORM='podman'
            shift
            ;;
        --compose)
            OCI_COMPOSE='y'
            shift
            ;;
        # output
        --append-inventory)
            OUT_INVENTORY_PATH=$2
            shift
            shift
            ;;
        --print-password)
            OUT_PRINT_PASSWORD='y'
            shift
            ;;
        --copy-ssh-key)
            OUT_SSH_KEY_PATH=$2
            shift
            shift
            if [ -f "$OUT_SSH_KEY_PATH" ]; then
                echo '[NEWBORN] Error: file '$OUT_SSH_KEY_PATH' already exists'
                exit 1
            fi
            ;;
        # other
        --help)
            echo
            echo 'Newborn setups new server with dockerized Ansible.'
            echo
            echo 'Usage: ./newborn.sh [options]'
            echo
            echo 'Connection options:'
            echo '  -i, --inventory <path>       Path to Ansible inventory file'
            echo '  --ip <ip>                    IP address of the server (password will be asked)'
            echo
            echo 'Setup options:'
            echo '  -n, --name <name>            Server name (will only be used in Bash prompt, default: "server")'
            echo '  -u, --user <name>            New user name (default: "user")'
            echo '  --user-sudo                  Add new user to sudoers'
            echo '  --ask-new-password           Ask for new user password (otherwise random password will be generated)'
            echo '  -k, --ssh-key <path>         Path to SSH key (otherwise new key will be generated)'
            echo
            echo 'Packages options:'
            echo '  --docker                     Install Docker'
            echo '  --podman                     Install Podman'
            echo '  --compose                    Install Docker Compose / Podman Compose'
            echo
            echo 'Output options:'
            echo '  --append-inventory <path>    Append processed hosts to Ansible inventory'
            echo '  --print-password             Print new user password to stdout'
            echo '  --copy-ssh-key <path>        Copy SSH key to file'
            echo
            exit 0
            ;;
        -*|--*)
            newborn_say 'Unknown argument '$1
            exit 1
            ;;
        *)
            newborn_say 'Unknown argument '$1
            exit 1
            ;;
    esac
done

rm -rf ./.run > /dev/null

mkdir -p ./.run > /dev/null
mkdir -p ./.run/input > /dev/null
mkdir -p ./.run/output > /dev/null

if [ -z "$INVENTORY_FILE" ]; then
    if [ -z "$IP" ]; then
        echo "Error: neither IP address nor Ansible inventory provided"
        exit 1
    fi
    if [ -z "$PASSWORD" ]; then
        echo "Error: neither password nor Ansible inventory provided. Use -p or --password"
        exit 1
    fi

    echo "$IP ansible_ssh_pass=$PASSWORD" > .run/input/inventory
    INVENTORY_FILE="$PWD/.run/input/inventory"
fi

if [ -z "$SSH_KEY_PATH" ]; then
    if [ -z "$OUT_SSH_KEY_PATH" ]; then
        echo "Error: SSH key will be generated, but output path is not provided. Use -k or --ssh-key"
        exit 1
    fi

    SSH_KEY_GENERATE='y'
    SSH_KEY_PATH='/tmp/nothing'
    SSH_KEY_PATH_DOCKER='/tmp/nothing'
else
    SSH_KEY_GENERATE='n'
    SSH_KEY_PATH_DOCKER='/app/input/ssh_key'
fi

echo

docker build -t local/newborn .
docker run --rm \
           -v "$PWD/.run/output:/app/output" \
           -v "$INVENTORY_FILE:/app/input/inventory" \
           -v "$SSH_KEY_PATH:$SSH_KEY_PATH_DOCKER" \
           -e "NEWBORN_SERVER_NAME=$SERVER_NAME" \
           -e "NEWBORN_NEW_USER_NAME=$NEW_USER_NAME" \
           -e "NEWBORN_NEW_USER_PASSWORD=$NEW_USER_PASSWORD" \
           -e "NEWBORN_NEW_USER_SUDO=$NEW_USER_SUDO" \
           -e "NEWBORN_OCI_PLATFORM=$OCI_PLATFORM" \
           -e "NEWBORN_OCI_COMPOSE=$OCI_COMPOSE" \
           local/newborn

echo
newborn_say 'Setup complete!'

if [ "$OUT_INVENTORY_PATH" != '' ]; then
    if [ "$SSH_KEY_GENERATE" = 'y' ]; then
        SAVED_SSH_KEY_PATH=$OUT_SSH_KEY_PATH
    else
        SAVED_SSH_KEY_PATH=$SSH_KEY_PATH
    fi

    if [ "$SAVED_SSH_KEY_PATH" != '' ]; then
        cat .run/output/inventory | awk '{ print $0, "ansible_ssh_private_key_file='$(normalpath $SAVED_SSH_KEY_PATH)'" }' > .run/output/inventory.tmp
        mv .run/output/inventory.tmp .run/output/inventory
    fi

    cat .run/output/inventory >> $OUT_INVENTORY_PATH
    newborn_say 'Inventory appended to '$OUT_INVENTORY_PATH
else
    newborn_say 'SSH port is: '$(cat .run/output/inventory | grep ansible_ssh_port | cut -d'=' -f2)
fi

if [ "$OUT_PRINT_PASSWORD" = 'y' ]; then
    if [ -z "$NEW_USER_PASSWORD" ]; then
        newborn_say 'New password: '$(cat .run/output/password.txt)
    fi
fi

if [ "$OUT_SSH_KEY_PATH" != '' ]; then
    cp .run/output/ssh_key $OUT_SSH_KEY_PATH
    chmod 600 $OUT_SSH_KEY_PATH
    newborn_say 'SSH key copied to '$OUT_SSH_KEY_PATH
fi

rm -rf ./.run > /dev/null
