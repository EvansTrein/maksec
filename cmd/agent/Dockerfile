FROM alpine:3.20

RUN apk add --no-cache openssh inotify-tools bash

ENV SSH_PASSWORD=rootpassword

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh \
    && mkdir -p /run/sshd

EXPOSE 22

ENTRYPOINT ["/entrypoint.sh"]
