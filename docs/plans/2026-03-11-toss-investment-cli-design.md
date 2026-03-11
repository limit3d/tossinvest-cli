# Toss Investment CLI Design

## Summary

This project aims to build an unofficial, read-only CLI for Toss Securities that feels as easy to use as `notebooklm-py`, but is shaped for the Toss Securities web product and for a Go-first ecosystem.

The first release targets developers, not general consumers. It will support browser-based login, local session reuse, and simple commands for account, portfolio, orders, watchlist, quote, and export workflows. The CLI must stay read-only by design.

## Goals

- Provide a simple CLI for Toss Securities web data.
- Make the tool scriptable with `table`, `json`, and `csv` output.
- Keep the main implementation in Go where practical.
- Use a browser-assisted login flow so users do not need to extract cookies by hand.
- Publish the project as an OSS developer tool with clear warnings about unofficial API risk.

## Non-Goals

- Trade placement, order cancelation, or order modification.
- Full mobile app reverse engineering.
- Real-time streaming in the first release.
- A polished consumer-grade product with full support obligations.

## Constraints and Assumptions

- Reverse engineering will target the Toss Securities web experience first.
- The first release will be read-only and developer-oriented.
- The main CLI should be Go if the protocol is practical to implement there.
- A helper written in Node or Python is acceptable if browser login or session extraction is hard to do well in Go.
- The tool depends on undocumented APIs that may change without notice.

## Reference Model

The usability target is the pattern used by `notebooklm-py`:

- clear warning that the project is unofficial
- browser-assisted login for first-time setup
- simple command model
- reusable local session storage
- both human-friendly and automation-friendly output

This project should copy that product shape, not its language stack.

## Recommended Architecture

The recommended design is a Go CLI with a small browser login helper.

### 1. Main CLI

`tossctl` is the user-facing command. It should be a single Go binary built with a standard CLI library such as `cobra`.

Responsibilities:

- parse commands and flags
- load config and sessions
- call the internal client
- render output as `table`, `json`, or `csv`

### 2. Core Client

The Go client library handles authenticated HTTP calls, request shaping, response parsing, retries, and normalization into stable domain models.

Responsibilities:

- manage HTTP sessions and request headers
- map internal Toss Securities endpoints to typed client methods
- normalize API responses into stable models
- hide protocol details from CLI commands

### 3. Session Store

The session layer stores credentials and session artifacts outside the command layer.

Preferred strategy:

- macOS: Keychain first
- fallback: encrypted local file

The store should save only what is required to rehydrate a valid read-only session.

### 4. Auth Helper

The auth helper runs only during login or re-authentication. It opens a real browser, lets the user log in, and extracts the minimum session state needed by the Go client.

Preferred implementation:

- Playwright helper
- Node or Python runtime acceptable

Hard rule:

- keep all business logic and API bindings out of the helper
- the helper acquires session state only

## Command Surface

The first release should expose a small, stable command set:

```text
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

Design rules:

- nouns should reflect domain boundaries
- commands should read cleanly in shell scripts
- `--output table|json|csv` should be supported consistently
- export commands may wrap common data pulls, but should not invent a second API model

## Reverse Engineering Workflow

The project should reverse engineer the web product screen by screen instead of guessing endpoints.

### Capture Strategy

- use browser devtools, HAR export, or Playwright-assisted capture
- focus first on account and symbol-detail views
- catalog all requests involved in read-only data retrieval
- sanitize and store captured responses as reusable fixtures

### RPC Catalog

Before writing client code, create an endpoint catalog with:

- path
- method
- request headers
- auth dependency
- pagination behavior
- response shape
- break risk
- mapped domain capability

This catalog becomes the source of truth for implementation.

### Target Capabilities for Phase 1

- account list and summary
- portfolio positions and allocation
- orders history
- watchlist
- quote lookup

## Security and Legal Boundaries

This project must be read-only in both policy and code.

Required protections:

- enforce an allowlist of read-only endpoints
- block all mutation paths by default
- never log tokens, cookies, account numbers, or raw personal data
- redact sensitive fields in fixtures and debug artifacts
- document clearly that the project is unofficial and may violate user expectations or platform rules if misused

The README should state:

- unofficial project
- no affiliation with Toss Securities
- APIs may change or break
- use at your own risk
- intended for research, automation, and developer workflows

## Testing Strategy

The design depends on two test layers.

### Fixture Tests

Use sanitized captured responses to test:

- parsers
- response normalization
- output formatting
- allowlist enforcement

### Smoke Tests

Run small local end-to-end checks against a real account for:

- `auth status`
- `account summary`
- `quote get`

These tests should not run in public CI.

## Release Shape

The project should ship first as:

- a Go CLI
- an optional auth helper runtime
- documentation for macOS-first setup

Support order:

1. macOS
2. Linux if session storage and browser helper behavior are reliable
3. Windows later if needed

## Initial Repository Layout

```text
cmd/tossctl/
internal/auth/
internal/client/
internal/session/
internal/domain/
internal/output/
internal/export/
internal/replay/
auth-helper/
fixtures/har/
fixtures/responses/
docs/reverse-engineering/
docs/plans/
README.md
```

## Exit Criteria for the First Release

The first release is complete when a developer can:

1. install `tossctl`
2. run `tossctl auth login`
3. reuse the saved session
4. fetch account, portfolio, orders, watchlist, and quote data
5. export at least positions and orders as `json` or `csv`
6. understand the risks and limits from the README alone

## References

- `notebooklm-py`: <https://github.com/teng-lin/notebooklm-py>
- Toss Securities web: <https://www.tossinvest.com/>
