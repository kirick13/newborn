#!/usr/bin/env python3
"""Repair unreachable Ubuntu mirrors using only the Python 3 standard library."""

import argparse
from http.client import HTTPException
import json
import multiprocessing
import os
from pathlib import Path
import re
import tempfile
import time
from urllib.error import URLError
from urllib.request import urlopen


SOCKET_TIMEOUT = 15
CHECK_TIMEOUT = 30
SOURCE_FILES = ("sources.list", "sources.list.d/ubuntu.sources")


def repository_urls(content, deb822):
    """Read active repository URLs, ignoring comments and disabled stanzas."""
    content = re.sub(r"(?m)^[ \t]*#.*\n?", "", content)
    if deb822:
        urls = []
        for stanza in re.split(r"\n[ \t]*\n", content):
            fields = re.sub(r"\n[ \t]+", " ", stanza)
            if re.search(r"(?im)^Enabled:[ \t]*no[ \t]*$", fields):
                continue
            for value in re.findall(r"(?im)^URIs:[ \t]*(.*)$", fields):
                urls.extend(value.split())
    else:
        urls = re.findall(
            r"(?m)^[ \t]*deb(?:-src)?[ \t]+(?:\[[^\]\n]*\][ \t]+)?(\S+)",
            content,
        )
    return {url for url in urls if url.startswith(("http://", "https://"))}


def check_url(url):
    # GET supports mirrors which reject HEAD. Only the response headers are
    # needed; urlopen follows redirects and validates HTTPS certificates.
    try:
        with urlopen(url, timeout=SOCKET_TIMEOUT) as response:
            status = response.getcode()
            if status is None or not 200 <= status < 300:
                return "HTTP status {}".format(status)
    except (URLError, OSError, HTTPException, ValueError) as error:
        return str(error)
    return None


def failed_urls(urls):
    if not urls:
        return {}
    # Separate workers allow the deadline to stop even a stalled DNS lookup,
    # which is not necessarily bounded by a socket timeout.
    # Use the same startup method across Python versions, including when tested
    # from a process which already has threads.
    with multiprocessing.get_context("spawn").Pool(processes=len(urls)) as pool:
        jobs = {url: pool.apply_async(check_url, (url,)) for url in sorted(urls)}
        deadline = time.monotonic() + CHECK_TIMEOUT
        failures = {}
        for url, job in jobs.items():
            try:
                error = job.get(timeout=max(0, deadline - time.monotonic()))
            except multiprocessing.TimeoutError:
                error = "Check exceeded {} seconds".format(CHECK_TIMEOUT)
            if error is not None:
                failures[url] = error
        return failures


def official_sources(codename, architecture, deb822):
    if architecture == "amd64":
        archive = "http://archive.ubuntu.com/ubuntu"
        security = "http://security.ubuntu.com/ubuntu"
    else:
        archive = security = "http://ports.ubuntu.com/ubuntu-ports"
    components = "main restricted universe multiverse"
    groups = (
        (archive, [codename, codename + "-updates", codename + "-backports"]),
        (security, [codename + "-security"]),
    )
    entries = []
    for url, suites in groups:
        if deb822:
            entries.append(
                "Types: deb\nURIs: {}\nSuites: {}\nComponents: {}\n"
                "Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg".format(
                    url, " ".join(suites), components
                )
            )
        else:
            entries.extend(
                "deb {} {} {}".format(url, suite, components) for suite in suites
            )
    separator = "\n\n" if deb822 else "\n"
    return "# Managed by Newborn during first setup.\n" + separator.join(entries) + "\n"


def replace_file(path, content):
    # Replace atomically without leaving a backup or a partially written file.
    descriptor, temporary = tempfile.mkstemp(prefix=".newborn-", dir=str(path.parent))
    try:
        with os.fdopen(descriptor, "w", encoding="utf-8") as output:
            output.write(content)
        os.chmod(temporary, 0o644)
        os.replace(temporary, str(path))
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def repair_sources(apt_root, codename, architecture):
    files = {}
    urls = set()
    for name in SOURCE_FILES:
        path = apt_root / name
        if path.is_file():
            content = path.read_text(encoding="utf-8")
            file_urls = repository_urls(content, path.suffix == ".sources")
            files[path] = (content, file_urls)
            urls.update(file_urls)
    failures = failed_urls(urls)
    replaced = []
    for path, (content, file_urls) in files.items():
        if file_urls.intersection(failures):
            replacement = official_sources(codename, architecture, path.suffix == ".sources")
            if content != replacement:
                replace_file(path, replacement)
                replaced.append(str(path))
    return {"changed": bool(replaced), "checked": len(urls), "failed_urls": failures, "replaced": replaced}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--apt-root", type=Path, default=Path("/etc/apt"))
    parser.add_argument("--codename", required=True)
    # Debian package architecture, translated by the playbook's newborn_apt_arch map.
    parser.add_argument("--architecture", required=True,
                        choices=("amd64", "arm64"))
    args = parser.parse_args()
    if not re.fullmatch(r"[a-z][a-z0-9]+", args.codename):
        parser.error("invalid Ubuntu codename")
    print(json.dumps(repair_sources(args.apt_root, args.codename, args.architecture)))


if __name__ == "__main__":
    main()
