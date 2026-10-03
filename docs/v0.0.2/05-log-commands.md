# v0.0.2 doc 05: Log commands

## Goal

Make persisted logs actually usable: view them, follow the tail, and pass
Gradle verbosity flags — from both CLI and TUI.

## Current state

`logs` (`cmd/logs.go`) only lists ≤20 filenames in `logs/`; there is no
`show`, no `tail`, no way to read content except opening the file manually.
Build/test always run at default verbosity; no `--stacktrace/--info/--debug`
passthrough exists. No log index beyond filename sort by mtime.

## Design

Extend `cmd/logs.go` (thin; logic in `internal/log`):

    logs                    list (existing behavior, keep format)
    logs show <name>        print file (name may be a prefix — unambiguous match or error listing candidates)
    logs last               print newest log
    logs tail [-n 50] <name>|last   last N lines (default 50)

`show`/`tail` print a header line `# <abs path>` then content, so output stays
greppable. Binary-safety: cap output at ~1 MiB with a `…​truncated` note.

Gradle passthrough: `build`/`test` gain `--stacktrace --info --debug` bool
flags, appended verbatim to the Gradle argv (`--stacktrace`, `--info`,
`--debug`); rejected (clear error) for the `go` type since `go build/test`
do not accept them. The exact argv is already persisted as the `$ …​` header
line by `build.Run`, so reproducibility is free.

Log index: `logs/` listing gains an `index.json` written on each run
(`{name, command, started_at, exit_code, project}` appended, capped at 200
entries, best-effort — never fail a build for index errors). `logs last`
reads the index, falling back to mtime order.

## Files to touch

- `cmd/logs.go` (subcommands `show|last|tail`, `-n` flag)
- `internal/log/` (new `index.go`: `Append/Last/List`; reader with cap)
- `cmd/build.go`, `cmd/test.go` (three passthrough flags; go-type rejection)
- TUI `doLogs` (accept `show|tail` args, same output into viewport)

## Steps

1. Add index read/write + tests.
2. Add `show|last|tail` to `logs` (+ prefix matching + truncation).
3. Add the three flags to `build`/`test` with go-type rejection.
4. Wire TUI `logs [show|tail] …​`.
5. Document in `RELEASE.md` troubleshooting (which log to open first).

## Tests to add

- Index append/cap/`Last` fallback with missing index.
- Prefix-match: unique, ambiguous, no-match cases.
- Tail `-n` boundaries; 1 MiB truncation note.
- Flags land verbatim in the `$` header line of a fixture run.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- On this repo: `go run . test >/dev/null; go run . logs last | head -5`
  shows the just-written log; `go run . logs tail -n 3 last` prints 3 lines.
- `go run . build --stacktrace` fails fast with the go-type rejection.

## Out of scope

Live `tail -f` (covered by TUI streaming in 13), log rotation/deletion
commands, remote log shipping.
