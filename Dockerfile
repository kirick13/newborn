
FROM		simplepackages/ansible
COPY		./src /app
WORKDIR		/app
ENTRYPOINT	[ "/bin/sh", "./entrypoint.sh" ]
