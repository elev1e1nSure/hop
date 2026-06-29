# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# hop

TUI SSH manager written in Go. Reads `~/.ssh/config`, lets you browse, filter, add, edit, delete servers, and connect to them. Built with Bubble Tea (charmbracelet stack).

## Stack

- Go 1.23, module name `hop`
- TUI: `bubbletea` + `bubbles` (list, textinput, spinner) + `lipgloss`
- No external state, no database — `~/.ssh/config` is the source of truth; history is a separate JSON file

## Architecture

```
cmd/hop/        — entry point; parses CLI args, wires translator, calls app.Run
internal/app/    — main loop: run UI → connect → save history → repeat
internal/cli/    — flag parsing (--language)
internal/config/ — XDG/OS-aware paths for ssh config and history file
internal/domain/ — Server, HistoryRecord structs (shared data models)
internal/sshconfig/ — parse, resolve, mutate ~/.ssh/config (owning package)
internal/sshclient/ — locate ssh binary, build args, exec connect
internal/history/   — load/save connection history JSON
internal/i18n/      — en/ru translations via Translator interface
internal/ui/        — Bubble Tea model: browse / form / confirm-delete / connecting modes
internal/apperr/    — typed errors with wrapping
internal/pathenv/   — add/remove binary directory from user PATH; platform-specific (Windows vs others)
internal/util/      — fs helpers (AtomicWrite with symlink resolution), fuzzy match, truncate, sanitize
```

## Key invariants

- `sshconfig` owns writes to `~/.ssh/config`. Never write that file from anywhere else.
- Before every mutation (Add/Edit/Delete), `config.syncFromDisk()` re-reads the file to avoid overwriting parallel external changes.
- On write failure, in-memory `config.Lines` and `config.HadTrailing` are rolled back to the pre-mutation snapshot.
- `AtomicWrite` resolves symlinks before rename — if `~/.ssh/config` is a symlink, the write targets the real path, preserving the link.
- `domain.Server` is read-only after construction — mutations go through `sshconfig` functions that reparse and save.
- `ui.Model` is a value type (Bubble Tea convention) — all state mutations return a new copy, not pointers, except for slice/map fields modified via index.
- `i18n.Translator` is always passed down; never call `translator.T(...)` at package init or in `init()`.
- All user-controlled strings from SSH config must be sanitized via `util.Sanitize` before rendering in TUI (strips control chars, ESC sequences).

## Connection model

- SSH is invoked as `ssh <alias>` (or `ssh user@<alias>`). No `-p`, `-i`, or `HostName` flags are passed — all directives (ProxyJump, StrictHostKeyChecking, IdentityFile, etc.) are read by `ssh` from its config.
- TCP liveness checks are rate-limited (20 concurrent) and skip hosts with ProxyJump/ProxyCommand directives.

## Parsing model

- The parser recognizes `HostName`, `User`, `Port`, `IdentityFile`, `ProxyJump`, `ProxyCommand`.
- Only the first `HostName`/`User`/`Port` value is resolved into `domain.Server`; duplicate `IdentityFile` lines are preserved through edits (not dropped).
- Multi-host block edits preserve all directives from the original block — only the edited canonical four are updated.
- Global directives and `Include` are preserved as text but not semantically processed.

## Running

```bash
just run          # auto-detect language
just run ru       # Russian
just run en       # English
just build        # compile binary
just lint         # go vet
just fmt          # gofmt + golines
```

CLI flags: `--language en|ru`, `--path add|remove` (adds/removes the binary directory from PATH), `-h/--help`.

Language auto-detected from `LC_ALL` / `LC_MESSAGES` / `LANG`; falls back to English.

## Tests

```bash
just test                                    # all packages
go test ./internal/sshconfig/...            # single package
go test -run TestFoo ./internal/sshconfig/  # single test
```

Tests exist for: `cli`, `sshclient`, `history`, `sshconfig`, `i18n`. No TUI tests.

## Conventions

- Error wrapping: always use `apperr.Wrap(apperr.ErrXxx, fmt.Errorf(...))` — never raw `errors.New` for domain errors.
- i18n keys live in `internal/i18n/i18n.go` as constants (`MsgXxx`). Add key + both translations together. Remove both together when deleting.
- TUI styles are centralized in `internal/ui/styles.go` — don't create ad-hoc `lipgloss.NewStyle()` in TUI code. Exception: `internal/cli/help.go` has its own local styles for CLI terminal output (not TUI), which is intentional.
- Dead code (unused constants, functions, i18n keys) must be cleaned up immediately.
