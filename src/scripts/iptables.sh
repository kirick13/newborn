#!/bin/sh

CHAIN='NEWBORN'

remove_ufw () {
    # Remove UFW rules
    if command -v ufw >/dev/null 2>&1; then
        ufw --force reset
        ufw --force disable
    fi

    iptables -S | grep -e '-A' | grep -e ' ufw-' | sed -e 's/-A/-D/' | while read -r line; do sudo iptables $line; done
    iptables -S | grep -e '-N' | grep -e ' ufw-' | sed -e 's/-N/-X/' | while read -r line; do sudo iptables $line; done
}

call_iptables () {
    iptables $@
    ip6tables $@
}

create_chain () {
    # Create chain
    call_iptables -N $CHAIN

    # Flush the chain
    call_iptables -F $CHAIN

    # Allow established and related connections
    call_iptables -A $CHAIN -m conntrack --ctstate ESTABLISHED,RELATED -j RETURN

    # Allow all access on localhost
    call_iptables -A $CHAIN -i lo -j RETURN

    # Allow all access on local networks
    iptables  -A $CHAIN -s '10.0.0.0/8'     -j RETURN
    iptables  -A $CHAIN -s '172.16.0.0/12'  -j RETURN
    iptables  -A $CHAIN -s '192.168.0.0/16' -j RETURN
    ip6tables -A $CHAIN -s 'fc00::/7'       -j RETURN

    # Allow ssh connections
    SSH_PORT=$(grep '^Port ' /etc/ssh/sshd_config | cut -d ' ' -f2)
    call_iptables -A $CHAIN -p tcp --dport ${SSH_PORT:-22} -j ACCEPT

    # Allow access to ports 80 and 443 from Cloudflare
    CF_IP_RANGES_V4=$(curl -s https://www.cloudflare.com/ips-v4)
    for ip in ${CF_IP_RANGES_V4}
    do
        iptables -A $CHAIN -p tcp -s $ip --dport 80  -j RETURN
        iptables -A $CHAIN -p tcp -s $ip --dport 443 -j RETURN
    done

    CF_IP_RANGES_V6=$(curl -s https://www.cloudflare.com/ips-v6)
    for ip in ${CF_IP_RANGES_V6}
    do
        ip6tables -A $CHAIN -p tcp -s $ip --dport 80  -j RETURN
        ip6tables -A $CHAIN -p tcp -s $ip --dport 443 -j RETURN
    done

    # Drop all other traffic
    call_iptables -A $CHAIN -j DROP
}

link_chain () {
    # Link chain to the INPUT chain
    call_iptables -D INPUT -j $CHAIN
    call_iptables -I INPUT -j $CHAIN
    call_iptables -D FORWARD -j $CHAIN
    call_iptables -I FORWARD -j $CHAIN

    # Link chain to the DOCKER-USER
    if iptables -L DOCKER-USER -n >/dev/null 2>&1; then
        iptables -D DOCKER-USER -j $CHAIN
        iptables -I DOCKER-USER -j $CHAIN
    fi
    if ip6tables -L DOCKER-USER -n >/dev/null 2>&1; then
        ip6tables -D DOCKER-USER -j $CHAIN
        ip6tables -I DOCKER-USER -j $CHAIN
    fi
}

remove () {
    call_iptables -F $CHAIN
    call_iptables -D INPUT -j $CHAIN

    # Unlink chain from DOCKER-USER
    if iptables -L DOCKER-USER -n >/dev/null 2>&1; then
        iptables -D DOCKER-USER -j $CHAIN
    fi
    if ip6tables -L DOCKER-USER -n >/dev/null 2>&1; then
        ip6tables -D DOCKER-USER -j $CHAIN
    fi

    call_iptables -X $CHAIN
}

if [ "$1" = 'remove_ufw' ]; then
    remove_ufw
elif [ "$1" = 'remove' ]; then
    if [ -n "$2" ]; then
        CHAIN=$2
    fi

    remove
elif [ "$1" = '' ]; then
    create_chain
    link_chain
else
    echo 'Unknown argument '$1
    exit 1
fi

# Save the rules
iptables-save > /etc/iptables/rules.v4
ip6tables-save > /etc/iptables/rules.v6
