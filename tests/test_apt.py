"""Offline regression checks using the project's Ansible runner image."""

import copy
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import os
from pathlib import Path
import shlex
import socket
import subprocess
import sys
import tempfile
import threading
import time
import unittest

import yaml
from ansible.parsing.dataloader import DataLoader
from ansible.template import Templar


ROOT = Path(__file__).resolve().parents[1]
CONTAINER = ROOT / "container"
PLAYBOOK = yaml.safe_load((CONTAINER / "playbook.yaml").read_text())[0]
sys.path.insert(0, str(CONTAINER / "scripts"))
import apt_sources


class MirrorHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        with self.server.requests_lock:
            self.server.requests.append(self.path)
        status = 200
        if self.path.startswith("/parallel-"):
            # Both requests must arrive before either can finish successfully.
            try:
                self.server.barrier.wait(timeout=10)
            except threading.BrokenBarrierError:
                status = 503
        elif self.path == "/slow":
            time.sleep(18)
        elif self.path == "/disconnect":
            self.connection.close()
            return
        elif self.path == "/missing":
            status = 404
        elif self.path == "/error":
            status = 503
        elif self.path == "/empty":
            status = 204
        elif self.path == "/redirect":
            status = 302
        self.send_response(status)
        if status == 302:
            self.send_header("Location", "/ok")
        self.send_header("Content-Length", "0")
        self.end_headers()

    def log_message(self, *_args):
        pass


class AptTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = ThreadingHTTPServer(("127.0.0.1", 0), MirrorHandler)
        cls.server.requests = []
        cls.server.requests_lock = threading.Lock()
        cls.server.barrier = threading.Barrier(2)
        cls.server_thread = threading.Thread(target=cls.server.serve_forever, daemon=True)
        cls.server_thread.start()
        cls.mirror = f"http://127.0.0.1:{cls.server.server_port}"

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()
        cls.server.server_close()
        cls.server_thread.join()

    def setUp(self):
        self.server.requests.clear()

    def run_play(self, directory, tasks, variables):
        play = [{
            "hosts": "all",
            "gather_facts": False,
            "vars": {**PLAYBOOK.get("vars", {}), "ansible_python_interpreter": sys.executable, **variables},
            "tasks": tasks,
        }]
        path = directory / "test.yaml"
        path.write_text(yaml.safe_dump(play, sort_keys=False))
        return subprocess.run(
            ["ansible-playbook", "-i", "localhost,", "-c", "local", str(path)],
            env={**os.environ, "ANSIBLE_NOCOLOR": "1"},
            capture_output=True, text=True, timeout=180,
        )

    def source_task(self, apt_root, release="noble", arch="x86_64"):
        task = copy.deepcopy(next(
            task for task in PLAYBOOK["tasks"]
            if task["name"] == "Setup | Check and repair Ubuntu repositories"
        ))
        task["script"]["cmd"] = task["script"]["cmd"].replace(
            "scripts/apt_sources.py", str(CONTAINER / "scripts/apt_sources.py"), 1
        ) + " --apt-root {{ newborn_apt_root | quote }}"
        task["vars"] = {
            "newborn_apt_root": str(apt_root),
            "ansible_distribution_release": release,
            "ansible_architecture": arch,
        }
        return task

    def write_sources(self, apt_root, sources):
        for name, content in sources.items():
            path = apt_root / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(content)

    def assert_files_unchanged(self, apt_root, originals):
        for name, content in originals.items():
            self.assertEqual((apt_root / name).read_text(), content)

    def assert_no_extra_files(self, apt_root, originals):
        self.assertEqual(
            {str(path.relative_to(apt_root)) for path in apt_root.rglob("*") if path.is_file()},
            set(originals),
        )

    def test_platform_validation_rejects_unsupported_hosts_before_changes(self):
        validation = PLAYBOOK["tasks"][0]
        cases = [
            ("Ubuntu", "22", "jammy", "x86_64", True),
            ("Ubuntu", "26", "resolute", "aarch64", True),
            ("Ubuntu", "24", "noble", "armv7l", False),
            ("Ubuntu", "24", "noble", "ppc64le", False),
            ("Ubuntu", "24", "noble", "i386", False),
            ("Ubuntu", "24", "noble", "riscv64", False),
            ("Ubuntu", "24", "noble", "s390x", False),
            ("Debian", "24", "noble", "x86_64", False),
            ("Ubuntu", "20", "focal", "x86_64", False),
            ("Ubuntu", "28", "future", "x86_64", False),
            ("Ubuntu", "24", "noble", "unknown", False),
        ]
        with tempfile.TemporaryDirectory() as temporary:
            for distro, version, release, arch, accepted in cases:
                with self.subTest(distro=distro, version=version, arch=arch):
                    result = self.run_play(Path(temporary), [validation], {
                        "ansible_distribution": distro,
                        "ansible_distribution_major_version": version,
                        "ansible_distribution_release": release,
                        "ansible_architecture": arch,
                    })
                    self.assertEqual(result.returncode == 0, accepted, result.stdout + result.stderr)

    def test_ubuntu_and_docker_sources_use_mapped_package_architecture(self):
        sources = next(task for task in PLAYBOOK["tasks"] if task["name"] == "Setup | Check and repair Ubuntu repositories")
        docker = next(task for task in PLAYBOOK["tasks"] if task["name"] == "Docker | Add repository")
        cases = {
            "x86_64": "amd64", "aarch64": "arm64",
        }
        for arch, package_arch in cases.items():
            with self.subTest(arch=arch):
                templar = Templar(loader=DataLoader(), variables={
                    **PLAYBOOK["vars"], "ansible_architecture": arch,
                    "ansible_distribution_release": "noble",
                })
                command = shlex.split(templar.template(sources["script"]["cmd"]))
                self.assertEqual(command[command.index("--architecture") + 1], package_arch)
                repo = templar.template(docker["apt_repository"]["repo"])
                self.assertIn(f"[arch={package_arch} ", repo)

    def test_compose_asset_names_use_mapped_ansible_architecture(self):
        compose = next(task for task in PLAYBOOK["tasks"] if task["name"] == "Docker | Install Compose")
        version = PLAYBOOK["vars"]["newborn_compose_version"]
        for arch, asset in {"x86_64": "x86_64", "aarch64": "aarch64"}.items():
            with self.subTest(arch=arch):
                templar = Templar(loader=DataLoader(), variables={
                    **PLAYBOOK["vars"], "ansible_architecture": arch,
                })
                url = templar.template(compose["get_url"]["url"])
                self.assertEqual(url, f"https://github.com/docker/compose/releases/download/{version}/docker-compose-linux-{asset}")

    def test_compose_rejects_missing_binary_before_changes(self):
        validation = next(task for task in PLAYBOOK["tasks"] if task["name"] == "Validate Docker Compose architecture")
        with tempfile.TemporaryDirectory() as temporary:
            result = self.run_play(Path(temporary), [validation], {
                "ansible_architecture": "i386",
                "newborn_oci_runtime": "docker", "newborn_oci_compose": "y",
            })
            self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertIn("has no binary for i386", result.stdout)

    def test_healthy_sources_are_preserved_and_checked_concurrently(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            apt_root = directory / "apt"
            originals = {
                "sources.list": (
                    f"# deb {self.mirror}/error noble main\n"
                    f"deb [arch=amd64 signed-by=/key.gpg] {self.mirror}/parallel-one noble main # comment\n"
                    f"deb-src {self.mirror}/parallel-one noble main\n"
                ),
                "sources.list.d/ubuntu.sources": (
                    f"Types: deb\nURIs: {self.mirror}/parallel-two\n"
                    f" {self.mirror}/redirect {self.mirror}/Case {self.mirror}/case\n"
                    f" {self.mirror}/empty\n"
                    "Suites: noble noble-updates\nComponents: main\n"
                    f"Signed-By: {self.mirror}/ignored-key\n\n"
                    f"# URIs: {self.mirror}/ignored-comment\n"
                    f"Types: deb\nURIs: {self.mirror}/disabled\nEnabled:\n no\n"
                ),
                "sources.list.d/docker.list": f"deb {self.mirror}/ignored-docker noble stable\n",
            }
            self.write_sources(apt_root, originals)
            result = self.run_play(directory, [self.source_task(apt_root)], {})
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assert_files_unchanged(apt_root, originals)
            self.assert_no_extra_files(apt_root, originals)
            self.assertIn("changed=0", result.stdout)
            self.assertCountEqual(self.server.requests, ["/parallel-one", "/parallel-two", "/redirect", "/ok", "/Case", "/case", "/empty"])

    def test_only_failing_files_are_replaced_without_backups(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            cases = []
            tasks = []
            for release in ("jammy", "noble", "resolute"):
                for arch in ("x86_64", "aarch64"):
                    apt_root = directory / f"{release}-{arch}"
                    failing = "sources.list" if release == "jammy" else "sources.list.d/ubuntu.sources"
                    legacy_url = "/missing" if failing == "sources.list" else "/ok"
                    deb822_url = "/error" if failing.endswith(".sources") else "/ok"
                    originals = {
                        "sources.list": f"deb {self.mirror}{legacy_url} {release} main\n",
                        "sources.list.d/provider.list": "deb http://untouched.invalid noble main\n",
                        "sources.list.d/ubuntu.sources": (
                            f"Types: deb\nURIs: {self.mirror}{deb822_url} {self.mirror}/other\n"
                            f"Suites: {release}\nComponents: main\n"
                        ),
                    }
                    self.write_sources(apt_root, originals)
                    cases.append((apt_root, release, arch, originals, failing))
                    tasks.append(self.source_task(apt_root, release, arch))
            result = self.run_play(directory, tasks, {})
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertIn("changed=6", result.stdout)
            for apt_root, release, arch, originals, failing in cases:
                with self.subTest(release=release, arch=arch):
                    self.assert_no_extra_files(apt_root, originals)
                    self.assert_files_unchanged(apt_root, {name: text for name, text in originals.items() if name != failing})
                    sources = (apt_root / failing).read_text()
                    self.assertNotIn(self.mirror, sources)
                    for suite in (release, f"{release}-updates", f"{release}-backports", f"{release}-security"):
                        self.assertIn(suite, sources)
                    self.assertIn("main restricted universe multiverse", sources)
                    if arch == "x86_64":
                        self.assertIn("http://archive.ubuntu.com/ubuntu", sources)
                        self.assertIn("http://security.ubuntu.com/ubuntu", sources)
                        self.assertNotIn("ports.ubuntu.com", sources)
                    else:
                        self.assertIn("http://ports.ubuntu.com/ubuntu-ports", sources)
                        self.assertNotIn("archive.ubuntu.com", sources)
            self.assertEqual(self.server.requests.count("/other"), len(cases))

    def test_timeouts_and_connection_errors_trigger_replacement(self):
        with tempfile.TemporaryDirectory() as temporary, socket.socket() as refused:
            # A bound socket without listen() reliably refuses connections.
            refused.bind(("127.0.0.1", 0))
            directory = Path(temporary)
            apt_root = directory / "apt"
            originals = {
                "sources.list": f"deb {self.mirror}/slow noble main\n",
                "sources.list.d/ubuntu.sources": (
                    f"Types: deb\nURIs: http://127.0.0.1:{refused.getsockname()[1]}/ubuntu\n"
                    f" {self.mirror}/disconnect\nSuites: noble\nComponents: main\n"
                ),
            }
            self.write_sources(apt_root, originals)
            result = self.run_play(directory, [self.source_task(apt_root)], {})
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            for name in originals:
                self.assertIn("http://archive.ubuntu.com/ubuntu", (apt_root / name).read_text())
            self.assert_no_extra_files(apt_root, originals)
            self.assertCountEqual(self.server.requests, ["/slow", "/disconnect"])

    def test_checks_enforce_an_overall_deadline(self):
        previous = apt_sources.CHECK_TIMEOUT
        apt_sources.CHECK_TIMEOUT = 0.5
        try:
            started = time.monotonic()
            failures = apt_sources.failed_urls({self.mirror + "/slow"})
            self.assertLess(time.monotonic() - started, 5)
            self.assertIn("exceeded", failures[self.mirror + "/slow"])
        finally:
            apt_sources.CHECK_TIMEOUT = previous

    def test_absent_or_inactive_sources_are_left_alone(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            cases = [
                {},
                {"sources.list": f"# deb {self.mirror}/error noble main\n"},
                {"sources.list.d/ubuntu.sources": f"Types: deb\nURIs: {self.mirror}/error\nEnabled: no\n"},
            ]
            tasks = []
            for index, originals in enumerate(cases):
                apt_root = directory / str(index)
                self.write_sources(apt_root, originals)
                tasks.append(self.source_task(apt_root))
            result = self.run_play(directory, tasks, {})
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            for index, originals in enumerate(cases):
                self.assert_files_unchanged(directory / str(index), originals)
                self.assert_no_extra_files(directory / str(index), originals)
            self.assertEqual(self.server.requests, [])


if __name__ == "__main__":
    unittest.main(verbosity=2)
