## Быстрый старт
```bash
make stend-up
```

Все поднимется в docker, миграции при старте сервиса

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `SSH_PASSWORD` | `rootpassword` | Пароль root для SSH на целевом хосте (`maksec_target`). Используется в запросах `create` как `"password"`. Переопределение: `SSH_PASSWORD=mypass make stend-up` |

Все прочие настройки (порты, БД, пути) — в `config.yaml` и `config.docker.yaml`.

## Создание скрипта

Успешный запрос (201):
```bash
curl --request POST \
  --url http://localhost:8080/scripts/v1/create \
  --header 'Content-Type: application/json' \
  --data '{
  "host": "target-host",
  "user": "root",
  "password": "rootpassword",
  "template": "template1"
}'
```

Неверный SSH-пароль — 401, `scripts.ssh.authFailed`:
```bash
curl --request POST \
  --url http://localhost:8080/scripts/v1/create \
  --header 'Content-Type: application/json' \
  --data '{
  "host": "target-host",
  "user": "root",
  "password": "invalidpassword",
  "template": "template1"
}'
```

Неизвестный шаблон — 404, `scripts.template.notFound`:
```bash
curl --request POST \
  --url http://localhost:8080/scripts/v1/create \
  --header 'Content-Type: application/json' \
  --data '{
  "host": "target-host",
  "user": "root",
  "password": "rootpassword",
  "template": "templateNotfound"
}'
```
