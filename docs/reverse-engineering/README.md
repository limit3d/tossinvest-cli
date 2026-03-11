# Reverse Engineering Notes

This directory will hold the Toss Securities web protocol research artifacts used to implement the read-only client.

Planned contents:

- RPC catalog
- auth notes
- breakage log
- fixture sanitization rules
- public fixture refresh workflow

Do not commit raw captures with sensitive cookies, tokens, account numbers, or personal data.

Useful scripts:

- `python3 tools/sanitize_har.py <input.har> <output.har>`
- `python3 tools/fetch_public_fixtures.py fixtures/responses/public`
