
FROM       simplepackages/ansible
COPY       ./src /app
WORKDIR    /app
ENV        ANSIBLE_SSH_PIPELINING=True
ENTRYPOINT [ "/bin/sh", "./entrypoint.sh" ]
