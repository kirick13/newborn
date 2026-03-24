#!/bin/bash

SWAP_SIZE=$1
PATH_SWAP_FILE_NEW='/swapfile'

# [TASK] Delete current swap
# Turn off swap
swapoff -a
# Get swap file paths, even commented out ones
SWAP_FILES=$(grep swap /etc/fstab | sed 's/^\s*#*//g' | awk '{print $1}')
for SWAP_FILE in $SWAP_FILES; do
	# Check if the swap file exists
	if [ -e "$SWAP_FILE" ]; then
		rm -f $SWAP_FILE
		echo "Swap file $SWAP_FILE has been removed"
	else
		echo "Swap file $SWAP_FILE not found"
	fi

	# Remove the swap file from /etc/fstab
	sed -i.bak "\|$SWAP_FILE|d" /etc/fstab
done

# [TASK] Create new swap
# if swap_size exists, create a new swap file
if [ -n "$SWAP_SIZE" ]; then
	if [ -f "$PATH_SWAP_FILE_NEW" ]; then
		rm $PATH_SWAP_FILE_NEW
	fi

	fallocate -l $SWAP_SIZE $PATH_SWAP_FILE_NEW
	chmod 600 $PATH_SWAP_FILE_NEW
	mkswap $PATH_SWAP_FILE_NEW
	swapon $PATH_SWAP_FILE_NEW

	# Add the swap file to /etc/fstab
	echo "$PATH_SWAP_FILE_NEW none swap sw 0 0" >> /etc/fstab

	# Set swappiness and cache pressure
	PATH_TMP_SYSCTL=$(mktemp)
	sysctl vm.swappiness=10
	sysctl vm.vfs_cache_pressure=50
	cat /etc/sysctl.conf | awk '!/^#/' | awk NF | awk '!/^vm.swappiness/' | awk '!/^vm.vfs_cache_pressure/' > $PATH_TMP_SYSCTL
	echo 'vm.swappiness=10' >> $PATH_TMP_SYSCTL
	echo 'vm.vfs_cache_pressure=50' >> $PATH_TMP_SYSCTL
	mv $PATH_TMP_SYSCTL /etc/sysctl.conf
fi
