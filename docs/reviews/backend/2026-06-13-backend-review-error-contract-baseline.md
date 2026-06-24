# Backend Review: Error Contract Baseline

## Review 对象

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-error-contract-baseline`
- Branch: `task/2026-06-13-backend-error-contract-baseline`
- Task slug: `2026-06-13-backend-error-contract-baseline`
- Implementation plan: `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-contract-baseline-implementation-plan.md`
- Review target:
  - Independent gate review target: commit `2fee39e9f60cc1040e29efb861b093368feb154d`, diff basis `HEAD~1..HEAD`
  - Post-review branch sync: merge commit `1cf88c8c5` only absorbed `main` 上已落地的 task-group / API design 文档事实，不改变本 slice 的 Go 行为面
- Reviewer: independent `codex exec` read-only context

## 结论

- Classification check: 同意按 `非琐碎任务` 处理，这次改动触达 HTTP transport、challenge contract adapter、contest 跨模块 adapter、public error contract 和 architecture guardrail。
- Gate verdict: `pass`
- Material findings: 无当前未修复 material finding。
- History:
  - 首次独立 review 曾阻塞 `ContestChallengeLookupAdapter` 没有把新的 `challengecontracts.ErrChallengeNotFound` 回映射为 `contestports.ErrContestChallengeEntityNotFound`。
  - 实现上下文已补 adapter 兼容分支和 contest command 回归测试，并重跑 focused tests 与 `completion-full`。
  - 最终独立只读 re-review 返回 `pass`，未发现新的 blocker / material finding。

## Findings

- 无 blocker / material finding。
- 关键路径复核：
  - `code/backend/internal/module/challenge/infrastructure/contract_repository.go:24` 已把 `gorm.ErrRecordNotFound` 映射为 `challengecontracts.ErrChallengeNotFound`。
  - `code/backend/internal/app/router_routes.go:68` 已改为只通过 `response.FromError` 消费 public error。
  - `code/backend/internal/module/contest/infrastructure/contest_challenge_lookup_adapter.go:26` 已补齐 `challengecontracts.ErrChallengeNotFound` 到 `contestports.ErrContestChallengeEntityNotFound` 的兼容映射。
  - `code/backend/tests/architecture/test_architecture_test.go:270` 的 source guardrail 已覆盖 `internal/app` 非 composition、`internal/module/*/api` 和 `internal/middleware`，没有明显误伤合法 wiring。

## 必须重跑的验证

- 无新增必须重跑项；实现上下文提供的以下证据足够支撑 gate pass：
  - `cd code/backend && go test ./internal/apperror ./internal/module/challenge/infrastructure ./internal/module/contest/infrastructure ./internal/module/contest/application/commands ./tests/architecture -run 'Test(AppError|ContractRepository|TransportLayers|ContestChallengeLookupAdapter|ChallengeServiceAddChallengeToContestTreatsChallengeSentinelAsErrChallengeNotFound|TestContestAWDServiceSyncContestChallengeRelationTreatsChallengeSentinelAsErrChallengeNotFound)' -count=1`
  - `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`

## 残余风险

- 当前 guardrail 只约束 transport/API/middleware，不覆盖所有 application service 对 persistence/cache/runtime sentinel 的消费。
- 这个残余风险与本 slice scope 一致，后续由 `backend-redis-error-boundary`、`backend-container-runtime-error-boundary` 和 `backend-application-error-migration-core` 继续收口。

## Touched Known-Debt Status

- `internal/app/router_routes.go` 上已知的 transport 直接消费 GORM sentinel 债务已在本次收口。
- `contest` 跨模块 adapter 的 not-found 语义兼容已补齐，没有把新的 challenge public sentinel 漏成 public 500。
