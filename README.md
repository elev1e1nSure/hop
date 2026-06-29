# sshm

Модульная версия SSH-менеджера с TUI.

## Структура

- `cmd/sshm/` — точка входа CLI.
- `internal/app/` — жизненный цикл приложения.
- `internal/cli/` — разбор аргументов командной строки.
- `internal/config/` — системные пути приложения.
- `internal/sshconfig/` — чтение, разбор и изменение `~/.ssh/config`.
- `internal/sshclient/` — поиск и запуск `ssh`.
- `internal/history/` — хранение истории подключений.
- `internal/i18n/` — русская и английская локализация.
- `internal/ui/` — Bubble Tea TUI без изменения поведения и визуальной схемы.
- `internal/domain/` — доменные модели.
- `internal/util/` — общие файловые и текстовые функции.

## Запуск

```bash
go run ./cmd/sshm
go run ./cmd/sshm --language ru
go run ./cmd/sshm --language en
```

Без `--language` язык определяется по `LC_ALL`, `LC_MESSAGES` или `LANG`; при неизвестной локали используется английский.
