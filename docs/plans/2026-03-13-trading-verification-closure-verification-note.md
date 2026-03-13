# tossinvest-cli Trading Verification Closure Verification Note

Date: 2026-03-13
Status: Live verification executed
Scope: US buy limit / KRW / non-fractional only

## Completed Checks

### Automated

- `go test ./...`
  - result: pass
- `make build`
  - result: pass

### Safe CLI Readiness

- `tossctl version`
  - result: pass
- `tossctl doctor`
  - result: pass
  - session file exists
  - trading permission file exists but temporary permission is expired
  - config file does not exist yet, so trading actions default to disabled
- `tossctl auth doctor`
  - result: pass
  - auth helper importable
  - playwright installed
  - chromium installed
  - stored session valid
- `tossctl auth status`
  - result: active session
  - provider: `playwright-storage-state`
  - live check: valid

### Read-only Order Visibility

- `tossctl orders list --output json`
  - result: pass
  - observed state: no pending orders at the time of check
- `tossctl orders completed --market us --output json`
  - result: pass
  - observed state: completed-history lookup works for current month US orders
  - observed statuses in history: `체결완료`, `취소`, `실패`
- `tossctl order show 2026-03-11/25 --market us --output json`
  - result: pass
  - observed state: canceled order lookup works through the single-order surface
- `tossctl order show 2026-03-11/1 --market us --output json`
  - result: pass
  - observed state: completed order lookup works through the single-order surface
- `tossctl order preview --symbol TSLL --market us --side buy --type limit --qty 1 --price 500 --currency-mode KRW --output json`
  - result: pass
  - observed state: preview emits canonical intent and confirm token
  - observed state: `live_ready=true`, `mutation_ready=false` while config remains disabled

## Current Blockers for Live Mutation Verification

- none for basic execution readiness

## Live Execution Results

### Live place

- command: `tossctl order place --symbol TSLL --market us --side buy --type limit --qty 1 --price 500 --currency-mode KRW --execute ...`
- result: success
- returned mutation status: `accepted_pending`
- returned order reference: `2026-03-13/1`
- follow-up:
  - `orders list` showed the order as `체결대기`
  - `order show 2026-03-13/1` also resolved the pending order

### Live amend

- command target: pending order `2026-03-13/1`, new price `700 KRW`
- first observed issue:
  - the implementation sent the user-facing order reference into `available-actions`
  - broker returned `404`
- second observed issue:
  - even with the broker raw order id, the value was not path-escaped, which also caused `404`
- code fix applied during verification:
  - resolve the raw broker order id from the pending order payload
  - path-escape that id for `available-actions`
  - treat `400` and `404` from `available-actions` as soft preflight failure so the real mutation path can continue
- post-fix live result:
  - mutation reached the broker but stopped with `interactive trade authentication required`
- conclusion:
  - `amend` is not yet end-to-end verified for this account/session
  - the path construction bug is fixed
  - the remaining blocker is broker-side interactive auth

### Live cancel

- command target: pending order `2026-03-13/1`
- result: success
- returned mutation status: `canceled`
- follow-up:
  - `orders list` became empty
  - completed history did not keep the original reference `2026-03-13/1`
  - completed history recorded the canceled order as `2026-03-13/2`
  - `order show 2026-03-13/2` resolved the canceled order successfully
  - later code change added a local lineage cache so `order show <original-id>` can follow this rollover when the mutation was executed through the same config dir

### Post-lineage live retest

- new pending place:
  - `TSLL` 1주 `500 KRW`
  - result: `accepted_pending`
  - returned order reference: `2026-03-13/3`
- amend retest:
  - target: `2026-03-13/3`
  - new price: `700 KRW`
  - result: still blocked by `interactive trade authentication required`
- cancel retest:
  - target: `2026-03-13/3`
  - immediate result: `canceled`
  - immediate reconciliation: pending disappeared, but completed row was not yet visible so no `current_order_id` was returned
  - delayed reconciliation:
    - `orders completed --market us` later showed the canceled order as `2026-03-13/4`
    - this gap is now addressed in code by storing unresolved cancel recovery hints and retrying completed-history lookup on `order show <old-id>`
- conclusion:
  - `amend` remains blocked by broker-side interactive auth for this account/session
  - cancel rollover can appear later than the mutation reconciliation window
  - delayed cancel rollover now has a same-machine on-demand recovery path, but ambiguous candidates still require manual inspection

## Code Changes Found Necessary During Live Verification

- `GetOrderAvailableActions` now uses the resolved raw pending-order id instead of the user-facing reference id
- the broker raw id is path-escaped before calling `available-actions`
- `400` and `404` responses from `available-actions` are treated as soft failures because cancel/amend do not consume that payload today
- interactive-auth user-facing text now refers to the generic trade action instead of saying "cancel" unconditionally
- regression tests added for:
  - resolved broker order id path construction
  - path escaping of raw order ids
  - soft-failure handling for `400`
  - cancel completed-history rollover reconciliation
  - local alias lookup for `order show`
  - delayed cancel rollover recovery from lineage hints
  - ambiguous delayed cancel rollover rejection

