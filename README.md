# toss-investment-cli

Unofficial, read-only CLI for Toss Securities web data.

## Status

This repository is in bootstrap stage. The current codebase provides:

- a Go-first CLI skeleton
- local path and session conventions
- a placeholder auth boundary for a future Python-assisted browser login helper
- command groups for the planned read-only surface

Trading is out of scope. The first release targets developer workflows only.

## Architecture

- `Go`: main CLI, domain model, read-only client, output rendering, session lifecycle
- `Python`: future browser login helper and reverse-engineering utilities
- `Rust`: optional later addition for isolated performance-sensitive workers if real need appears

The design and implementation plan live in:

- [`docs/plans/2026-03-11-toss-investment-cli-design.md`](docs/plans/2026-03-11-toss-investment-cli-design.md)
- [`docs/plans/2026-03-11-toss-investment-cli-implementation-plan.md`](docs/plans/2026-03-11-toss-investment-cli-implementation-plan.md)

## Current Command Surface

```bash
tossctl auth login
tossctl auth status
tossctl auth logout

tossctl account list
tossctl account summary
tossctl portfolio positions
tossctl portfolio allocation
tossctl orders list
tossctl watchlist list
tossctl quote get <symbol>
tossctl export positions --format csv
tossctl export orders --format json
```

`auth login`, `auth import-playwright-state`, `auth status`, `auth logout`, `quote get <symbol>`, `account list`, `account summary`, `orders list`, `portfolio positions`, `portfolio allocation`, and `watchlist list` work today.

`auth status` performs a live validation check when a stored session exists. Authenticated commands return a re-login prompt when the stored session is missing or rejected.

## Local Paths

By default, the CLI uses OS-native paths:

- config dir: `$(os.UserConfigDir)/tossctl`
- cache dir: `$(os.UserCacheDir)/tossctl`
- session file: `<config dir>/session.json`

During development you can override paths with:

- `--config-dir`
- `--session-file`

## Development

```bash
make tidy
make fmt
make build
make test
cd auth-helper && python3 -m pip install -e . && python3 -m playwright install chromium
./bin/tossctl --help
./bin/tossctl auth login
./bin/tossctl auth status
./bin/tossctl quote get A005930
./bin/tossctl account list
./bin/tossctl account summary
./bin/tossctl portfolio positions
./bin/tossctl watchlist list --output json
```

## Safety Boundary

This project is intended to stay read-only. The future client layer will enforce an allowlist of read-only endpoints and reject unknown or mutating paths by default.

## Warning

This project is unofficial and not affiliated with Toss Securities. Internal web APIs can change or break without notice. Use it only if you understand those risks.
