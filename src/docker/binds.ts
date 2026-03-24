export class DockerBinds {
	args: string[] = [];

	add(path_source: string, path_bind: string) {
		this.args.push('-v', `${path_source}:${path_bind}:ro`);

		return path_bind;
	}
}
