# tossinvest-cli Dangerous Automation Config Implementation Plan

Date: 2026-03-13
Status: Drafted from approved design
Scope: config surface, legacy compatibility, and doctor/config UX only

## Objective

위험한 live trading 설정을 더 명확한 이름으로 재구성하고, `doctor`와 `config show`에서 그 의미를 사람이 바로 이해할 수 있게 만든다.

## Phase 1. Config Model and Compatibility

### Work

- `schema_version`을 `2`로 올린다.
- `allow_live_order_actions`를 새 canonical field로 추가한다.
- `dangerous_automation` object를 추가한다.
- 기존 `allow_dangerous_execute`를 계속 읽되 내부 canonical field로 변환한다.

### Deliverables

- updated `internal/config`
- updated JSON schema
- config load/init tests

## Phase 2. Doctor and Config Output

### Work

- `config show`에서 새 canonical field를 표시한다.
- `doctor`에 아래 해석 check를 추가한다.
  - enabled trading actions
  - live order actions meaning
  - dangerous automation meaning
  - legacy config translation

### Deliverables

- updated `internal/output/config.go`
- updated `internal/doctor`
- doctor tests if useful

## Phase 3. Docs

### Work

- `README.md`
- `docs/configuration.md`
- 필요한 경우 `docs/architecture.md`

### Deliverables

- user-facing docs aligned to new naming

## Verification

- `go test ./...`
- `make build`
- `tossctl config init`
- `tossctl config show`
- `tossctl doctor`

## Risks

- old config files may still point to the same schema URL while using v1 fields
- dangerous automation config may appear stronger than current runtime support if docs are vague

## Expected Outcome

사용자는 config와 doctor만 보고도 아래를 이해할 수 있다.

- 어떤 trading actions가 열려 있는가
- 실계좌 주문 액션이 가능한가
- 어떤 위험한 자동 분기가 허용되는가
- 현재 config가 legacy translation 위에서 동작 중인가
