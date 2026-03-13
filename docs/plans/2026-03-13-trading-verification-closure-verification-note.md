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

## Still Pending

- live re-test of `order amend` after lineage/reconciliation changes
- evidence-driven confirmation of whether the observed interactive-auth branch for `amend` is account-specific or generally expected

## Next Operator Steps

1. Run full `go test ./...` after the live-driven fixes.
2. Live re-test `order show <old-id>` on a delayed cancel rollover after the new on-demand recovery path lands.
3. Live re-test `order amend` again only if broker-side interactive auth can be satisfied or explicitly automated.