## Post-Run Safety State

- trading permission revoked
- local `config.json` restored to all trading flags disabled

## Funding Guidance Live Retest

- focused live validation target:
  - verify that `order place` now shows funding-specific operator guidance instead of a generic prepare failure
- live inputs:
  - `TSLL` 1주 `500 KRW`
  - `TSLL` 1주 `1000 KRW`
- observed broker rejection:
  - both commands stopped at `prepare`
  - broker message: `환전에 필요한 원화 출금가능금액이 부족합니다.`
- initial result:
  - the first implementation treated this message as a generic prepare rejection because the wording did not match the earlier `계좌 잔액이 부족해요` fixture
- follow-up fix:
  - classifier now treats `원화 출금가능금액 부족` wording as `funding_required`
- live rerun after the fix:
  - CLI showed funding-specific step-by-step guidance
  - message included:
    - cause summary: `주문 준비 단계에서 잔액 또는 주문가능금액이 부족해 진행이 중단되었습니다.`
    - broker message echo
    - retry preview command
    - retry place command template with `--confirm <new-confirm-token>`
- safety outcome:
  - no pending order was created
  - trading permission revoked again after verification
  - local `config.json` restored to disabled state again

## Post-Funding Live Retest

- operator prepared `600 KRW` of orderable / withdrawable KRW before the rerun
- live inputs:
  - `TSLL` 1주 `500 KRW`
- observed result:
  - `order place` no longer stopped at funding guidance
  - result: `accepted_pending`
  - pending order reference: `2026-03-13/5`
- interpretation:
  - the funding gate was cleared for this low-price pending order
  - no `fx_consent_required` branch appeared on this specific input
- safety follow-up:
  - the pending order was canceled immediately
  - immediate cancel result still ended with the known warning:
    - `Pending order disappeared, but the canceled completed-history row is not visible yet.`
- delayed rollover live result:
  - `orders completed --market us` later showed the canceled row as `2026-03-13/6`
  - the first `order show 2026-03-13/5` retry still failed
  - root cause:
    - completed-history rows used `version` as the meaningful recency timestamp
    - `version` arrived as a timezone-free local timestamp string
    - some completed rows also had `symbol: null`, so matching had to fall back to `stockName`
  - after fixing the parser and rerunning read-only verification:
    - `order show 2026-03-13/5 --market us` resolved successfully to `2026-03-13/6`
    - `resolved_from_id` was populated correctly
- safety outcome:
  - pending orders returned to `[]`
  - trading permission revoked again after verification
  - local `config.json` restored to disabled state again

## Thousand-KRW Live Retest

- operator requested a rerun with `TSLL` 1주 `1000 KRW` because the mobile app was showing a foreign-exchange prompt on the same general path
- live preflight state:
  - account summary showed `orderable_amount_krw=1090`
  - US market summary showed `orderable_amount_krw=503`, `orderable_amount_usd=0.34`
  - pending orders were `[]`
- live inputs:
  - `TSLL` 1주 `1000 KRW`
- observed result:
  - `order preview` returned `mutation_ready=true`
  - `order place` succeeded as `accepted_pending`
  - pending order reference: `2026-03-13/7`
- interpretation:
  - this CLI / broker path still did not surface `fx_consent_required`
  - despite the requested notional being above the reported US-market KRW buying-power snapshot, the broker accepted the pending order without an FX-consent stop during placement
  - the currently observed behavior is therefore closer to `pending orders may be accepted first, with FX handling deferred outside this specific placement branch`
- safety follow-up:
  - the pending order was canceled immediately
  - cancel resolved directly to a new completed-order reference: `2026-03-13/8`
  - `order show 2026-03-13/7 --market us` resolved successfully to `2026-03-13/8`
- safety outcome:
  - pending orders returned to `[]`
  - trading permission revoked again after verification
  - local `config.json` restored to disabled state again

## Desktop Web FX Prompt Investigation

- investigation goal:
  - determine whether the missing CLI `fx_consent_required` branch is a direct broker rejection or a later web-only confirmation step
- method:
  - reused the stored Toss web session inside a headed Playwright browser
  - opened `https://www.tossinvest.com/stocks/US20220809012/order`
  - replayed the same `TSLL` 1주 `1000 KRW` order inputs already tested via CLI
