// oxlint-disable no-console

import * as inquirer from '@inquirer/prompts';

/**
 * Inquire the user for software configuration.
 * @returns -
 */
export async function inquireSoftware() {
	console.log();
	console.log('What software do you want on servers?');

	const oci_runtime = await inquirer.select({
		message: 'What container runtime do you want to use?',
		choices: [
			{ name: 'Docker', value: 'docker' as const },
			{ name: 'Podman', value: 'podman' as const },
			{ name: 'None', value: '' as const },
		],
	});

	let oci_compose = false;
	if (oci_runtime !== '') {
		oci_compose = await inquirer.confirm({
			message: `Do you want to use ${oci_runtime} compose?`,
			default: false,
		});
	}

	const k8s_runtime = await inquirer.select({
		message: 'What Kubernetes runtime do you want to use?',
		choices: [
			{ name: 'k0s', value: 'k0s' as const },
			{ name: 'Microk8s', value: 'microk8s' as const },
			{ name: 'None', value: '' as const },
		],
	});

	let remove_snap = false;
	if (k8s_runtime !== 'microk8s') {
		remove_snap = await inquirer.confirm({
			message: `Do you want to remove snap?`,
		});
	}

	return {
		oci_runtime,
		oci_compose,
		k8s_runtime,
		remove_snap,
	};
}
