# Claude Code Instructions

## Language

Communicate in Spanish unless the user switches to English.

## Issue Tracking (beads)

Issues are managed with **bd** (beads). The database is `.beads/issues.jsonl`.

### Workflow

1. At session start, run `bd ready` to see pending work
2. Use `bd show <id>` to read full context before starting
3. Use `bd update <id> --status in_progress` when starting a task
4. Use `bd close <id>` when done — never edit the JSONL manually
5. Use `bd create -t "Title" -d "Description"` to create new issues

### Rules

- **Never edit `.beads/issues.jsonl` directly** — always use `bd` CLI
- Close issues only after the code compiles and tests pass
- If a task spawns follow-up work, create new issues with `bd create`

## Code Quality

- Run `go build ./...` and `go vet ./...` before committing
- Run `make lint` to check all linters (staticcheck, gofumpt, golines, golangci-lint, govulncheck, markdownlint)
- Max line length: 120 characters (enforced by golines)
- Formatter: gofumpt (not gofmt)

## Git

- Commit messages in English
- Push after committing — work is not done until pushed
- Run `bd sync` before pushing
