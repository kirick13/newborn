#!/bin/bash

swapoff -a

SWAP_SIZE=$1
FSTAB_TMP_PATH='/tmp/fstab.tmp'
SYSCTL_CONF_TMP_FILE='/tmp/sysctl.conf.tmp'

if [ -f /swapfile ]; then
	rm /swapfile
fi
fallocate -l $SWAP_SIZE /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile

SWAP_DEVICE=$(cat /etc/fstab | awk '{print $1,$3}' | awk '/ swap$/' | awk '{print $1}')
if [ -n "$swap_device" ]; then
	echo "Swap device is $swap_device"
	cat /etc/fstab | awk -vD="$swap_device" '!index($0,D)' > $FSTAB_TMP_PATH
else
	echo "Swap device not found"
	cp /etc/fstab $FSTAB_TMP_PATH
fi
echo '/swapfile none swap sw 0 0' >> $FSTAB_TMP_PATH
# mv /etc/fstab /etc/fstab.old
mv $FSTAB_TMP_PATH /etc/fstab

sysctl vm.swappiness=10
sysctl vm.vfs_cache_pressure=50
cat /etc/sysctl.conf | awk '!/^#/' | awk NF | awk '!/^vm.swappiness/' | awk '!/^vm.vfs_cache_pressure/' > $SYSCTL_CONF_TMP_FILE
echo 'vm.swappiness=10' >> $SYSCTL_CONF_TMP_FILE
echo 'vm.vfs_cache_pressure=50' >> $SYSCTL_CONF_TMP_FILE
# mv /etc/sysctl.conf /etc/sysctl.conf.old
mv $SYSCTL_CONF_TMP_FILE /etc/sysctl.conf

rm $FSTAB_TMP_PATH
rm $SYSCTL_CONF_TMP_FILE

unset SWAP_SIZE
unset SWAP_DEVICE
unset FSTAB_TMP_PATH
unset SYSCTL_CONF_TMP_FILE
