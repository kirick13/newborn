
#!/bin/bash

newborn_say () {
	echo "[NEWBORN] $1"
}

echo

if [ -z "$NEWBORN_NEW_USER_PASSWORD" ]; then
	newborn_say 'Generating random password...'
	NEWBORN_NEW_USER_PASSWORD="$(tr -dc A-Za-z0-9 < /dev/urandom | head -c 100)"
fi
NEWBORN_NEW_USER_PASSWORD_SALT="$(tr -dc A-Za-z0-9 < /dev/urandom | head -c 16)"
echo $NEWBORN_NEW_USER_PASSWORD > output/password.txt

if [ -f input/ssh_key ]; then
	newborn_say 'Generating SSH public key from private key...'
	ssh-keygen -y \
	           -f input/ssh_key \
			   > input/ssh_key.pub
else
	newborn_say 'Generating SSH key pair...'
	ssh-keygen -t ecdsa \
	           -m PEM \
	           -b 521 \
	           -N '' \
	           -f input/ssh_key \
	           > /dev/null
	chmod 666 input/ssh_key
	cp input/ssh_key output/ssh_key
fi

newborn_say "Appending random SSH port to every host in inventory..."
cat input/inventory \
	| while read line
	do \
		SSH_PORT=$(python3 -c 'import random; print(random.randint(1025,65535))')
		echo ${line} | awk '{ print $1, "ansible_ssh_port='$SSH_PORT'" }' >> output/inventory
		echo ${line}" newborn_ssh_port=$SSH_PORT"
	done \
	| tee ansible/inventory > /dev/null

cd ansible

newborn_say "Running Ansible playbook..."
ansible-playbook --extra-vars "newborn_server_name_global=$NEWBORN_SERVER_NAME newborn_swap_size=$NEWBORN_SWAP_SIZE newborn_user_name=$NEWBORN_NEW_USER_NAME newborn_user_password=$NEWBORN_NEW_USER_PASSWORD newborn_user_password_salt=$NEWBORN_NEW_USER_PASSWORD_SALT newborn_user_sudo=$NEWBORN_NEW_USER_SUDO newborn_oci_platform=$NEWBORN_OCI_PLATFORM newborn_oci_compose=$NEWBORN_OCI_COMPOSE" playbook.yml
