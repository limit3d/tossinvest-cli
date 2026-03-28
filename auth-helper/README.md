# auth-helper

This directory contains the Python-based browser login helper used by `tossctl auth login`.

Responsibilities:

- open a real browser session
- let the user complete Toss Securities web login
- extract the minimum session state needed by the Go client
- return sanitized session payloads to `tossctl`

Hard boundary:

- no Toss Securities domain logic
- no read-only client bindings
- no trading logic

The helper exists to isolate browser automation from the main CLI.

## CLI

```bash
cd auth-helper
python3 -m pip install -e .
python3 -m tossctl_auth_helper login --storage-state /tmp/tossctl-storage-state.json
```

> Google Chrome이 시스템에 설치되어 있어야 합니다. `playwright install chromium`은 더 이상 필요하지 않습니다.

The helper emits JSON on stdout. On success it returns `status=ok` and the path of the saved Playwright storage-state file.
