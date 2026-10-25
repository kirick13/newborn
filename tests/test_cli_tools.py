"""Exercise CLI installer error handling without downloading or installing tools."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest

import yaml


ROOT = Path(__file__).resolve().parents[1]
PLAYBOOK = yaml.safe_load((ROOT / "container/playbook.yaml").read_text())[0]
INSTALL_SCRIPT = next(
    task["shell"] for task in PLAYBOOK["tasks"]
    if task["name"] == "Setup | Install CLI tools"
)

# Stub the external commands; run the real playbook shell body with /bin/sh.
STUBS = """
mise() {
    if [ "$1" = use ]; then
        printf 'download\\n' >> "$CALLS"
        [ "$FAIL_STEP" != download ]
    else
        printf 'lookup %s\\n' "$2" >> "$CALLS"
        # Output a plausible path even on failure to catch masked lookup errors.
        printf '/fake/%s\\n' "$2"
        [ "$FAIL_STEP" != "lookup:$2" ]
    fi
}
install() {
    printf 'copy %s\\n' "$4" >> "$CALLS"
    [ "$FAIL_STEP" != "copy:$4" ]
}
"""


class CliToolsTests(unittest.TestCase):
    def run_installer(self, fail_step=""):
        with tempfile.TemporaryDirectory() as temporary:
            calls = Path(temporary) / "calls"
            result = subprocess.run(
                ["/bin/sh", "-c", STUBS + INSTALL_SCRIPT],
                env={**os.environ, "CALLS": str(calls), "FAIL_STEP": fail_step},
                capture_output=True, text=True, timeout=10,
            )
            return result, calls.read_text().splitlines()

    def test_download_failure_stops_before_copying(self):
        result, calls = self.run_installer("download")
        self.assertNotEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["download"])

    def test_lookup_failure_stops_even_if_it_prints_a_path(self):
        result, calls = self.run_installer("lookup:bat")
        self.assertNotEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["download", "lookup bat"])

    def test_copy_failure_stops_before_later_tools(self):
        result, calls = self.run_installer("copy:/usr/local/bin/bat")
        self.assertNotEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["download", "lookup bat", "copy /usr/local/bin/bat"])

    def test_success_copies_all_six_tools(self):
        result, calls = self.run_installer()
        self.assertEqual(result.returncode, 0, result.stderr)
        expected = ["download"]
        for binary in ("bat", "btm", "dust", "eza", "fd", "rg"):
            expected.extend([f"lookup {binary}", f"copy /usr/local/bin/{binary}"])
        self.assertEqual(calls, expected)


if __name__ == "__main__":
    unittest.main(verbosity=2)
