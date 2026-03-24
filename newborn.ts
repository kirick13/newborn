#!/usr/bin/env bun run

// oxlint-disable no-console

import node_path from 'node:path';
import { DockerBinds } from './src/docker/binds.js';
import { inquireHosts } from './src/inquire/hosts.js';
import { inquireSetup } from './src/inquire/setup.js';
import { inquireSoftware } from './src/inquire/software.js';
import { exec } from './src/utils.js';

console.log('Welcome to Newborn!');

const hosts = await inquireHosts();
const setup = await inquireSetup();
const software = await inquireSoftware();

const dockerBinds = new DockerBinds();

const inventory_path = node_path.join('/tmp', `inventory.${Date.now()}.yaml`);
dockerBinds.add(inventory_path, '/app/inventory.yaml');
const inventory = {
	all: {
		hosts: Object.fromEntries(
			hosts.map((host) => {
				const params: Record<string, unknown> = {};

				if (host.connect.ssh_port !== 22) {
					params.ansible_ssh_port = host.connect.ssh_port;
				}

				if (host.connect.password) {
					params.ansible_ssh_pass = host.connect.password;
				}

				if (host.connect.ssh_key_path) {
					params.ansible_ssh_private_key_file = dockerBinds.add(
						host.connect.ssh_key_path,
						`/opt/bind/ssh/${host.setup.name}.key`,
					);
				}

				params.newborn_hostname = host.setup.hostname;
				params.newborn_name = host.setup.name;
				params.newborn_user = host.setup.username;
				params.newborn_password = host.setup.password;
				params.newborn_password_salt = host.setup.password_salt;
				params.newborn_ssh_port = host.setup.ssh_port;
				params.newborn_ssh_key_public = host.setup.ssh_key_public;

				return [host.connect.ip, params];
			}),
		),
	},
};

console.log();

await exec(
	'docker',
	'build',
	'-t',
	'local/newborn',
	node_path.join(import.meta.dir, 'container'),
);

await Bun.write(inventory_path, Bun.YAML.stringify(inventory, null, 2), {
	mode: 0o600,
});

await exec(
	'docker',
	'run',
	'-t',
	'--rm',
	...dockerBinds.args,
	'local/newborn',
	'-e',
	JSON.stringify({
		// setup
		newborn_swap: setup.swap,
		newborn_firewall_http: setup.firewall_http,
		// software
		newborn_oci_runtime: software.oci_runtime,
		newborn_oci_compose: software.oci_compose ? 'y' : '',
		newborn_k8s_runtime: software.k8s_runtime,
		newborn_remove_snap: software.remove_snap ? 'y' : '',
	} satisfies Record<string, string>),
);

console.log('Setup complete!');

for (const host of hosts) {
	console.log();
	console.log(`Host ${host.setup.name}:`);
	console.log(`  IP:               ${host.connect.ip}`);
	console.log(`  SSH Port:         ${host.setup.ssh_port}`);
	console.log(`  Hostname:         ${host.setup.hostname}`);
	console.log(`  User:             ${host.setup.username}`);
	console.log(`  Password:         ${host.setup.password}`);
	console.log(`  SSH private key:  ${host.setup.ssh_key_path}`);
}
