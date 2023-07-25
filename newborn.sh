#!/bin/bash

newborn_say () {
	echo '[NEWBORN] '$1
}

normalpath () {
	python3 -c 'import os,sys;print(os.path.abspath(os.path.expanduser(sys.argv[1])))' $1
}

RUN_ID=$(date +%s)
RUN_DIRECTORY="$PWD/.run/$RUN_ID"

mkdir -p $RUN_DIRECTORY > /dev/null

# connection
IP=''
PASSWORD=''
CONN_SSH_KEY_PATH=''
# setup
SERVER_NAME='server'
SWAP_SIZE=''
NEW_USER_NAME=$(LC_ALL=C tr -dc a-z0-9 < /dev/urandom | head -c 7)
NEW_USER_SUDO='n'
NEW_USER_PASSWORD=''
SSH_KEY_PATH=''
FIREWALL='n'
# packages
OCI_PLATFORM='none'
OCI_COMPOSE='n'
K8S=''
# output
OUT_INVENTORY_PATH=''
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
		--firewall)
			FIREWALL='y'
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
		--microk8s)
			K8S='microk8s'
			shift
			;;
		# output
		--append-inventory)
			OUT_INVENTORY_PATH=$2
			shift
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
			echo '  --ip, -h <ip>               IP address of the server'
			echo '  --password, -p <password>   Root user'"'"'s password'
			echo '  --password-stdin            Read root user'"'"'s password from stdin'
			echo '  --ssh-connect-key <path>    Path to SSH private key to connect to server'
			echo
			echo 'Setup options:'
			echo '  --name, -n <name>           Server name to use in Bash prompt; default: "server"'
			echo '  --swap <size>               Swap to add (e.g. "500M", "1G", "2G", "4G", etc.)'
			echo '  --user, -u <name>           New user name (default: "user")'
			echo '  --user-sudo                 Add the user to sudoers'
			echo '  --ask-new-password          Ask for new user password; otherwise random password will be generated'
			echo '  --ssh-key <path>            Path to new SSH key; otherwise it will be generated'
			echo '  --firewall                  Setup UFW firewall'
			echo
			echo 'Sowtware options:'
			echo '  --docker                    Install Docker'
			echo '  --podman                    Install Podman'
			echo '  --compose                   Install Docker Compose / Podman Compose'
			echo '  --microk8s                  Install MicroK8s'
			echo
			echo 'Output options:'
			echo '  --append-inventory <path>   Append processed hosts to Ansible inventory'
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

mkdir -p $RUN_DIRECTORY > /dev/null
mkdir -p $RUN_DIRECTORY/input > /dev/null
mkdir -p $RUN_DIRECTORY/output > /dev/null

INVENTORY_FILE="$RUN_DIRECTORY/input/inventory"
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
           -v "$RUN_DIRECTORY/output:/app/output" \
           -v "$INVENTORY_FILE:/app/input/inventory" \
		   -v "$CONN_SSH_KEY_PATH:/app/input/ssh.private.key" \
           -v "$SSH_KEY_PATH:$SSH_KEY_PATH_DOCKER" \
           -e "NEWBORN_SERVER_NAME=$SERVER_NAME" \
		   -e "NEWBORN_SWAP_SIZE=$SWAP_SIZE" \
           -e "NEWBORN_NEW_USER_NAME=$NEW_USER_NAME" \
           -e "NEWBORN_NEW_USER_PASSWORD=$NEW_USER_PASSWORD" \
           -e "NEWBORN_NEW_USER_SUDO=$NEW_USER_SUDO" \
		   -e "NEWBORN_FIREWALL=$FIREWALL" \
           -e "NEWBORN_OCI_PLATFORM=$OCI_PLATFORM" \
           -e "NEWBORN_OCI_COMPOSE=$OCI_COMPOSE" \
		   -e "NEWBORN_K8S=$K8S" \
           local/newborn

if [ $? -ne 0 ]; then
	newborn_say 'Error: Ansible playbook failed.'
	exit 1
fi

source $RUN_DIRECTORY/output/return.bash

echo
newborn_say 'Setup complete!'

export NEWBORN_IP=$IP

if [ "$OUT_INVENTORY_PATH" != '' ]; then
	if [ "$SSH_KEY_GENERATE" = 'y' ]; then
		SAVED_SSH_KEY_PATH=$OUT_SSH_KEY_PATH
	else
		SAVED_SSH_KEY_PATH=$SSH_KEY_PATH
	fi

	if [ "$SAVED_SSH_KEY_PATH" != '' ]; then
		cat $RUN_DIRECTORY/output/inventory | awk '{ print $0, "ansible_ssh_private_key_file='$(normalpath $SAVED_SSH_KEY_PATH)'" }' > $RUN_DIRECTORY/output/inventory.tmp
		mv $RUN_DIRECTORY/output/inventory.tmp $RUN_DIRECTORY/output/inventory
	fi

	cat $RUN_DIRECTORY/output/inventory >> $OUT_INVENTORY_PATH
	newborn_say 'Inventory appended to '$OUT_INVENTORY_PATH
else
	export NEWBORN_SSH_PORT=$(cat $RUN_DIRECTORY/output/inventory | grep ansible_ssh_port | cut -d'=' -f2)
fi

if [ "$OUT_SSH_KEY_PATH" != '' ]; then
	cp $RUN_DIRECTORY/output/ssh_key $OUT_SSH_KEY_PATH
	chmod 600 $OUT_SSH_KEY_PATH
	# newborn_say 'SSH key copied to '$OUT_SSH_KEY_PATH
	export NEWBORN_SSH_KEY_PATH=$OUT_SSH_KEY_PATH
else
	export NEWBORN_SSH_KEY_PATH=$(normalpath $SSH_KEY_PATH)
fi

export NEWBORN_USER=$NEW_USER_NAME
export NEWBORN_USER_PASSWORD
export NEWBORN_HOSTNAME

rm -rf $RUN_DIRECTORY > /dev/null
