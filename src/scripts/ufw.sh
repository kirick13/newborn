#!/bin/sh

# ChatGPT made it. Human supervised.

# First, reset UFW rules
ufw --force reset

# Default rules
ufw default deny incoming
ufw default allow outgoing

# Allow all access on localhost
ufw allow from 127.0.0.0/8
ufw allow from ::1/128

# Allow all access on local networks
ufw allow from 10.0.0.0/8
ufw allow from 172.16.0.0/12
ufw allow from 192.168.0.0/16
ufw allow from fc00::/7

# Allow ssh connections
SSH_PORT=$(grep '^Port ' /etc/ssh/sshd_config | cut -d ' ' -f2)
sudo ufw allow ${SSH_PORT:-22}/tcp

# Fetch the current list of Cloudflare IP ranges
CF_IP_RANGES_V4=$(curl -s https://www.cloudflare.com/ips-v4)
CF_IP_RANGES_V6=$(curl -s https://www.cloudflare.com/ips-v6)
# Now iterate over the Cloudflare IP ranges and allow incoming connections on ports 80 and 443
for ip in ${CF_IP_RANGES_V4} ${CF_IP_RANGES_V6}
do
    ufw allow proto tcp from $ip to any port 80,443
done

# Enable UFW
ufw --force enable
