FROM       simplepackages/ansible-core:2.17.0
RUN        ansible-galaxy collection install community.general
WORKDIR    /app
COPY       . .
ENV        ANSIBLE_SSH_PIPELINING=True
ENTRYPOINT ["ansible-playbook", "playbook.yaml", "-i", "inventory.yaml"]
