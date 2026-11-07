"""Exercise the firewall script with stubbed iptables and curl commands."""

import os
from pathlib import Path
import re
import subprocess
import tempfile
import unittest

import yaml
from ansible.parsing.dataloader import DataLoader
from ansible.template import Templar


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = (ROOT / "container/scripts/iptables.sh").read_text()
PLAYBOOK = yaml.safe_load((ROOT / "container/playbook.yaml").read_text())[0]

# One stub, installed under every external command name the script runs.
STUB = """#!/bin/sh
name=$(basename "$0")
echo "$name $*" >> "$CALLS"
[ "$FAIL" != "$name" ] || exit 1
case "$name" in
    curl) case "$*" in *ips-v4*) echo 173.245.48.0/20 ;; *) echo 2400:cb00::/32 ;; esac ;;
    *-restore) cp "$2" "$OUT/$name" ;;
    *-save) echo "saved by $name" ;;
    iptables|ip6tables)
        [ "$1" = -S ] && [ -n "$LOADED" ] && echo "-A NEWBORN -s loaded-by-$name -m comment --comment Cloudflare -j RETURN"
        [ "$1" != -C ] && [ "$1" != -L ] ;;
esac
"""
COMMANDS = ["curl", "iptables", "ip6tables", "iptables-restore", "ip6tables-restore",
            "iptables-save", "ip6tables-save"]
SSH_RULE = "-A NEWBORN -p tcp -m conntrack --ctorigdstport 2222 --ctdir ORIGINAL -j ACCEPT"


def set_http_access(script, mode):
    """Apply the playbook's lineinfile task (backrefs: true) like Ansible would."""
    task = next(task["lineinfile"] for task in PLAYBOOK["tasks"]
                if task["name"] == "Firewall | Set HTTP access in script")
    line = Templar(loader=DataLoader(), variables={"newborn_firewall_http": mode}).template(task["line"])
    pattern = re.compile(task["regexp"])
    lines = script.splitlines(keepends=True)
    matches = [index for index, text in enumerate(lines) if pattern.search(text)]
    assert len(matches) == 1, matches
    lines[matches[0]] = pattern.search(lines[matches[0]]).expand(line) + "\n"
    return "".join(lines)


class IptablesTests(unittest.TestCase):
    def run_script(self, mode, fail="", loaded=False):
        with tempfile.TemporaryDirectory() as temporary:
            temporary = Path(temporary)
            for name in ("bin", "out", "tmp", "iptables"):
                (temporary / name).mkdir()
            for command in COMMANDS:
                stub = temporary / "bin" / command
                stub.write_text(STUB)
                stub.chmod(0o755)
            (temporary / "sshd_config").write_text("Port 2222\n")
            script = SCRIPT.replace("/etc/ssh/sshd_config", str(temporary / "sshd_config"))
            script = script.replace("/etc/iptables/", f"{temporary}/iptables/")
            script = set_http_access(script, mode)
            result = subprocess.run(
                ["/bin/sh", "-c", script], capture_output=True, text=True, timeout=10,
                env={**os.environ, "PATH": f"{temporary}/bin:{os.environ['PATH']}",
                     "CALLS": str(temporary / "calls"), "OUT": str(temporary / "out"),
                     "TMPDIR": str(temporary / "tmp"), "FAIL": fail,
                     "LOADED": "y" if loaded else ""},
            )
            calls = (temporary / "calls").read_text().splitlines()
            rules = {path.name: path.read_text().splitlines() for path in (temporary / "out").iterdir()}
            saved = sorted(path.name for path in (temporary / "iptables").iterdir() if path.stat().st_size)
            self.assertEqual(list((temporary / "tmp").iterdir()), [], "rule files were not removed")
            return result, calls, rules, saved

    def test_cloudflare_rules_are_restored_before_linking_and_saved(self):
        result, calls, rules, saved = self.run_script("cloudflare")
        self.assertEqual(result.returncode, 0, result.stderr)
        for name, cloudflare_range in (("iptables-restore", "173.245.48.0/20"),
                                       ("ip6tables-restore", "2400:cb00::/32")):
            lines = rules[name]
            self.assertEqual(lines[:2], ["*filter", ":NEWBORN - [0:0]"])
            self.assertEqual(lines[-2:], ["-A NEWBORN -j DROP", "COMMIT"])
            self.assertIn(SSH_RULE, lines)
            for port in (80, 443):
                self.assertIn(f"-A NEWBORN -s {cloudflare_range} -p tcp -m conntrack --ctorigdstport {port} "
                              "--ctdir ORIGINAL -m comment --comment Cloudflare -j RETURN", lines)
            self.assertNotIn("--dport", "\n".join(lines))
        self.assertIn("-A NEWBORN -p ipv6-icmp -j RETURN", rules["ip6tables-restore"])
        names = [call.split()[0] for call in calls]
        self.assertLess(names.index("ip6tables-restore"), calls.index("iptables -I INPUT -j NEWBORN"))
        self.assertEqual(saved, ["rules.v4", "rules.v6"])

    def test_anywhere_and_nowhere_do_not_download_ranges(self):
        for mode, opened in (("anywhere", True), ("nowhere", False)):
            with self.subTest(mode=mode):
                result, calls, rules, _ = self.run_script(mode)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertFalse([call for call in calls if call.startswith("curl")])
                https_rule = "-A NEWBORN -p tcp -m conntrack --ctorigdstport 443 --ctdir ORIGINAL -j RETURN"
                self.assertEqual(https_rule in rules["ip6tables-restore"], opened)

    def test_failed_download_reapplies_loaded_ranges_with_new_ssh_port(self):
        result, _, rules, saved = self.run_script("cloudflare", fail="curl", loaded=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        for name, command in (("iptables-restore", "iptables"), ("ip6tables-restore", "ip6tables")):
            self.assertIn(f"-A NEWBORN -s loaded-by-{command} -m comment --comment Cloudflare -j RETURN", rules[name])
            self.assertIn(SSH_RULE, rules[name])
        self.assertEqual(saved, ["rules.v4", "rules.v6"])

    def test_failed_download_without_loaded_ranges_keeps_current_rules(self):
        result, calls, rules, saved = self.run_script("cloudflare", fail="curl")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual([call.split()[0] for call in calls], ["curl", "iptables", "ip6tables"])
        self.assertEqual((rules, saved), ({}, []))

    def test_new_ssh_port_is_allowed_before_sshd_restarts(self):
        names = [task["name"] for task in PLAYBOOK["tasks"]]
        self.assertLess(names.index("OpenSSH | Execute the iptables script to reflect changes in SSH port"),
                        names.index("OpenSSH | Restart server"))

    def test_failed_restore_is_neither_linked_nor_saved(self):
        result, calls, _, saved = self.run_script("nowhere", fail="iptables-restore")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual([call.split()[0] for call in calls], ["iptables-restore"])
        self.assertEqual(saved, [])


if __name__ == "__main__":
    unittest.main(verbosity=2)
