# Practice Contest Scope Not-Found Contract Phase 5 Slice 18 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除 `practice/application/commands/contest_instance_scope.go` 和 `contest_awd_operations.go` 对 `gorm.ErrRecordNotFound` 的直接依赖，同时不改 raw repository 的全局错误语义。

**Architecture:** `practice` 新增两个局部 adapter port。contest scope adapter 负责把 practice raw repository 的 contest/challenge/service/team/registration not-found 映射成模块内 sentinel；runtime subject adapter 负责把 challenge contract 的 challenge/topology not-found 映射成模块内 sentinel。application 只依赖这些模块内 sentinel，runtime/composition 负责装配 adapter。

**Tech Stack:** Go, GORM, modular monolith ports/infrastructure, repository adapter tests

---

## Objective

- 删除 `practice/application/commands/contest_instance_scope.go -> gorm.io/gorm`
- 删除 `practice/application/commands/contest_awd_operations.go -> gorm.io/gorm`
- 保持 contest 实例开题、AWD 编排查询、practice runtime subject 加载的行为与错误码不变

## Non-goals

- 不处理 `practice/application/commands/manual_review_service.go -> gorm.io/gorm`
- 不处理 `practice/application/commands/submission_service.go -> gorm.io/gorm`
- 不处理 `practice/application/queries/score_service.go -> gorm.io/gorm`
- 不把 `challenge/infrastructure.Repository` 或 `practice/infrastructure.Repository` 的全局 not-found 语义整体改掉

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/app/composition/practice_module.go`

## Ownership Boundary

- `practice/application/commands/contest_instance_scope.go`
  - 负责：contest scope 业务编排、errcode 映射、可见性判断、contest/team/registration 状态约束
  - 不负责：知道 `gorm.ErrRecordNotFound`
- `practice/application/commands/contest_awd_operations.go`
  - 负责：AWD orchestration 业务编排和响应整形
  - 不负责：知道底层 contest 查不到时是 GORM not-found
- `practice/infrastructure/contest_scope_repository.go`
  - 负责：把 practice raw repository 的 not-found 映射成 `practice/ports` sentinel
  - 不负责：决定上层该返回哪个 errcode
- `practice/infrastructure/runtime_subject_repository.go`
  - 负责：把 challenge contract 的 challenge/topology not-found 映射成 `practice/ports` sentinel
  - 不负责：决定实例启动、续期或 provisioning 的业务分支
- `practice/runtime/module.go` / `app/composition/practice_module.go`
  - 负责：装配这两个 adapter
  - 不负责：把 GORM concrete 重新带回 practice application surface

## Change Surface

- Add: `.harness/reuse-decisions/practice-contest-scope-not-found-contract-phase5-slice18.md`
- Add: `docs/plan/impl-plan/2026-05-13-practice-contest-scope-not-found-contract-phase5-slice18-implementation-plan.md`
- Add: `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- Add: `code/backend/internal/module/practice/infrastructure/contest_scope_repository_test.go`
- Add: `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- Add: `code/backend/internal/module/practice/infrastructure/runtime_subject_repository_test.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/app/composition/practice_module.go`
- Modify: 受影响的 practice command tests / helpers
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task 1: 定义局部 port 与 adapter

**Files:**
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Add: `code/backend/internal/module/practice/infrastructure/contest_scope_repository.go`
- Add: `code/backend/internal/module/practice/infrastructure/contest_scope_repository_test.go`
- Add: `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
- Add: `code/backend/internal/module/practice/infrastructure/runtime_subject_repository_test.go`

- [ ] 定义 practice 模块内 not-found sentinel，以及 contest scope / runtime subject 的窄 port
- [ ] 为 raw practice repository 写 contest scope adapter，把 GORM not-found 映射成 practice sentinel
- [ ] 为 challenge contract 写 runtime subject adapter，把 challenge / topology not-found 映射成 practice sentinel
- [ ] 跑 adapter 相关测试

验证：

- `cd code/backend && go test ./internal/module/practice/infrastructure -run 'ContestScopeRepository|RuntimeSubjectRepository' -count=1 -timeout 5m`

Review focus：

- sentinel 是否一一对应上层真正需要分支的 not-found 语义
- adapter 是否只做错误映射，没有偷偷接管业务分支

## Task 2: 应用服务切到局部 sentinel

**Files:**
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/app/composition/practice_module.go`
- Modify: 受影响的 practice command tests / helpers

- [ ] 给 practice command service 增加两个局部 adapter setter
- [ ] 把 contest scope / orchestration 逻辑改成只看 practice sentinel，不再 import `gorm`
- [ ] 在 runtime wiring 与测试 helper 中注入 adapter，保证已有行为路径继续可用
- [ ] 跑受影响的 practice command / runtime 测试

验证：

- `cd code/backend && go test ./internal/module/practice/application/commands -run 'StartChallenge|StartContest|ProvisionInstance|RunProvisioningLoop|GetContestAWDInstanceOrchestration|LoadRuntimeSubject' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/practice/runtime -run '^$' -count=1 -timeout 5m`

Review focus：

- application 是否已经完全去掉 `gorm` concrete
- 旧测试 helper 是否都拿到 adapter，避免运行时因 repo 未配置而退化

## Task 3: 删除 allowlist 并同步文档

**Files:**
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

- [ ] 删除本 slice 实际收口的 allowlist
- [ ] 更新 phase5 当前状态与 practice 的结构事实
- [ ] 跑架构 / 文档 / workflow 完整性检查

验证：

- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `timeout 600 bash scripts/check-workflow-complete.sh`
- `git diff --check`

Review focus：

- 只删除已经真实收口的 allowlist
- 文档是否准确描述“局部 adapter 收口”，而不是误写成 raw repository 全局语义改变

## Risks

- 如果 contest scope adapter 映射不全，application 可能把真实 not-found 误包成 internal error
- 如果 runtime subject adapter 没有正确注入，实例启动 / provisioning 路径可能因为 lookup repo 未配置而失败
- 如果测试 helper 仍直接传 raw repo，局部 sentinel 分支可能在测试里失效

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/infrastructure -run 'ContestScopeRepository|RuntimeSubjectRepository' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/practice/application/commands -run 'StartChallenge|StartContest|ProvisionInstance|RunProvisioningLoop|GetContestAWDInstanceOrchestration|LoadRuntimeSubject' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/practice/runtime -run '^$' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`
8. `git diff --check`

## Architecture-Fit Evaluation

- owner 明确：application 只保留业务分支，adapter 承担 not-found concrete 映射
- reuse point 明确：复用 phase5 里已经验证过的“sentinel 在 ports、映射在 infrastructure”的模式
- 不会留下“先把 gorm 藏到别的 application 文件里，下一轮再拆”的隐性返工
- touched debt 明确：这刀只收 `contest_instance_scope.go` / `contest_awd_operations.go`；`manual_review`、`submission`、`score_service` 仍留待后续独立 slice
