
FROM ubuntu:jammy

RUN apt update \
    && apt install -y \
        software-properties-common \
        python3 \
    && apt-add-repository ppa:ansible/ansible \
    && apt update \
    && apt install -y ansible \
    && rm -rf /var/lib/apt/lists/*

COPY . /app
WORKDIR /app

ENTRYPOINT [ "/bin/bash", "./newborn.sh" ]
