# 本地架构 Guardrail 补强 Review

## Review Target

- Repository: `ctf`
- Diff source: local staged candidate for `code/backend/internal/module/architecture_test.go`, `code/frontend/src/__tests__/architectureBoundaries.test.ts`, `scripts/check-architecture.sh`, `docs/architecture/README.md`, `docs/plan/impl-plan/2026-05-08-local-architecture-guardrails-expansion-plan.md`
- Classification: non-trivial harness / architecture guardrail change

## Gate Verdict

Pass with residual review-process risk.

## Findings

No material implementation findings found in same-context review.

## Assessment

The implementation uses baseline allowlists rather than broad rule exceptions. This is the correct shape for the current repository because backend application code and frontend legacy component directories already contain historical coupling. The tests now prevent new coupling while allowing existing debt to be reduced safely; stale allowlist checks force cleanup when entries disappear.

The main tradeoff is allowlist size. It is acceptable for this slice because the allowlist makes current debt explicit and executable. If the lists keep growing, the next maintenance step should split allowlists into dedicated files or generated fixtures.

## Required Re-validation

- `bash scripts/check-architecture.sh --full`
- `bash scripts/check-consistency.sh`
- `cd code/backend && go test ./internal/module`
- `npx --prefix code/frontend eslint --quiet code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `npx --prefix code/frontend prettier --check code/frontend/src/__tests__/architectureBoundaries.test.ts docs/architecture/README.md docs/plan/impl-plan/2026-05-08-local-architecture-guardrails-expansion-plan.md`
- `git diff --check`

## Residual Risk

This review was performed in the same agent context because no explicit subagent delegation was requested in this turn. It does not satisfy a fully independent subagent review gate. The remaining risk is review-process independence, not an observed implementation defect.

## Touched Known Debt

The change touches existing architecture debt surfaces by recording them as baselines:

- backend application direct GORM / Redis / HTTP usage
- backend domain dependency on internal DTO/model/config
- backend cross-module private imports
- frontend legacy `components/*Page.vue`
- frontend components/widgets direct API and feature imports

The debt is not removed in this slice; it is constrained from expanding.
