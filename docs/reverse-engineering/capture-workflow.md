# Toss Securities Web Capture Workflow

Verified against the public web surface on 2026-03-11.

## Goal

Capture Toss Securities web traffic in a way that is:

- repeatable
- safe to store
- useful for a read-only CLI
- explicit about what is still unknown

This workflow is for reverse engineering the web product, not for bypassing login or automating trading.

## Scope

Priority screens for Milestone 1:

1. `/account`
2. a stock detail page such as `/stocks/A005930/order`
3. watchlist and holdings views after manual login

The first pass should focus on read-only data only:

- account summary
- positions
- orders history
- watchlist
- quotes

## Safe Workflow

### 1. Start with public pages

Use public pages to identify:

- API hostnames
- bootstrap endpoints
- stock-detail endpoints
- chart and ranking endpoints
- login page redirects

This work can happen before any authenticated capture.

### 2. Use a clean browser profile

When capturing authenticated traffic:

- use a fresh browser profile
- log in only with a test or personal account you control
- avoid keeping other tabs open
- avoid mixing mobile and web login flows in the same capture

### 3. Capture one screen at a time

For each screen:

- open the screen
- wait for data to settle
- export a HAR
- record the screen name and date
- describe the user-visible action that triggered the requests

Good capture units:

- account overview page load
- positions tab load
- orders history tab load
- watchlist page load
- symbol detail page load

### 4. Sanitize before commit

Never commit raw captures.

Store raw files outside git tracking, then sanitize them into `fixtures/har/` or convert them into smaller JSON fixtures under `fixtures/responses/`.

Use:

```bash
python3 tools/sanitize_har.py path/to/raw.har fixtures/har/account-overview.har
```

### 5. Build the RPC catalog first

Before writing Go client code:

- list the endpoint
- classify it as public, guest, or authenticated
- capture the request method
- note required params and headers
- map it to a CLI capability

## What to Record Per Endpoint

- hostname
- method
- path
- query parameters
- auth requirement
- request body shape if present
- response top-level shape
- whether the endpoint appears safe for read-only use
- which CLI command depends on it

## Redaction Rules

Always remove or replace:

- `cookie`
- `set-cookie`
- `authorization`
- `x-csrf-token`
- `x-xsrf-token`
- `x-device-id`
- `x-session-id`
- `x-request-id`
- `phoneNumber`
- `name`
- `residentRegistrationNumber`
- account numbers
- order IDs that can be tied to a real user
- comments or community text tied to a logged-in identity

Masking rule:

- preserve structure
- replace secrets with stable placeholders such as `<REDACTED_COOKIE>`
- do not delete entire objects unless there is no safe way to keep their shape

## Public Observations Captured on 2026-03-11

Public web navigation exposed these routes:

- `/`
- `/feed/recommended`
- `/screener`
- `/account`
- `/signin?redirectUrl=%2Faccount`

Visiting `/account` without an authenticated session redirected to `/signin?redirectUrl=%2Faccount`.

Observed API hostnames:

- `wts-api.tossinvest.com`
- `wts-info-api.tossinvest.com`
- `wts-cert-api.tossinvest.com`
- `cdn-api.tossinvest.com`
- `tuba-static.tossinvest.com`
- `log.tossinvest.com`

Observed public or guest-accessible endpoints included:

- `GET /api/v3/init`
- `GET /api/v1/time`
- `GET /api/v1/user-setting`
- `GET /api/v2/system/trading-hours/integrated`
- `GET /api/v1/dashboard/wts/overview/trading-info`
- `GET /api/v1/dashboard/wts/overview/exchange-rates`
- `GET /api/v1/rankings/realtime/stock`
- `GET /api/v2/stock-infos/{code}`
- `GET /api/v1/stock-detail/ui/{code}/common`
- `GET /api/v1/c-chart/...`
- `GET /api/v1/product/stock-prices`

These observations are enough to start a read-only catalog. Authenticated captures are still needed for account, holdings, and order history.

## Next Milestone 1 Outputs

- `rpc-catalog.md`
- `auth-notes.md`
- first sanitized HAR captures
- small JSON fixtures for stock detail and quotes

