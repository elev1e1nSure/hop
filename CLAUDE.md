# sshm

TUI SSH manager written in Go. Reads `~/.ssh/config`, lets you browse, filter, add, edit, delete servers, and connect to them. Built with Bubble Tea (charmbracelet stack).

## Stack

- Go 1.23, module name `sshm`
- TUI: `bubbletea` + `bubbles` (list, textinput, spinner) + `lipgloss`
- No external state, no database — `~/.ssh/config` is the source of truth; history is a separate JSON file

## Architecture

```
cmd/sshm/        — entry point; parses CLI args, wires translator, calls app.Run
internal/app/    — main loop: run UI → connect → repeat
internal/cli/    — flag parsing (--language)
internal/config/ — XDG/OS-aware paths for ssh config and history file
internal/domain/ — Server, HistoryRecord structs (shared data models)
internal/sshconfig/ — parse, resolve, mutate ~/.ssh/config
internal/sshclient/ — locate ssh binary, exec connect
internal/history/   — load/save connection history JSON
internal/i18n/      — en/ru translations via Translator interface
internal/ui/        — Bubble Tea model: browse / form / confirm-delete / connecting modes
internal/apperr/    — typed errors with wrapping
internal/util/      — fs helpers, fuzzy match, Max
```

## Key invariants

- `sshconfig` owns writes to `~/.ssh/config`. Never write that file from anywhere else.
- `domain.Server` is read-only after construction — mutations go through `sshconfig` functions that reparse and save.
- `ui.Model` is a value type (Bubble Tea convention) — all state mutations return a new copy, not pointers, except for slice/map fields modified via index.
- `i18n.Translator` is always passed down; never call `translator.T(...)` at package init or in `init()`.

## Running

```bash
go run ./cmd/sshm
go run ./cmd/sshm --language ru
go run ./cmd/sshm --language en
```

Language auto-detected from `LC_ALL` / `LC_MESSAGES` / `LANG`; falls back to English.

## Tests

```bash
go test ./...
```

Tests exist for: `cli`, `sshclient`, `history`, `sshconfig`, `i18n`. No TUI tests.

## Conventions

- Error wrapping: always use `apperr.Wrap(apperr.ErrXxx, fmt.Errorf(...))` — never raw `errors.New` for domain errors.
- i18n keys live in `internal/i18n/i18n.go` as constants (`MsgXxx`). Add key + both translations together.
- Styles are centralized in `internal/ui/styles.go` — don't create ad-hoc `lipgloss.NewStyle()` elsewhere.
