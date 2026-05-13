# Practice Manual Review Not-Found Contract Phase 5 Slice 20 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `practice/application/commands/manual_review_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持人工评阅、教师列表/详情查询与“我的提交记录”查询的业务行为不变。

**Architecture:** `practice` 新增一个局部 manual review adapter，把 raw practice repository 的人工评阅记录 / 已通过提交 / 用户 not-found 映射成模块内 sentinel；challenge not-found 复用现有 `runtime_subject_repository` adapter。application command service 只依赖 `practice/ports` 错误契约，runtime / app composition 负责装配 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `practice/application/commands/manual_review_service.go -> gorm.io/gorm`
- 保持人工评阅详情不存在时仍返回 `errcode.ErrNotFound`
- 保持 challenge 不存在时仍返回 `errcode.ErrChallengeNotFound`
- 保持教师用户不存在时仍按当前语义返回 `errcode.ErrUnauthorized`

## Non-goals

- 不处理 `practice/application/commands/submission_service.go -> gorm.io/gorm`
- 不改 `practice/infrastructure.Repository` 的全局 not-found 语义
- 不把 `runtime_subject_repository` 重命名或做大范围职责调整

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/practice/runtime/module.go`

## Ownership Boundary

- `practice/application/commands/manual_review_service.go`
  - 负责：人工评阅业务编排、errcode 映射、权限检查、零/存在分支语义
  - 不负责：知道底层 not-found 来自 `gorm`
- `practice/infrastructure/manual_review_repository.go`
  - 负责：把 raw practice repository 的人工评阅相关 not-found 映射成 `practice/ports` sentinel
  - 不负责：决定上层该返回哪个 `errcode`
- `practice/infrastructure/runtime_subject_repository.go`
  - 负责：继续把 challenge not-found 映射成 `ErrPracticeChallengeNotFound`
  - 不负责：承担人工评阅记录或用户查询语义
- `practice/runtime/module.go` / `internal/app/practice_flow_integration_test.go`
  - 负责：注入 manual review adapter 和 runtime subject adapter
  - 不负责：把 raw GORM concrete 重新带回 command service

## Change Surface

- Add: `.harness/reuse-decisions/practice-manual-review-not-found-contract-phase5-slice20.md`
- Add: `docs/plan/impl-plan/2026-05-13-practice-manual-review-not-found-contract-phase5-slice20-implementation-plan.md`
- Add: `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`
- Add: `code/backend/internal/module/practice/infrastructure/manual_review_repository_test.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- Add: `code/backend/internal/module/practice/infrastructure/manual_review_repository_test.go`

- [ ] 为 manual review service 补一个 sentinel 分支测试，先证明当前实现还不识别 practice manual-review not-found contract
- [ ] 为 manual review adapter 补 GORM not-found 映射测试，先证明 adapter 尚不存在
- [ ] 跑最小测试，确认红灯来自目标行为缺失

验证：

- `cd code/backend && go test ./internal/module/practice/application/commands -run 'GetTeacherManualReviewSubmissionTreatsPracticeManualReviewSubmissionNotFoundAsNotFound|ListMyChallengeSubmissionsTreatsPracticeChallengeNotFoundAsChallengeNotFound' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/practice/infrastructure -run 'ManualReviewRepository' -count=1 -timeout 5m`

Review focus：

- service 测试是否真正在约束模块内 sentinel，而不是继续借 GORM sentinel 过关
- adapter 测试是否只覆盖错误映射，不夹带业务判断

## Task 2: 实现局部 sentinel 与 adapter

**Files:**
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Add: `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`

- [ ] 在 `practice/ports` 新增 manual review 相关 not-found sentinel 与窄 repository port
- [ ] 实现 manual review adapter，把 raw repo 的 submission/user not-found 映射成模块内 sentinel
- [ ] 让 `manual_review_service.go` 改成只看 practice sentinel，并复用 runtime subject adapter 处理 challenge not-found
- [ ] 在 runtime wiring、app wiring 和测试 helper 中注入 adapter

验证：

- `cd code/backend && go test ./internal/module/practice/application/commands -run 'ReviewManualReviewSubmission|ListTeacherManualReviewSubmissions|GetTeacherManualReviewSubmission|ListMyChallengeSubmissions' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/practice/infrastructure -run 'ManualReviewRepository|RuntimeSubjectRepository' -count=1 -timeout 5m`

Review focus：

- application surface 是否已经完全去掉 `gorm` concrete
- manual review adapter 是否仍然窄，只承接人工评阅相关 not-found 映射

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
- 文档是否准确描述“manual review adapter + runtime subject adapter”的 owner 分工

## Risks

- 如果 manual review adapter 注入不全，人工评阅详情/列表路径可能退化成 internal error
- 如果 `ListMyChallengeSubmissions` 没切到 runtime subject adapter，文件里的 challenge not-found concrete 会残留
- 如果测试 helper 仍直接构造未注入 adapter 的 service，相关测试会在 nil surface 上误失败

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/commands -run 'ReviewManualReviewSubmission|ListTeacherManualReviewSubmissions|GetTeacherManualReviewSubmission|ListMyChallengeSubmissions' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/practice/infrastructure -run 'ManualReviewRepository|RuntimeSubjectRepository' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/practice/runtime -run '^$' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/app -run '^$' -count=1 -timeout 5m`
5. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
6. `python3 scripts/check-docs-consistency.py`
7. `bash scripts/check-consistency.sh`
8. `timeout 600 bash scripts/check-workflow-complete.sh`
9. `git diff --check`

## Architecture-Fit Evaluation

- owner 明确：manual review command service 只保留业务分支和 errcode 映射，manual review adapter 承担 practice raw repo 的 not-found concrete 映射
- reuse point 明确：challenge not-found 复用现有 runtime subject adapter，不重复造第二个同语义 adapter
- 不会留下“先把 gorm 挪到别的 command helper，下一轮再拆”的隐性返工
- touched debt 明确：这刀只收 `manual_review_service.go`；`submission_service.go` 仍留作下一 slice
