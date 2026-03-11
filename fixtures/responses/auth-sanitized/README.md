# Authenticated Sanitized Fixtures

These fixtures are fetched from an authenticated Toss Securities web session and sanitized before commit.

Refresh them with:

```bash
python3 tools/fetch_auth_fixtures.py output/playwright/tossinvest-account-state.json fixtures/responses/auth-sanitized
```

Rules:

- never commit raw storage state
- never commit raw account numbers, cookies, or tokens
- inspect diffs before commit when fixture values change

