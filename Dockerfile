
FROM		kirickme/ansible
COPY		./src /app
WORKDIR		/app
ENTRYPOINT	[ "/bin/bash", "./entrypoint.sh" ]
