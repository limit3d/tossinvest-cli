# tossinvest-cli Dangerous Automation Config Design

Date: 2026-03-13
Status: Approved
Scope: config surface, legacy compatibility, and doctor/config UX only

## Goal

`config.json`에서 live trading 관련 위험 설정이 이름만 보고도 의미가 보이도록 정리한다.

이번 단계의 목적은 두 가지다.

- `allow_dangerous_execute`처럼 의미가 불분명한 이름을 명확한 이름으로 교체한다.
- `doctor`와 `config show`가 raw key를 나열하는 대신, 실제로 어떤 위험한 행동이 허용되는지 사람 말로 설명하게 만든다.

## In Scope

- `schema_version: 2`
- `allow_dangerous_execute` -> `allow_live_order_actions`
- `dangerous_automation` config 공간 추가
- 첫 버전 dangerous automation fields
  - `complete_trade_auth`
  - `accept_product_ack`
- legacy v1 config loading compatibility
- `config show` UX 개선
- `doctor` UX 개선

## Out of Scope

- 브로커 분기 자동화 handler 구현
- CLI flag rename
- funding / FX consent branch automation

## Recommended Config Shape

```json
{
  "$schema": "https://raw.githubusercontent.com/JungHoonGhae/tossinvest-cli/main/schemas/config.schema.json",
  "schema_version": 2,
  "trading": {
    "grant": false,
    "place": false,
    "cancel": false,
    "amend": false,
    "allow_live_order_actions": false,
    "dangerous_automation": {
      "complete_trade_auth": false,
      "accept_product_ack": false
    }
  }
}
```

## Naming Rationale

- `allow_live_order_actions`
  - 실계좌에 영향을 주는 주문 액션 허용 여부가 이름만 보고 드러난다.
- `dangerous_automation`
  - 단순한 automation이 아니라 위험한 자동 진행 설정이라는 점이 바로 보인다.
- `complete_trade_auth`
  - trade auth 분기를 "자동으로 끝까지 진행한다"는 행동이 이름에 들어간다.
- `accept_product_ack`
  - acknowledgement를 자동 수락한다는 점이 드러난다.

## Legacy Compatibility

- 기존 `schema_version: 1` 파일은 계속 읽는다.
- 기존 `trading.allow_dangerous_execute`는 내부에서 `trading.allow_live_order_actions`로 해석한다.
- `config init`은 앞으로 v2 모양만 생성한다.
- `config show`와 `doctor`는 새 이름을 기준으로 보여주고, legacy 변환이 일어나면 그 사실만 따로 알려준다.

## Doctor UX

`doctor`는 단순한 raw boolean 출력이 아니라, 현재 어떤 위험한 행동이 가능한지 해석해서 보여줘야 한다.

핵심 check:

- `trading_config`
  - 어떤 trading actions가 열려 있는지
- `live_order_actions`
  - 실계좌 주문 액션이 막혀 있는지 / 열려 있는지
- `dangerous_automation`
  - 어떤 위험한 자동 분기가 켜져 있는지
- `legacy_config`
  - v1 key를 변환해서 읽고 있는지

## Initial Dangerous Automation Scope

첫 버전은 아래 두 분기만 config에 노출한다.

- `complete_trade_auth`
- `accept_product_ack`

이유:

- trade auth는 live verification에서 실제로 관찰됐다.
- product acknowledgement는 discovery 문서에서 계속 등장한다.
- funding, FX consent는 아직 naming과 자동화 경계가 덜 닫혔다.

## Success Criteria

- 새 사용자는 `allow_live_order_actions`와 `dangerous_automation`만 보고도 의미를 이해할 수 있다.
- 기존 v1 config는 깨지지 않는다.
- `doctor`가 위험한 설정의 의미를 문장으로 설명한다.
- `config show`가 raw field 나열이 아니라 실행 정책을 더 명확히 드러낸다.
