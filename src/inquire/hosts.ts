// oxlint-disable no-await-in-loop, max-lines-per-function, no-console

import crypto from 'node:crypto';
import fs from 'node:fs/promises';
import { isIP } from 'node:net';
import os from 'node:os';
import path from 'node:path';
import * as inquirer from '@inquirer/prompts';
import { customAlphabet } from 'nanoid';
import * as sshpk from 'sshpk';

const nanoidAlphanumeric = customAlphabet(
	'1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz',
);
const nanoidAlphanumericLowercase = customAlphabet(
	'1234567890abcdefghijklmnopqrstuvwxyz',
);

interface Host {
	connect: {
		ip: string;
		password?: string;
		ssh_port: number;
		ssh_key_path?: string;
	};
	setup: {
		hostname: string;
		name: string;
		username: string;
		password: string;
		password_salt: string;
		ssh_port: number;
		ssh_key_path: string;
		ssh_key_public: string;
	};
}

// eslint-disable-next-line jsdoc/require-jsdoc
function resolvePath(value: string): string {
	if (value === '') {
		return value;
	}

	if (value.startsWith('~/')) {
		value = path.join(os.homedir(), value.slice(2));
	}

	return path.resolve(value);
}

/**
 * Prompts the user to define a list of hosts to configure.
 * @returns A promise that resolves to an array of Host objects.
 */
export async function inquireHosts() {
	const hosts: Host[] = [];

	console.log();
	console.log('Define a servers to configure:');

	let last_ssh_key_path: string | undefined = process.env.SSH_KEY_PATH;

	while (true) {
		const host: Host = {
			connect: {
				ip: await inquirer.input({
					message: 'Enter the IP address of the server:',
					default: process.env.IP,
					validate(value) {
						if (!isIP(value)) {
							return 'Please enter a valid IP address.';
						}

						return true;
					},
				}),
				password: undefined,
				ssh_port: await inquirer.number({
					message: 'Enter the port:',
					default: 22,
					required: true,
					validate(value) {
						if (Number.isNaN(value) || value < 1 || value > 65535) {
							return 'Please enter a valid port number.';
						}

						return true;
					},
				}),
				ssh_key_path: undefined,
			},
			setup: {
				hostname: `host-${nanoidAlphanumericLowercase(8)}`,
				name: '',
				username: nanoidAlphanumericLowercase(7),
				password: nanoidAlphanumeric(100),
				password_salt: nanoidAlphanumeric(16),
				ssh_port: crypto.randomInt(1025, 65536),
				ssh_key_path: '',
				ssh_key_public: '',
			},
		};

		const auth_type = await inquirer.select({
			message: 'Select authentication type:',
			choices: [
				{ name: 'Password', value: 'password' as const },
				{ name: 'SSH Key', value: 'ssh-key' as const },
			],
		});

		if (auth_type === 'ssh-key') {
			const ssh_key_path = await inquirer.input({
				message: `Enter the SSH key path to connect:`,
				default: last_ssh_key_path,
				async validate(value) {
					if (value.length === 0) {
						return 'Please enter a valid SSH key path.';
					}

					if ((await Bun.file(resolvePath(value)).exists()) !== true) {
						return 'The SSH key path does not exist.';
					}

					return true;
				},
			});

			last_ssh_key_path = ssh_key_path;

			host.connect.ssh_key_path = ssh_key_path
				? resolvePath(ssh_key_path)
				: undefined;
		} else {
			host.connect.password = await inquirer.password({
				message: 'Enter the password:',
				validate: (value) => value.length > 0,
			});
		}

		host.setup.name = await inquirer.input({
			message: 'Enter the name for the host:',
			required: true,
			default: process.env.NAME,
		});

		let ssh_key_path = await inquirer.input({
			message: 'Enter the SSH key path for new user:',
			required: true,
			default: process.env.SSH_KEY_NEW_PATH,
		});
		host.setup.ssh_key_path = ssh_key_path;
		ssh_key_path = resolvePath(ssh_key_path);

		let generate_ssh_key;
		if (await Bun.file(ssh_key_path).exists()) {
			generate_ssh_key = await inquirer.select({
				message: `File ${ssh_key_path} already exists.`,
				choices: [
					{ name: 'Generate a new key and overwrite file', value: true },
					{ name: 'Use existing key', value: false },
				],
			});
		} else {
			generate_ssh_key = true;
		}

		let ssh_key_private;
		if (generate_ssh_key) {
			ssh_key_private = sshpk.generatePrivateKey('ed25519');

			try {
				await Bun.file(ssh_key_path).unlink();
			} catch {}

			// await Bun.write(ssh_key_path, ssh_key_private.toString('openssh'), {
			// 	mode: 0o600,
			// });
			await fs.writeFile(ssh_key_path, ssh_key_private.toString('openssh'), {
				mode: 0o600,
			});
		} else {
			const ssh_key_contents = await Bun.file(resolvePath(ssh_key_path)).text();
			ssh_key_private = sshpk.parsePrivateKey(ssh_key_contents);
		}

		host.setup.ssh_key_public = ssh_key_private
			.toPublic()
			.toString('ssh')
			.replace(' (unnamed)', '');

		hosts.push(host);

		const another = await inquirer.confirm({
			message: 'Add another host?',
			default: false,
		});

		if (!another) {
			break;
		}
	}

	return hosts;
}
