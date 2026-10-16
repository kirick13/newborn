# Testing Newborn

The project currently has no automated test cases or VM test harness. This document
describes the recommended next step; it does not claim that Ubuntu compatibility
has been verified.

## Fast checks

Run the commands in the README from the repository root. They check compilation,
Go diagnostics, and shell syntax without contacting a server or a vault.

Useful future unit tests cover behavior at the external boundaries: escaped values
in inventory, shell arguments containing spaces and quotes, conflicting software
options, and `bw` payloads and errors. A fake executable can stand in for `bw` in
these tests while the real application continues to require Bitwarden.

## Fresh machines on a developer laptop

Use [Multipass](https://canonical.com/multipass/docs/latest/how-to-guides/install-multipass/)
to create disposable Ubuntu VMs. It supports macOS, including Apple Silicon, and
accepts [cloud-init configuration at launch](https://canonical.com/multipass/docs/latest/reference/command-line-interface/launch/).
Newborn changes swap, systemd services, SSH, and kernel firewall state, so a full VM
is the appropriate target for provisioning tests.

An automated runner should perform this lifecycle for each scenario:

1. Create a unique VM name, temporary working directory, and bootstrap SSH key.
2. Generate cloud-init that preserves the VM management user, enables initial root
   key authentication, and installs the bootstrap public key. Do not preconfigure
   the features that Newborn is meant to test.
3. Launch a clean Ubuntu image and wait for cloud-init and package operations to
   finish. Discover its address and check access from the Ansible Docker runner.
4. Generate test credentials and inventory, then invoke the same container/playbook
   used by the application with explicit scenario options.
5. Reconnect as the generated user on the new SSH port using its generated key.
   Check user groups, sudo access, hostname, timezone, locale, swap, installed
   software, and effective SSH/firewall configuration.
6. Reboot the VM and repeat the connection and persistence checks. Capture failure
   logs before cleanup, excluding credentials.
7. Destroy only the VM created for that scenario, even if a test fails, and remove
   its temporary keys and inventory.

For example, after the runner has generated `cloud-init.yaml`, VM creation can use:

```sh
multipass launch 24.04 --name newborn-test-2404 --cpus 2 --memory 4G --disk 20G \
  --cloud-init cloud-init.yaml
```

Cleanup is scoped to that instance:

```sh
multipass delete --purge newborn-test-2404
```

Multipass documents [per-instance permanent deletion](https://canonical.com/multipass/docs/latest/reference/command-line-interface/delete/).
The example is a lifecycle illustration, not a complete test command: the cloud-init
generator, inventory builder, verifier, and cleanup runner still need implementation.
Provisioning changes SSH access, so verify how the installed Multipass version can
manage/reboot/delete the VM afterward; retain a hypervisor-side recovery path.

## Coverage to build incrementally

Start with a minimal setup scenario on Ubuntu 22.04, 24.04, and 26.04, then add
separate fresh-VM scenarios for Docker + Compose, Podman + Compose, Snap removal,
swap/reserve options, and the Kubernetes installers. The product target remains
Ubuntu 22–26; coverage of interim releases has not yet been decided.

Run firewall checks for each HTTP policy. Local private-network traffic is allowed
by the current script, so successful access from the developer machine does not
prove that public traffic is filtered. Public-policy verification needs a separate
source/network outside the allowed private ranges, or a controlled network namespace
setup inside the test environment. Check IPv4 and IPv6 explicitly.

The first VM run should check compatibility with the Python version supplied by
each Ubuntu image: the Ansible image is currently pinned to ansible-core 2.17.0.
Treat that compatibility as unverified until provisioning has actually passed.

Apple Silicon VM runs cover ARM64. Add a Linux x86_64 VM runner or disposable cloud
instances if AMD64 coverage is required. Confirm Docker-to-VM routing on the local
host before relying on the laptop workflow; a Linux VM host is an alternative if
the local Docker and Multipass networks cannot communicate.

## Bitwarden and the complete user flow

Most provisioning checks can invoke Ansible directly and retain generated test
keys in temporary files. That isolates server behavior; it does not test the full
application or introduce a local-file mode into the product.

A separate end-to-end check should run the real UI and real `bw` CLI with a dedicated
test vault/account. Verify the saved SSH key and fields by reading them back and
using the saved key to connect to the test server. Track the IDs of created items
for precise cleanup. The existing production Bitwarden flow remains unchanged.

Because the UI is interactive and saves to Bitwarden after provisioning, fully
automating this end-to-end check will need terminal interaction and managed test
vault credentials. It is reasonable to keep this check manual while automating
the VM/playbook checks first.
