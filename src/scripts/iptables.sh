#!/bin/sh

CHAIN='NEWBORN_IMMUNITY'

# 1. Create chains
iptables  -N $CHAIN
ip6tables -N $CHAIN

# 2. Flush chains
iptables  -F $CHAIN
ip6tables -F $CHAIN

# 3. Link the chain to the INPUT and FORWARD chains
# use -I to insert rule at the top of the chain
# otherwise Docker containers will not be affected by these rules
iptables -I INPUT   -j $CHAIN
iptables -I FORWARD -j $CHAIN

# 4. Add rules to the chain
# 4.1. Allow all access on localhost
iptables  -A $CHAIN -i lo -j ACCEPT
ip6tables -A $CHAIN -i lo -j ACCEPT

# 4.2. Allow all access on local networks
iptables  -A $CHAIN -s '10.0.0.0/8'     -j ACCEPT
iptables  -A $CHAIN -s '172.16.0.0/12'  -j ACCEPT
iptables  -A $CHAIN -s '192.168.0.0/16' -j ACCEPT
ip6tables -A $CHAIN -s 'fc00::/7'       -j ACCEPT

# 4.3. Allow ssh connections
SSH_PORT=$(grep '^Port ' /etc/ssh/sshd_config | cut -d ' ' -f2)
iptables  -A $CHAIN -p tcp --dport ${SSH_PORT:-22} -j ACCEPT
ip6tables -A $CHAIN -p tcp --dport ${SSH_PORT:-22} -j ACCEPT

# 4.4. Allow access to ports 80 and 443 from Cloudflare
CF_IP_RANGES_V4=$(curl -s https://www.cloudflare.com/ips-v4)
for ip in ${CF_IP_RANGES_V4}
do
    iptables -A $CHAIN -p tcp -s $ip --dport 80  -j ACCEPT
    iptables -A $CHAIN -p tcp -s $ip --dport 443 -j ACCEPT
done

CF_IP_RANGES_V6=$(curl -s https://www.cloudflare.com/ips-v6)
for ip in ${CF_IP_RANGES_V6}
do
    ip6tables -A $CHAIN -p tcp -s $ip --dport 80  -j ACCEPT
    ip6tables -A $CHAIN -p tcp -s $ip --dport 443 -j ACCEPT
done

# 4.5. Drop all other traffic
iptables  -A $CHAIN -j DROP
ip6tables -A $CHAIN -j DROP

# 5. Save the rules
iptables-save > /etc/iptables/rules.v4
ip6tables-save > /etc/iptables/rules.v6
