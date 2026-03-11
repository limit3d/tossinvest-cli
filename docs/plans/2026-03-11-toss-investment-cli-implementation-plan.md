# Toss Investment CLI Implementation Plan

## Objective

Build a Go-first, read-only CLI for Toss Securities web data with browser-assisted login, reusable local sessions, stable export formats, and explicit safety boundaries.

## Success Criteria

- A developer can log in from the CLI with a browser flow.
- The CLI can fetch account, portfolio, orders, watchlist, and quote data.
- The CLI supports `table`, `json`, and `csv` output where relevant.
- Session reuse works across process restarts.
- The codebase blocks mutation endpoints by default.
- Documentation is good enough for another developer to reproduce setup and contribute.

## Stack Decision

Primary stack:

- Go for the CLI and client library
- Playwright-based auth helper for browser login

Fallback rule:

- if browser session extraction becomes materially easier in Python, keep Python only in `auth-helper/`
- do not move the main client or domain model out of Go without new evidence

## Milestones

### Milestone 0: Bootstrap

Goal:

- create the repository skeleton and developer workflow

Tasks:

- initialize Go module
- add CLI framework and command skeleton
- create `internal/` package layout
- add `auth-helper/` placeholder
- add `.gitignore`, `Makefile`, and basic README
- define config path and session storage path conventions

Deliverables:

- buildable `tossctl` stub
- documented local development commands

### Milestone 1: Reverse Engineering Baseline

Goal:

- understand the web protocol well enough to implement stable read-only commands

Tasks:

- capture Toss Securities web traffic for account and symbol-detail flows
- export HAR files and sanitize them
- create `docs/reverse-engineering/rpc-catalog.md`
- identify auth artifacts needed for session replay
- identify read-only endpoints for account, portfolio, orders, watchlist, and quote
- record pagination, locale, and market-specific behavior

Deliverables:

- fixture corpus
- endpoint catalog
- auth notes

Exit criteria:

- each target CLI command maps to at least one documented endpoint path

### Milestone 2: Authentication and Session Reuse

Goal:

- make login practical for developers

Tasks:

- implement `tossctl auth login`
- implement Playwright login helper
- extract and persist session state
- implement `auth status` and `auth logout`
- detect expiration and invalid session state
- prefer Keychain on macOS, with encrypted file fallback

Deliverables:

- working login flow
- reusable session store
- redacted auth debug output

Exit criteria:

- a user can log in once and run later commands without reopening the browser until the session expires

### Milestone 3: Core Read-Only Client

Goal:

- implement the internal Go client and domain models

Tasks:

- create typed client methods for each approved endpoint
- build domain models for account, position, order, watchlist item, and quote
- normalize inconsistent API fields
- add retries and defensive error classification
- enforce endpoint allowlist at the client boundary

Deliverables:

- reusable internal client package
- typed domain models
- protocol error handling

Exit criteria:

- the client can fetch all target data classes without CLI-specific logic mixed in

### Milestone 4: User Commands and Output

Goal:

- expose the first useful command set

Tasks:

- implement `account list`
- implement `account summary`
- implement `portfolio positions`
- implement `portfolio allocation`
- implement `orders list`
- implement `watchlist list`
- implement `quote get`
- implement `export` commands and shared output renderers

Deliverables:

- useful CLI command surface
- consistent rendering and exit codes

Exit criteria:

- the first release command set works end to end against a real session

### Milestone 5: Tests, Docs, and Release

Goal:

- make the project safe to publish as OSS

Tasks:

- add parser and normalization tests from fixtures
- add allowlist enforcement tests
- add local smoke-test script
- write README with risk disclosures and quick start
- add GitHub Actions for build and fixture tests
- document break-fix workflow when Toss changes internal endpoints

Deliverables:

- contributor-ready repo
- public-facing documentation
- initial tagged release

Exit criteria:

- another developer can install, authenticate, and retrieve data using the docs alone

## Work Breakdown by Area

### CLI and UX

- command structure
- shared flags
- output modes
- error messages
- account selection behavior if multiple accounts exist

### Protocol and Client

- request builder
- auth propagation
- response decoding
- pagination support
- read-only policy enforcement

### Auth

- browser login
- session extraction
- secure storage
- re-authentication path

### Developer Experience

- local commands
- fixture workflow
- test data sanitization
- release notes and troubleshooting

## Risks and Mitigations

### Risk: Toss changes web endpoints frequently

Mitigation:

- keep an RPC catalog
- maintain fixtures for every supported command
- isolate protocol bindings from command code

### Risk: Login depends on unstable browser state

Mitigation:

- keep the helper narrow
- extract the minimum viable session state
- add clear debug mode with redaction

### Risk: Hidden mutation endpoints slip into the client

Mitigation:

- allowlist read-only paths only
- test that unknown paths are rejected
- keep trading features out of the repository scope

### Risk: Go alone is insufficient for auth automation

Mitigation:

- allow a Python or Node helper
- keep that dependency isolated to login only

## Suggested Task Order

1. Bootstrap the Go repo and command skeleton.
2. Capture and sanitize the first web sessions.
3. Write the RPC catalog before implementing endpoints.
4. Implement auth login and session reuse.
5. Implement one vertical slice: `quote get`.
6. Implement account and portfolio commands.
7. Implement orders and watchlist commands.
8. Add export formats.
9. Add tests, docs, and release automation.

## Definition of Done

The project is ready for an initial OSS release when:

- the approved command set works against real Toss Securities web sessions
- session reuse is reliable on macOS
- all supported commands have fixture coverage
- mutation endpoints are blocked in code
- the README states the unofficial nature and risks of the project
- installation and login work without requiring users to inspect browser cookies manually
