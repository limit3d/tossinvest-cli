# TODOS

## Reconciliation market 파라미터화

- **What:** `internal/client/trading.go`의 reconcile 함수들에 하드코딩된 `Market: "us"` 8곳을 intent/order에서 전달받도록 파라미터화
- **Why:** 국내주식(kr) 지원 시 reconciliation이 올바른 market을 사용해야 함. 현재는 US만 지원하므로 동작에 영향 없지만, 국내주식 확장의 전제 조건.
- **Pros:** 국내주식 확장 시 변경량 감소, reconcile 함수의 범용성 확보
- **Cons:** 현재 US-only에서는 기능적 변화 없음, 테스트 8곳 수정 필요
- **Context:** sell order 디자인 리뷰에서 스코프 밖으로 이연됨 (2026-03-21). `reconcilePlacedOrder`, `reconcileAmendedOrder`, `reconcileCanceledOrder` 및 관련 `findMatchingCompletedOrder` 호출부. 10x 비전(full trading: 시장가 + 국내주식)의 일부.
- **Depends on:** sell order 구현 완료 후 진행 권장 (동시 변경 시 diff 복잡)

## Sell 주문 live 검증 확대

- **What:** 분할 매도(일부 수량), 전량 매도, 보유량 초과 요청 시 API 동작 검증
- **Why:** 분할 매도는 sell의 일반 케이스. prerequisite dry-run에서 기본 happy-path는 확인하지만, 수량 관련 edge case는 별도 live 검증 필요.
- **Pros:** sell 기능의 신뢰도 확보, 사용자가 예상치 못한 에러 방지
- **Cons:** 실제 보유 주식이 있어야 검증 가능 (테스트 비용)
- **Context:** Codex 리뷰에서 "partial sell is the normal case" 지적 (2026-03-21). FX consent 방향(USD→KRW), auth-required 분기, holdings rejection 등도 함께 검증 권장.
- **Depends on:** sell order 구현 + 첫 live sell 검증 완료 후
