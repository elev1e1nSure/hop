# Agent Instructions

See [CLAUDE.md](./CLAUDE.md) for full project context: architecture, conventions, key invariants, and how to run.

## Quick reference

- Language: Go 1.23, module `hop`
- Entry point: `cmd/hop/main.go`
- TUI framework: Bubble Tea (`bubbletea` + `bubbles` + `lipgloss`)
- Run: `just run`
- Test: `just test`

## Critical rules

- Only `internal/sshconfig` writes to `~/.ssh/config`.
- `ui.Model` is a value type — mutations return a new copy.
- All domain errors go through `apperr.Wrap`.
- i18n keys and both translations (en/ru) are added together in `internal/i18n/i18n.go`.
- Styles live in `internal/ui/styles.go` only.
