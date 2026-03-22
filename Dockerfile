FROM       simplepackages/ansible-core:2.17.0
COPY       ./src /app
WORKDIR    /app
ENV        ANSIBLE_SSH_PIPELINING=True
RUN        ansible-galaxy collection install community.general
ENTRYPOINT ["/bin/sh", "./entrypoint.sh"]
