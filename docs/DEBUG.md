# Отладка и профилирование

Профилировщик доступен только в debug-сборке. В обычной сборке код профилировщика
не попадает в бинарник — оверхед нулевой.

## Сборка

```bash
go build -tags debug ./cmd/maksec   # собрать с профилировщиком
# обычная сборка — файла просто нет в бинарнике
```

## Конфигурация (ENV)

| Переменная | По умолчанию | Описание |
|---|---|---|
| `DEBUG_PPROF_ADDR` | `127.0.0.1:6060` | Адрес HTTP-сервера pprof. По умолчанию только localhost — не выставлять наружу без необходимости |
| `DEBUG_BLOCK_RATE` | `100` | Частота сэмплирования block-профиля (~1/N событий ожидания). `0` — выключить |
| `DEBUG_MUTEX_FRACTION` | `100` | Частота сэмплирования mutex-профиля (~1/N событий захвата). `0` — выключить |

## Снятие профилей

```bash
curl http://localhost:6060/debug/pprof/heap            # память
curl -o cpu.prof "http://localhost:6060/debug/pprof/profile?seconds=10"  # CPU
curl -o trace.out "http://localhost:6060/debug/pprof/trace?seconds=5"    # трассировка
curl -o block.prof "http://localhost:6060/debug/pprof/block"            # где горутины ждут
curl -o mutex.prof "http://localhost:6060/debug/pprof/mutex"            # кто держит мьютекс

go tool pprof cpu.prof
go tool trace trace.out
go tool pprof block.prof
go tool pprof mutex.prof
```

## Когда какой профиль смотреть

- **CPU** — сервис грузит процессор, надо найти горячие функции.
- **heap** — растёт потребление памяти, ищем утечки.
- **block** — сервис медленный при простаивающем CPU: горутины ждут
  мьютексы, каналы, `select`.
- **mutex** — дополнение к block: показывает не кто ждёт, а какой код
  держит локу. Чинить нужно держателя.

## Живой анализ против работающего процесса

```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```
