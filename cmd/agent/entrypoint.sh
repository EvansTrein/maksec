#!/bin/sh
set -e

# chpasswd читает пару user:password со stdin; пароль в командной строке не светится
echo "root:${SSH_PASSWORD}" | chpasswd

# root может логиниться по паролю
sed -i 's/#\?PermitRootLogin.*/PermitRootLogin yes/' /etc/ssh/sshd_config
sed -i 's/#\?PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config

# Host-ключи: генерируются при первом старте контейнера, чтобы sshd мог подняться
ssh-keygen -A

# sshd в foreground — PID 1, корректно получает сигнал завершения контейнера
exec /usr/sbin/sshd -D
