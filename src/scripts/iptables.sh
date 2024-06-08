#!/bin/sh

CHAIN='NEWBORN'

remove_ufw () {
    # Remove UFW rules
    if command -v ufw >/dev/null 2>&1; then
        ufw --force reset
        ufw --force disable
    fi

    iptables -S | grep -e '-A' | grep -e ' ufw-' | sed -e 's/-A/-D/' | while read -r line; do iptables $line; done
    iptables -S | grep -e '-N' | grep -e ' ufw-' | sed -e 's/-N/-X/' | while read -r line; do iptables $line; done

    ip6tables -S | grep -e '-A' | grep -e ' ufw6-' | sed -e 's/-A/-D/' | while read -r line; do ip6tables $line; done
    ip6tables -S | grep -e '-N' | grep -e ' ufw6-' | sed -e 's/-N/-X/' | while read -r line; do ip6tables $line; done
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
    iptables  -A $CHAIN -s '169.254.0.0/16' -j RETURN
    iptables  -A $CHAIN -s '172.16.0.0/12'  -j RETURN
    iptables  -A $CHAIN -s '192.168.0.0/16' -j RETURN
    ip6tables -A $CHAIN -s 'fc00::/7'       -j RETURN
    ip6tables -A $CHAIN -s 'fe80::/10'      -j RETURN
    ip6tables -A $CHAIN -s 'ff00::/8'       -j RETURN

    # Allow IPv6 networking
    IPV6_SUBNETS=$(ip a | grep inet6 | awk '{print $2}')
    for ip in ${IPV6_SUBNETS}
    do
        ip6tables -A $CHAIN -s $ip -j RETURN
    done

    # Allow ssh connections
    SSH_PORT=$(grep '^Port ' /etc/ssh/sshd_config | cut -d ' ' -f2)
    call_iptables -A $CHAIN -p tcp --dport ${SSH_PORT:-22} -j ACCEPT

    # Allow access to ports 80 and 443 from Cloudflare
    CF_IP_RANGES_V4=$(curl -Ls https://www.cloudflare.com/ips-v4)
    for ip in ${CF_IP_RANGES_V4}
    do
        iptables -A $CHAIN -p tcp -s $ip --dport 80  -m comment --comment 'Cloudflare' -j RETURN
        iptables -A $CHAIN -p tcp -s $ip --dport 443 -m comment --comment 'Cloudflare' -j RETURN
    done
    CF_IP_RANGES_V6=$(curl -Ls https://www.cloudflare.com/ips-v6)
    for ip in ${CF_IP_RANGES_V6}
    do
        ip6tables -A $CHAIN -p tcp -s $ip --dport 80  -m comment --comment 'Cloudflare' -j RETURN
        ip6tables -A $CHAIN -p tcp -s $ip --dport 443 -m comment --comment 'Cloudflare' -j RETURN
    done

    # Drop all other traffic
    call_iptables -A $CHAIN -j DROP
}

link_chain () {
    # Link chain to the INPUT chain
    call_iptables -D INPUT -j $CHAIN
    call_iptables -I INPUT -j $CHAIN

    # Link chain to the FORWARD chain
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

apply () {
    create_chain
    link_chain
}

remove () {
    call_iptables -F $CHAIN
    call_iptables -D INPUT -j $CHAIN
    call_iptables -D FORWARD -j $CHAIN

    # Unlink chain from DOCKER-USER
    if iptables -L DOCKER-USER -n >/dev/null 2>&1; then
        iptables -D DOCKER-USER -j $CHAIN
    fi
    if ip6tables -L DOCKER-USER -n >/dev/null 2>&1; then
        ip6tables -D DOCKER-USER -j $CHAIN
    fi

    call_iptables -X $CHAIN
}

fix () {
    IPV4_CLOUDFLARE_LINES=$(iptables -S | grep -e '-A' | grep -e ' --comment Cloudflare -j ' | wc -l)
    IPV6_CLOUDFLARE_LINES=$(ip6tables -S | grep -e '-A' | grep -e ' --comment Cloudflare -j ' | wc -l)
    if [ $IPV4_CLOUDFLARE_LINES -eq 0 ] || [ $IPV6_CLOUDFLARE_LINES -eq 0 ]; then
        echo 'Cloudflare rules not found, reapplying...'
        apply
    else
        echo 'Cloudflare rules are here, nothing to fix.'
    fi
}

if [ "$1" = 'remove_ufw' ]; then
    remove_ufw
elif [ "$1" = 'remove' ]; then
    if [ -n "$2" ]; then
        CHAIN=$2
    fi

    remove
elif [ "$1" = 'fix' ]; then
    fix
elif [ "$1" = '' ]; then
    apply
else
    echo 'Unknown argument '$1
    exit 1
fi

# Save the rules
iptables-save > /etc/iptables/rules.v4
ip6tables-save > /etc/iptables/rules.v6
