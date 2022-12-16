
#!/bin/bash

SERVER_NAME="server"
NEW_USER_NAME="user"
NEW_USER_SUDO="n"
OCI_PLATFORM="none"
OCI_COMPOSE="n"
REST_ARGS=()

while [[ $# -gt 0 ]]; do
	case $1 in
		-n|--name)
			SERVER_NAME=$2
			shift
			shift
			;;
		--user)
			NEW_USER_NAME=$2
			shift
			shift
			;;
		--user-sudo)
			NEW_USER_SUDO="y"
			shift
			;;
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
		-*|--*)
			echo "Unknown argument $1"
			exit 1
			;;
		*)
			REST_ARGS+=("$1")
			shift
			;;
	esac
done

set -- "${REST_ARGS[@]}"

echo
echo '----- [NEWBORN] -----'
echo

newborn_say () {
	echo "[NEWBORN] $1"
}

echo
newborn_say "Creating directory for user's credentials..."
mkdir -p /data/newborn > /dev/null
newborn_say "done."

echo
newborn_say "Append random SSH port to every host in inventory..."
rm /data/newborn/inventory
cat ansible/inventory \
	| while read line
	do \
		SSH_PORT=$(python3 -c 'import random; print(random.randint(1025,65535))')
		echo ${line} | awk '{ print $1, "ansible_port='$SSH_PORT'" }' >> /data/newborn/inventory
		echo ${line}" newborn_ssh_port=$SSH_PORT"
	done \
	| tee ansible/inventory.local > /dev/null
newborn_say "done."

echo
newborn_say "Creating SSH key for user \"$NEW_USER_NAME\"..."
cd /data/newborn
ssh-keygen -t ecdsa \
           -m PEM \
           -b 521 \
           -N '' \
           -f SSHKEY \
           > /dev/null
cat SSHKEY.pub | awk '{ print $1,$2 }' > ssh.public.key
rm SSHKEY.pub
mv SSHKEY ssh.private.key
newborn_say "done."
cd /app

echo
newborn_say "Creating password for user \"$NEW_USER_NAME\"..."
NEW_USER_PASSWORD="$(tr -dc A-Za-z0-9 < /dev/urandom | head -c 100)"
NEW_USER_PASSWORD_SALT="$(tr -dc A-Za-z0-9 < /dev/urandom | head -c 16)"
echo $NEW_USER_PASSWORD > /data/newborn/password.txt
newborn_say "done."

echo
cd ansible
newborn_say "Running Ansible playbook..."
echo
# ansible all -m gather_facts
ansible-playbook --extra-vars "newborn_server_name_global=$SERVER_NAME newborn_user_name=$NEW_USER_NAME newborn_user_password=$NEW_USER_PASSWORD newborn_user_password_salt=$NEW_USER_PASSWORD_SALT newborn_user_sudo=$NEW_USER_SUDO newborn_oci_platform=$OCI_PLATFORM newborn_oci_compose=$OCI_COMPOSE" playbook.yml
echo
cd ..
newborn_say "done."

echo
newborn_say "Complete."

unset SERVER_NAME
unset NEW_USER_NAME
unset NEW_USER_PASSWORD
unset NEW_USER_PASSWORD_SALT
unset NEW_USER_SUDO
unset OCI_PLATFORM
unset OCI_COMPOSE
unset REST_ARGS
