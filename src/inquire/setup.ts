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

			if (/^\d{1,3}(?:M|G)$/u.test(value)) {
				return true;
			}

			return 'Please enter a valid size like "100M" or "1G".';
		},
	});

	const reserve_file = await inquirer.confirm({
		message:
			'Do you want to create a 2G disk reserve file? (Can be deleted to free space in emergencies)',
		default: false,
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
		reserve_file,
		firewall_http,
	};
}
