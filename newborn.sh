#!/bin/bash

newborn_say () {
	echo '[NEWBORN] '$1
}

normalpath () {
	python3 -c 'import os,sys;print(os.path.abspath(os.path.expanduser(sys.argv[1])))' $1
}

# connection
IP=''
PASSWORD=''
CONN_SSH_KEY_PATH=''
# setup
SERVER_NAME='server'
SWAP_SIZE=''
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
		-h|--ip)
			IP=$2
			shift
			shift
			;;
		-p|--password)
			PASSWORD=$2
			shift
			shift
			;;
		--password-stdin)
			read -s -p '[NEWBORN] Enter current password for root@'$IP': ' PASSWORD
			echo
			shift
			;;
		--ssh-connect-key)
			CONN_SSH_KEY_PATH=$2
			shift
			shift
			;;
		# setup
		-n|--name)
			SERVER_NAME=$2
			shift
			shift
			;;
		--swap)
			SWAP_SIZE=$2
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
			read -s -p '[NEWBORN] Enter new password for '$NEW_USER_NAME'@'$IP': ' NEW_USER_PASSWORD
			echo
			shift
			;;
		--ssh-key)
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
			echo '  --ip <ip>                   IP address of the server'
			echo '  -p, --password <password>   Root user'"'"'s password'
			echo '  --ssh-connect-key <path>    Path to SSH private key to connect to server'
			echo
			echo 'Setup options:'
			echo '  -n, --name <name>           Server name to use in Bash prompt; default: "server"'
			echo '  --swap <size>               Swap to add (e.g. "500M", "1G", "2G", "4G", etc.)'
			echo '  -u, --user <name>           New user name (default: "user")'
			echo '  --user-sudo                 Add the user to sudoers'
			echo '  --ask-new-password          Ask for new user password; otherwise random password will be generated'
			echo '  --ssh-key <path>            Path to new SSH key; otherwise it will be generated'
			echo
			echo 'Sowtware options:'
			echo '  --docker                    Install Docker'
			echo '  --podman                    Install Podman'
			echo '  --compose                   Install Docker Compose / Podman Compose'
			echo
			echo 'Output options:'
			echo '  --append-inventory <path>   Append processed hosts to Ansible inventory'
			echo '  --print-password            Print new user password to stdout'
			echo '  --copy-ssh-key <path>       Copy SSH key to file'
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

INVENTORY_FILE="$PWD/.run/input/inventory"
if [ ! -z "$PASSWORD" ]; then
	echo "$IP ansible_ssh_pass=$PASSWORD" > $INVENTORY_FILE
	CONN_SSH_KEY_PATH='/tmp/nothing'
elif [ ! -z "$CONN_SSH_KEY_PATH" ]; then
	echo "$IP ansible_ssh_private_key_file=/app/input/ssh.private.key" > $INVENTORY_FILE
else
	newborn_say 'Error: neither password nor SSH private key provided.'
	exit 1
fi

if [ -z "$SSH_KEY_PATH" ]; then
	if [ -z "$OUT_SSH_KEY_PATH" ]; then
		newborn_say 'Error: SSH key will be generated, but output path is not provided. Use --ssh-key'
		exit 1
	fi

	SSH_KEY_GENERATE='y'
	SSH_KEY_PATH='/tmp/nothing'
	SSH_KEY_PATH_DOCKER='/tmp/nothing'
else
	if [ ! -f "$SSH_KEY_PATH" ]; then
		newborn_say 'Error: SSH key file '$SSH_KEY_PATH' does not exist.'
		exit 1
	fi

	SSH_KEY_GENERATE='n'
	SSH_KEY_PATH_DOCKER='/app/input/ssh_key'
fi

echo

docker build -t local/newborn .
docker run --interactive \
           --tty \
           --rm \
           -v "$PWD/.run/output:/app/output" \
           -v "$INVENTORY_FILE:/app/input/inventory" \
		   -v "$CONN_SSH_KEY_PATH:/app/input/ssh.private.key" \
           -v "$SSH_KEY_PATH:$SSH_KEY_PATH_DOCKER" \
           -e "NEWBORN_SERVER_NAME=$SERVER_NAME" \
		   -e "NEWBORN_SWAP_SIZE=$SWAP_SIZE" \
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
