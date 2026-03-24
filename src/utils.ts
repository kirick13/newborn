/**
 * Executes a command with the given arguments and returns the exit code.
 * @param command The command to execute.
 * @param args The arguments to pass to the command.
 * @returns A promise that resolves to the exit code of the command.
 */
export async function exec(command: string, ...args: string[]) {
	const proc = Bun.spawn({
		cmd: [command, ...args],
		stdout: 'inherit',
		stderr: 'inherit',
		terminal: {
			cols: process.stdout.columns ?? 80,
			rows: process.stdout.rows ?? 24,
			data(_term, data) {
				process.stdout.write(data);
			},
		},
	});

	process.on('SIGWINCH', () => {
		if (proc.terminal) {
			proc.terminal.resize(
				process.stdout.columns ?? 80,
				process.stdout.rows ?? 24,
			);
		}
	});

	await proc.exited;
	return proc.exitCode!;
}