- observed web sequence:
  1. first `구매하기` click
     - `POST /api/v2/wts/trading/order/prepare` returned `200`
     - prepare response included `preparedOrderInfo.needExchange: 0.68`
     - UI showed only the standard confirmation dialog `TSLL 구매 1주`
  2. confirmation-dialog `구매` click
     - browser fetched:
       - `GET /api/v1/trading/settings/toggle/find?categoryName=GETTING_BACK_KRW`
       - `GET /api/v1/exchange/current-quote/for-buy`
     - UI then showed the FX prompt:
       - `0.68달러가 부족해요`
       - `주식 구매를 위해 환전할게요`
       - `주문이 취소되면 계좌에는 달러로 남아있어요.`
- important negative evidence:
  - no `order/create` request was observed before the FX modal
  - the FX prompt therefore appears after successful `prepare`, not as a direct `prepare` rejection
- follow-up capture on FX-modal `확인`:
  - `POST /api/v2/wts/trading/order/create`
  - `POST /api/v1/trading/settings/toggle` with `{"categoryName":"EXCHANGE_INFO_CHECK","turnedOn":true}`
  - no separate FX-confirm mutation was observed
- implication:
  - the current CLI gap is not only missing prepare-failure classification
  - the web product has a second-stage FX confirmation branch after `prepare` succeeds
  - CLI parity requires either:
    - surfacing a post-prepare `fx_consent_required` stop, or
    - explicitly auto-consuming the branch before `create`
- safety outcome:
  - the FX modal was canceled in the browser
  - no order was created
  - no pending order remained
  - local screenshot saved to `output/playwright/fx-consent-web-prompt-tsll-1000krw.png`

## Post-Prepare FX Branch CLI Verification

- implementation change:
  - CLI now reads `preparedOrderInfo.needExchange`
  - when `needExchange > 0`, CLI fetches:
    - `GET /api/v1/trading/settings/toggle/find?categoryName=GETTING_BACK_KRW`
    - `GET /api/v1/exchange/current-quote/for-buy`
  - CLI stops before `order/create` and returns a post-prepare `fx_consent_required` branch
- live verification on 2026-03-13:
  - input: `TSLL` 1주 `1000 KRW`
  - account state before run:
    - `orderable_amount_krw=1591`
    - `orderable_amount_usd=0`
    - US market buying power remained `0`
  - `order preview` still returned `mutation_ready=true`
  - `order place` no longer reached `accepted_pending`
  - CLI output showed:
    - `주문 준비는 통과했지만, 웹과 동일한 환전 확인 단계에서 중단되었습니다.`
    - `0.68달러가 부족해요.`
    - `예상 환전 금액: 1,022원`
    - `예상 환율: 1,503.63원/USD`
    - `주의: 주문이 취소되면 계좌에는 달러로 남아있어요.`
- important outcome:
  - the CLI now stops on the same branch family that the desktop web flow exposed
  - no pending order was created
  - `orders list` remained `[]`
  - trading permission was revoked again after verification
  - local `config.json` was restored to the disabled state again

## FX Consent Automation Capture

- browser capture on 2026-03-13:
  - FX-modal `확인` did not call a dedicated consent endpoint
  - instead the browser issued:
    - `POST /api/v2/wts/trading/order/create`
    - `POST /api/v1/trading/settings/toggle` with `EXCHANGE_INFO_CHECK=true`
- implementation outcome:
  - `dangerous_automation.accept_fx_consent=true` now means:
    - keep the post-prepare read calls
    - continue through `order/create`
    - then best-effort mark `EXCHANGE_INFO_CHECK=true`

## FX Consent Automation Live Verification

- config used only for this retest:
  - `schema_version: 2`
  - `trading.grant=true`
  - `trading.place=true`
  - `trading.cancel=true`
  - `trading.allow_live_order_actions=true`
  - `trading.dangerous_automation.accept_fx_consent=true`
- live place on 2026-03-13:
  - input: `TSLL` 1주 `1000 KRW`
  - `order preview` returned confirm token `9c80a781bae7`
  - `order place` reached `accepted_pending` instead of stopping on FX consent
  - created pending ref: `2026-03-13/9`
- live cleanup:
  - `order cancel` succeeded
  - cancel preview token: `46f51d16e687`
  - surviving completed ref: `2026-03-13/10`
  - `order show 2026-03-13/9 --market us` resolved to `2026-03-13/10`
  - `orders list` returned `[]`
- safety cleanup:
  - trading permission was revoked again
  - external `config.json` was restored to the original disabled `schema_version: 1` state

## Still Pending

- live re-test of `order amend` after lineage/reconciliation changes
- evidence-driven confirmation of whether the observed interactive-auth branch for `amend` is account-specific or generally expected

## Next Operator Steps

1. Keep `order show <old-id>` and delayed cancel rollover verification in the regular regression loop.
2. Live re-test `order amend` again only if broker-side interactive auth can be satisfied or explicitly automated.
3. Decide whether `complete_trade_auth` deserves the same evidence-first automation treatment next.
