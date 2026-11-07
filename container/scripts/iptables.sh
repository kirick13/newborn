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

# Append a rule to the files that apply loads with iptables-restore
add_rule () {
    add_rule4 "$@"
    add_rule6 "$@"
}

add_rule4 () {
    echo "-A $CHAIN $*" >> "$RULES4"
}

add_rule6 () {
    echo "-A $CHAIN $*" >> "$RULES6"
}

# Allow access to ports 80 and 443 from anywhere, from Cloudflare, or nowhere (default).
# Ports are matched before Docker's DNAT rewrites them to container ports.
allow_http () {
    case "$1" in
        anywhere)
            add_rule -p tcp -m conntrack --ctorigdstport 80  --ctdir ORIGINAL -j RETURN
            add_rule -p tcp -m conntrack --ctorigdstport 443 --ctdir ORIGINAL -j RETURN
            ;;
        cloudflare)
            if CF_IP_RANGES_V4=$(curl -fsSL --max-time 30 https://www.cloudflare.com/ips-v4) &&
                CF_IP_RANGES_V6=$(curl -fsSL --max-time 30 https://www.cloudflare.com/ips-v6); then
                for ip in ${CF_IP_RANGES_V4}
                do
                    add_rule4 -s $ip -p tcp -m conntrack --ctorigdstport 80  --ctdir ORIGINAL -m comment --comment Cloudflare -j RETURN
                    add_rule4 -s $ip -p tcp -m conntrack --ctorigdstport 443 --ctdir ORIGINAL -m comment --comment Cloudflare -j RETURN
                done
                for ip in ${CF_IP_RANGES_V6}
                do
                    add_rule6 -s $ip -p tcp -m conntrack --ctorigdstport 80  --ctdir ORIGINAL -m comment --comment Cloudflare -j RETURN
                    add_rule6 -s $ip -p tcp -m conntrack --ctorigdstport 443 --ctdir ORIGINAL -m comment --comment Cloudflare -j RETURN
                done
            else
                # Reapply the loaded ranges so other rules, like a new SSH port, still update
                iptables  -S $CHAIN 2>/dev/null | grep -e '--comment Cloudflare' >> "$RULES4"
                ip6tables -S $CHAIN 2>/dev/null | grep -e '--comment Cloudflare' >> "$RULES6"
                # Nothing is loaded on the first run: keep the current rules
                grep -q Cloudflare "$RULES4" && grep -q Cloudflare "$RULES6" || exit 1
            fi
            ;;
        *)
            ;;
    esac
}

create_chain () {
    # Declaring the chain makes iptables-restore empty it in the same transaction
    printf '*filter\n:%s - [0:0]\n' $CHAIN | tee "$RULES4" > "$RULES6"

    # Allow established and related connections
    add_rule -m conntrack --ctstate ESTABLISHED,RELATED -j RETURN

    # Allow all access on localhost
    add_rule -i lo -j RETURN

    # Allow all access on local networks
    add_rule4 -s 10.0.0.0/8     -j RETURN
    add_rule4 -s 169.254.0.0/16 -j RETURN
    add_rule4 -s 172.16.0.0/12  -j RETURN
    add_rule4 -s 192.168.0.0/16 -j RETURN
    add_rule6 -s fc00::/7       -j RETURN
    add_rule6 -s fe80::/10      -j RETURN

    # Allow ICMPv6, IPv6 needs it for neighbor discovery
    add_rule6 -p ipv6-icmp -j RETURN

    # Allow ssh connections
    SSH_PORT=$(grep '^Port ' /etc/ssh/sshd_config | cut -d ' ' -f2)
    add_rule -p tcp -m conntrack --ctorigdstport ${SSH_PORT:-22} --ctdir ORIGINAL -j ACCEPT

    # The playbook sets this argument from newborn_firewall_http
    allow_http 'nowhere'

    # Drop all other traffic
    add_rule -j DROP

    echo 'COMMIT' | tee -a "$RULES4" >> "$RULES6"
}

# Link chain from $2 using $1 (iptables or ip6tables), never leaving it unlinked
link () {
    $1 -C $2 -j $CHAIN >/dev/null 2>&1 || $1 -I $2 -j $CHAIN
}

link_chain () {
    # Link chain to the INPUT chain
    link iptables  INPUT
    link ip6tables INPUT

    # Link chain to the FORWARD chain
    link iptables  FORWARD
    link ip6tables FORWARD

    # Link chain to the DOCKER-USER
    if iptables -L DOCKER-USER -n >/dev/null 2>&1; then
        link iptables DOCKER-USER
    fi
    if ip6tables -L DOCKER-USER -n >/dev/null 2>&1; then
        link ip6tables DOCKER-USER
    fi
}

apply () {
    RULES4=$(mktemp) && RULES6=$(mktemp) || exit 1
    trap 'rm -f "$RULES4" "$RULES6"' EXIT

    create_chain

    # Each restore is atomic: on error nothing changes and the rules are not saved
    iptables-restore --noflush "$RULES4" || exit 1
    ip6tables-restore --noflush "$RULES6" || exit 1

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

if [ "$1" = 'remove_ufw' ]; then
    remove_ufw
elif [ "$1" = 'remove' ]; then
    if [ -n "$2" ]; then
        CHAIN=$2
    fi

    remove
elif [ "$1" = '' ]; then
    apply
else
    echo 'Unknown argument '$1
    exit 1
fi

# Save the rules
iptables-save > /etc/iptables/rules.v4
ip6tables-save > /etc/iptables/rules.v6
