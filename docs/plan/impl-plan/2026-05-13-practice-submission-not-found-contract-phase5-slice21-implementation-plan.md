# Practice Submission Not-Found Contract Phase 5 Slice 21 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `practice/application/commands/submission_service.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时保持 flag 提交、重复解题、动态 flag 和 solve grace 行为不变。

**Architecture:** `practice` 继续复用 `runtime_subject_repository` 处理 challenge not-found，新加一个窄 `solved_submission_repository` 处理正确提交 not-found；实例 lookup 不再由 application 识别 GORM sentinel，而是直接以 `instance == nil` 表达“没有可用实例”。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `practice/application/commands/submission_service.go -> gorm.io/gorm`
- 保持 challenge 不存在时仍返回 `errcode.ErrChallengeNotFound`
- 保持已通过提交不存在时仍按“未解出”继续后续提交流程
- 保持动态 flag 和 solve grace 路径在实例不存在时仍按当前语义静默返回 false / nil

## Non-goals

- 不改 `practice/infrastructure.Repository` 的全局 `FindCorrectSubmission` 返回约定
- 不处理 `practice` 之外模块的 GORM concrete
- 不重写 `SubmitFlag` 的计分、事件、限流或 audit 逻辑

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/application/commands/service_audit_test.go`
- `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- `code/backend/internal/module/practice/infrastructure/manual_review_repository.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`

## Ownership Boundary

- `practice/application/commands/submission_service.go`
  - 负责：flag 提交业务编排、challenge 发布状态判定、已解题分支、errcode 映射
  - 不负责：知道底层 not-found 是否来自 `gorm`
- `practice/infrastructure/solved_submission_repository.go`
  - 负责：把 raw practice repository 的正确提交 not-found 映射成 `ports.ErrPracticeSolvedSubmissionNotFound`
  - 不负责：决定上层返回哪个 `errcode`
- `practice/infrastructure/runtime_subject_repository.go`
  - 负责：继续把 challenge not-found 映射成 `ErrPracticeChallengeNotFound`
  - 不负责：承担提交记录或实例 lookup 语义
- `practice/runtime/module.go` / `internal/app/practice_flow_integration_test.go`
  - 负责：注入 solved-submission adapter 与 runtime subject adapter
  - 不负责：把 raw GORM concrete 重新带回 command service

## Change Surface

- Add: `.harness/reuse-decisions/practice-submission-not-found-contract-phase5-slice21.md`
- Add: `docs/plan/impl-plan/2026-05-13-practice-submission-not-found-contract-phase5-slice21-implementation-plan.md`
- Add: `code/backend/internal/module/practice/infrastructure/solved_submission_repository.go`
- Add: `code/backend/internal/module/practice/infrastructure/solved_submission_repository_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/submission_service.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_audit_test.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task 1: 先写失败测试

**Files:**
- Modify: `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- Add: `code/backend/internal/module/practice/infrastructure/solved_submission_repository_test.go`

- [ ] 为 `SubmitFlag` 补 challenge not-found / solved-submission not-found 的 sentinel 分支测试，先证明当前实现还依赖 GORM sentinel
- [ ] 为 solved-submission adapter 补 GORM not-found 映射测试，先证明 adapter 尚不存在
- [ ] 跑最小测试，确认红灯来自目标行为缺失

验证：

- `cd code/backend && go test ./internal/module/practice/application/commands -run 'SubmitFlagTreatsPracticeChallengeNotFoundAsChallengeNotFound|SubmitFlagTreatsPracticeSolvedSubmissionNotFoundAsUnsolved' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/practice/infrastructure -run 'SolvedSubmissionRepository' -count=1 -timeout 5m`

Review focus：

- service 测试是否真正在约束 `practice` sentinel，而不是继续借 GORM sentinel 过关
- adapter 测试是否只覆盖错误映射，不夹带业务判断

## Task 2: 实现 adapter 与 wiring

**Files:**
- Add: `code/backend/internal/module/practice/infrastructure/solved_submission_repository.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/submission_service.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/service_audit_test.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`

- [ ] 新增 solved-submission adapter，并把 `FindCorrectSubmission` 的 not-found 映射成已有 sentinel
- [ ] 给 command service 增加 solved-submission dependency 的显式 setter
- [ ] 让 `SubmitFlag` 改成只看 practice sentinel，并复用 runtime subject adapter 处理 challenge not-found
- [ ] 把实例不存在分支改成只看 `instance == nil`，不再识别 GORM sentinel
- [ ] 在 runtime wiring、app wiring 和测试 helper 中注入 adapter

验证：

- `cd code/backend && go test ./internal/module/practice/application/commands -run 'SubmitFlag' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/practice/infrastructure -run 'SolvedSubmissionRepository|RuntimeSubjectRepository' -count=1 -timeout 5m`

Review focus：

- application surface 是否已经完全去掉 `gorm` concrete
- solved-submission adapter 是否保持窄，只承接正确提交 not-found 映射

## Task 3: 删除 allowlist 并同步文档

**Files:**
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

- [ ] 删除本 slice 实际收口的 allowlist
- [ ] 更新 phase5 当前状态与 practice 剩余 debt 描述
- [ ] 跑架构 / 文档 / workflow 完整性检查

验证：

- `cd code/backend && go test ./internal/module -run 'Test(ApplicationConcreteDependencyAllowlistIsCurrent|ModuleDependencyAllowlistIsCurrent)' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `timeout 600 bash scripts/check-workflow-complete.sh`
- `git diff --check`

Review focus：

- 只删除已经真实收口的 allowlist
- 文档是否准确描述“runtime subject adapter + solved-submission adapter”的 owner 分工

## Risks

- 如果 solved-submission adapter 注入不全，`SubmitFlag` 会退化成 internal error
- 如果测试仍直接构造未注入 adapter 的 service，`SubmitFlag` 相关用例会在 nil surface 上误失败
- 如果误改 `instanceRepo` not-found 语义，可能影响 solve grace 或动态 flag 现有流程

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/commands -run 'SubmitFlag' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/practice/infrastructure -run 'SolvedSubmissionRepository|RuntimeSubjectRepository' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/practice/runtime -run '^$' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/app -run '^$' -count=1 -timeout 5m`
5. `cd code/backend && go test ./internal/module -run 'Test(ApplicationConcreteDependencyAllowlistIsCurrent|ModuleDependencyAllowlistIsCurrent)' -count=1 -timeout 5m`
6. `python3 scripts/check-docs-consistency.py`
7. `bash scripts/check-consistency.sh`
8. `timeout 600 bash scripts/check-workflow-complete.sh`
9. `git diff --check`

## Architecture-Fit Evaluation

- owner 明确：submission application 只保留 challenge / solved-state 分支和 errcode 映射，solved-submission adapter 承担 raw repo 的 not-found concrete 映射
- reuse point 明确：challenge not-found 继续复用现有 runtime subject adapter，不重复造 challenge lookup adapter
- 实例 lookup 不再人为抽象一层未被生产实现需要的 not-found contract，避免过度设计
- touched debt 明确：这刀只收 `submission_service.go` 的 GORM concrete；practice 其他已清掉的 surface 不回退
