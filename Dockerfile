
FROM kirickme/ansible

COPY . /app
WORKDIR /app

ENTRYPOINT [ "/bin/bash", "./newborn.sh" ]
