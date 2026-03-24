// oxlint-disable no-console

import * as inquirer from '@inquirer/prompts';

/**
 * Inquire the user for the setup method (sudo or non-sudo).
 * @returns -
 */
export async function inquireSetup() {
	console.log();
	console.log('How do you want to set up the server?');

	const swap = await inquirer.input({
		message:
			'How much swap space do you want to allocate? (Leave blank for none)',
		validate(value) {
			if (value === '') {
				return true;
			}

			if (/^\d{1,3}(M|G)$/.test(value)) {
				return true;
			}

			return 'Please enter a valid size like "100M" or "1G".';
		},
	});

	const firewall_http = await inquirer.select({
		message: 'From where will HTTP/HTTPS traffic be allowed?',
		choices: [
			{ name: 'From anywhere', value: 'anywhere' as const },
			{ name: 'From Cloudflare only', value: 'cloudflare' as const },
			{ name: 'From nowhere', value: 'nowhere' as const },
		],
	});

	return {
		swap,
		firewall_http,
	};
}
