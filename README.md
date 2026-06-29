# hop

TUI-менеджер SSH-хостов для терминала. Работает поверх `~/.ssh/config` — не требует отдельной базы, не ломает существующий конфиг.

## Что умеет

- Быстрый поиск хостов по алиасу, IP, пользователю
- Добавление, редактирование, удаление записей прямо из TUI
- Подключение одной клавишей — все директивы (`ProxyJump`, `IdentityFile` и др.) читаются из конфига
- Индикатор доступности хостов в реальном времени
- История подключений с частотой использования
- Русский и английский интерфейс (автоопределение по локали)

## Установка

```bash
go install github.com/elev1e1nSure/hop/cmd/hop@latest
```

или собрать локально:

```bash
just build
```

## Использование

```bash
hop
hop --help
hop --language en
hop --language ru
hop --path add
hop --path remove
```

If you change PATH in PowerShell and want the current session to pick it up without reopening the shell:

```powershell
$env:Path = [Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [Environment]::GetEnvironmentVariable('Path','User')
```

| Клавиша | Действие |
|---------|----------|
| `↑` / `↓` | навигация |
| `Enter` | подключиться |
| `Ctrl+N` | добавить хост |
| `Ctrl+E` | редактировать |
| `Ctrl+D` | удалить |
| `Tab` | показать/скрыть помощь |
| `Esc` | сбросить фильтр |
| `Ctrl+C` | выйти |
