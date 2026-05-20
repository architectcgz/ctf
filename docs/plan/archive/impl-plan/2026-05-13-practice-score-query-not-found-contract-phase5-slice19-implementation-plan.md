# Practice Score Query Not-Found Contract Phase 5 Slice 19 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `practice/application/queries/score_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持用户无得分记录时返回零分信息的行为不变。

**Architecture:** `practice` 新增一个局部 score query adapter。adapter 负责把 raw score repository 的 `gorm.ErrRecordNotFound` 映射成模块内 sentinel；query application 只依赖 `practice/ports` 错误契约；runtime 和手工集成 wiring 改为注入这个 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `practice/application/queries/score_service.go -> gorm.io/gorm`
- 保持 `GetUserScore` 在用户尚无得分记录时仍返回零分响应
- 不影响 `GetRanking`、score cache 和现有 ranking/user directory 查询路径

## Non-goals

- 不处理 `practice/application/commands/manual_review_service.go -> gorm.io/gorm`
- 不处理 `practice/application/commands/submission_service.go -> gorm.io/gorm`
- 不处理 `contest` 或 `challenge` 模块剩余 concrete allowlist
- 不把 `practice/infrastructure.Repository.FindUserScore` 的全局 not-found 语义整体改成 sentinel

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/practice/application/queries/score_service.go`
- `code/backend/internal/module/practice/infrastructure/score_repository.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/app/practice_flow_integration_test.go`

## Ownership Boundary

- `practice/application/queries/score_service.go`
  - 负责：score query 业务编排、缓存回退、零分响应整形
  - 不负责：知道底层 not-found 来自 `gorm`
- `practice/infrastructure/score_query_repository.go`
  - 负责：把 raw score repository 的 not-found 映射成 `practice/ports` sentinel
  - 不负责：决定上层返回什么 DTO 或是否命中缓存
- `practice/runtime/module.go` 与 `internal/app/practice_flow_integration_test.go`
  - 负责：装配 score query adapter
  - 不负责：重新把 raw GORM concrete 暴露给 query application

## Change Surface

- Add: `.harness/reuse-decisions/practice-score-query-not-found-contract-phase5-slice19.md`
- Add: `docs/plan/impl-plan/2026-05-13-practice-score-query-not-found-contract-phase5-slice19-implementation-plan.md`
- Add: `code/backend/internal/module/practice/infrastructure/score_query_repository.go`
- Add: `code/backend/internal/module/practice/infrastructure/score_query_repository_test.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/application/queries/score_service.go`
- Modify: `code/backend/internal/module/practice/application/queries/score_service_test.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/practice/application/queries/score_service_test.go`
- Add: `code/backend/internal/module/practice/infrastructure/score_query_repository_test.go`

- [ ] 为 query service 补一个 sentinel 分支测试，先证明当前实现不识别 `practice` not-found contract
- [ ] 为 score query adapter 补一个 GORM not-found 映射测试，先证明 adapter 尚不存在
- [ ] 跑最小测试，确认红灯来自目标行为缺失

验证：

- `cd code/backend && go test ./internal/module/practice/application/queries -run 'GetUserScoreTreatsPracticeScoreNotFoundAsZeroScore' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/practice/infrastructure -run 'ScoreQueryRepository' -count=1 -timeout 5m`

Review focus：

- 失败是否来自“还不认识 practice sentinel”，而不是测试夹具错误
- adapter 测试是否只约束错误映射，不引入额外业务语义

## Task 2: 实现局部 sentinel 与 adapter

**Files:**
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Add: `code/backend/internal/module/practice/infrastructure/score_query_repository.go`
- Modify: `code/backend/internal/module/practice/application/queries/score_service.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`

- [ ] 在 `practice/ports` 新增 score query not-found sentinel
- [ ] 实现 score query adapter，把 raw score repository 的 GORM not-found 映射成模块内 sentinel
- [ ] 把 score query service 改成只看模块内 sentinel，不再 import `gorm`
- [ ] 在 runtime 和手工集成 wiring 中注入 adapter

验证：

- `cd code/backend && go test ./internal/module/practice/application/queries -run 'GetUserScore|GetRanking' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/practice/infrastructure -run 'ScoreQueryRepository' -count=1 -timeout 5m`

Review focus：

- application 是否已经完全去掉 `gorm` concrete
- adapter 是否只包一层错误映射，没有复制排行榜或用户目录查询逻辑

## Task 3: 删除 allowlist 并同步文档

**Files:**
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

- [ ] 删除本 slice 实际收口的 allowlist
- [ ] 更新 phase5 当前状态与 practice 剩余 debt 描述
- [ ] 跑架构 / 文档 / workflow 完整性检查

验证：

- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `timeout 600 bash scripts/check-workflow-complete.sh`
- `git diff --check`

Review focus：

- 只删除已经真实收口的 allowlist
- 文档是否准确描述“局部 score query adapter 收口”，而不是误写成 raw score repository 全局语义改变

## Risks

- 如果 adapter 没有注入到所有 score query wiring，零分 fallback 可能在部分集成路径退化成 internal error
- 如果 application 仍残留 `gorm` 判断，allowlist 会删不干净
- 如果 integration wiring 仍直接传 raw repo，后续真实流量和 runtime wiring 可能表现不一致

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/queries -run 'GetUserScore|GetRanking' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/practice/infrastructure -run 'ScoreQueryRepository' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/practice/runtime -run '^$' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`
8. `git diff --check`

## Architecture-Fit Evaluation

- owner 明确：query application 只保留缓存回退和零分响应，adapter 承担 not-found concrete 映射
- reuse point 明确：复用 phase5 已验证的“sentinel 在 ports、映射在 infrastructure”的模式
- 不会留下“先把 gorm 藏到别的 query helper 里，下一轮再拆”的隐性返工
- touched debt 明确：这刀只收 `practice/application/queries/score_service.go`；`manual_review_service.go` 与 `submission_service.go` 仍留待后续独立 slice
