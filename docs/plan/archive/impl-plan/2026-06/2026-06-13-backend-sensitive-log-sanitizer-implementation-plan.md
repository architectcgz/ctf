<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# Backend sensitive log sanitizer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking; flip each checkbox immediately after the expected result is reached.

**Goal:** 建立后端共享日志脱敏能力，阻止 password、token、secret 和高风险 cache key 明文进入结构化日志。

**Architecture:** 新增 `internal/platform/logsanitize` 作为跨模块、无业务 owner 的小型值清洗包；application、infrastructure 和 shared helper 可以显式调用它生成安全的 zap 字段值。源码级 architecture test 只禁止高风险明文字段名，允许业务 `node_key`、`network_key`、`phase_key` 这类非凭据标识继续记录。

**Tech Stack:** Go, zap, AST/source architecture tests, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-13-backend-sensitive-log-sanitizer`
- Parent Task Group: `2026-06-13-backend-error-management-group`
- Slice Index: `1/13`
- Depends On: `无` <!-- 前置依赖 task slug，多个用逗号分隔；无依赖写"无" -->
- Started At: `2026-06-12T23:39:11Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-sensitive-log-sanitizer`
- Branch: `task/2026-06-13-backend-sensitive-log-sanitizer`
- Plan Type: `slice` <!-- slice | roadmap -->

## Plan Status

- Status: `review-passed`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective:
  - 提供共享脱敏函数，覆盖 token/password/secret/cache key 四类日志输入。
  - 增加源码级 guardrail，禁止新增 `zap.String("password", ...)`、`zap.String("*token*", raw)`、`zap.String("*secret*", raw)` 这类高风险字段。
  - 迁移当前已发现的 session cache key 日志，避免完整 Redis session key 进入日志。
- Non-Goals:
  - 不在本 slice 建立 context-aware logging API；该能力由 `backend-context-logging-contract` slice 承接。
  - 不全量迁移所有 `lock_key` 日志；锁 key 是否敏感取决于下一步日志契约和运维排查需求。
  - 不修改 API 错误响应，也不改错误码。
  - 不改已有 AWD checker 的局部 secret 清洗逻辑；本 slice 只提供共享能力，后续可逐步复用。

## Problem Statement

- Current behavior / structure:
  - 后端已有少量局部 `sanitizeAWDCheckerText` / `sanitizeAWDCheckErrorWithSecrets`，但没有跨模块日志字段值脱敏工具。
  - `ops/infrastructure/dashboard_state_store.go` 会在解析在线 session 记录失败时记录完整 Redis key。
  - 架构测试尚未阻止新增 password/token/secret 明文字段。
- Target behavior / structure:
  - 高风险日志值通过 `internal/platform/logsanitize` 显式脱敏。
  - session/cache key 类字段只保留可排查的短前缀，不保留完整 token。
  - architecture test 能在代码 review 前拦住明显的敏感字段日志。
- Why this task is needed now:
  - 原错误管理计划把敏感信息泄漏列为安全关键项；该 slice 是后续 context logging 和错误日志覆盖率提升前的安全前置。

## Inputs

- Source docs:
  - `docs/plan/impl-plan/2026-06-12-backend-error-management-improvement-plan.md`
  - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/tests/README.md`
- Related prior work:
  - `code/backend/internal/module/contest/application/jobs/awd_probe_support.go`
  - `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 触达安全日志策略、共享平台包和 architecture guardrail，会影响后续所有错误日志新增方式。
  - 需要 TDD 证明脱敏行为和源码扫描护栏。

## Files

- Create:
  - `code/backend/internal/platform/logsanitize/sanitize.go`
  - `code/backend/internal/platform/logsanitize/sanitize_test.go`
- Modify:
  - `code/backend/internal/module/ops/infrastructure/dashboard_state_store.go`
  - `code/backend/tests/architecture/test_architecture_test.go`
  - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
  - `docs/plan/impl-plan/2026-06-13-backend-sensitive-log-sanitizer-implementation-plan.md`
- Review:
  - `code/backend/internal/module/contest/application/jobs/awd_probe_support.go`
  - `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
  - `code/backend/internal/shared/lockkeepalive/lockkeepalive.go`
- Test:
  - `code/backend/internal/platform/logsanitize/sanitize_test.go`
  - `code/backend/tests/architecture/test_architecture_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - AWD checker 已有局部 secret 清洗，但只服务 checker output，不适合作为全局日志包。
  - `internal/platform/*` 已承载跨模块、低业务语义的 events、randomstring、storage、clustersecret 能力。
  - `internal/infrastructure/logger` 当前只负责 zap logger 构造，application 不应为了清洗字段反向依赖该 builder 包。
- Reuse / extend / split / create-new decision:
  - 新建 `internal/platform/logsanitize`，函数保持小而显式。
  - 保留现有 AWD 局部清洗函数，不在本 slice 迁移，避免扩大影响面。
  - architecture test 用源码扫描约束高风险字段名，不用 runtime hook 包装 zap。
- Owner boundary:
  - `internal/platform/logsanitize`：日志字段值清洗 owner。
  - 调用方：决定哪个业务字段需要清洗，并传入清洗后的值。
  - `tests/architecture`：禁止明显敏感字段明文进入 zap field 的源码护栏。
- Why this is the narrowest safe surface:
  - 不重写 logger API，不改变日志输出格式，不影响已有业务逻辑。
  - 先迁移一个已发现的高风险 session key 日志，证明工具可用。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 该 slice 添加共享能力和安全 guardrail，需要先确认 owner 落点和误伤范围。
- grill-with-docs findings:
  - 架构事实要求 `internal/module/*` 保持 owner 边界；共享能力应放在 `internal/platform`，不放进单个 module。
  - 原计划示例中的 `pkg/logging/sanitizer.go` 不符合当前仓库结构；仓库没有 `code/backend/pkg`。
  - 直接禁止所有 `*key` 字段会误伤 `node_key`、`network_key`、`phase_key` 等非凭据日志字段。
- Plan adjustments after challenge:
  - 采用 `internal/platform/logsanitize`，不使用 `pkg/logging`。
  - Guardrail 只禁止 password/token/secret 以及裸 `zap.Any("req"/"request")`，不禁止所有 `key` 字段。
  - 对 session/cache key 通过调用方显式使用 `SanitizeKey` 迁移。

## Execution Slices

### Slice 1: Shared sanitizer behavior

- Goal: 先用 TDD 建立共享清洗函数行为。
- Dependencies: 无
- Files:
  - Create:
    - `code/backend/internal/platform/logsanitize/sanitize_test.go`
    - `code/backend/internal/platform/logsanitize/sanitize.go`
  - Modify:
  - Review:
  - Test:
    - `code/backend/internal/platform/logsanitize/sanitize_test.go`
- Steps:
  - [x] Step 1: 写 `SanitizePassword` / `SanitizeToken` / `SanitizeSecret` / `SanitizeKey` 失败测试。
  - [x] Step 2: 运行 `cd code/backend && go test ./internal/platform/logsanitize -run TestSanitize -count=1`，确认因 package/function 缺失失败。
  - [x] Step 3: 实现最小清洗函数。
  - [x] Step 4: 重跑 focused test，确认通过。
- Validation: `cd code/backend && go test ./internal/platform/logsanitize -run TestSanitize -count=1`
- Review focus: 输出不能包含原始 token/secret/password；key 只保留 namespace 和短前缀。
- Done criteria: 清洗函数测试通过。

### Slice 2: Source guardrail

- Goal: 阻止新增明显敏感 zap 字段。
- Dependencies: Slice 1
- Files:
  - Create:
  - Modify:
    - `code/backend/tests/architecture/test_architecture_test.go`
  - Review:
  - Test:
    - `code/backend/tests/architecture/test_architecture_test.go`
- Steps:
  - [x] Step 5: 写 `TestNoRawSensitiveZapFields`，先断言当前 `dashboard_state_store.go` 的 `zap.String("key", keys[index])` 为 violation。
  - [x] Step 6: 运行 `cd code/backend && go test ./tests/architecture -run TestNoRawSensitiveZapFields -count=1`，确认失败。
  - [x] Step 7: 迁移 `dashboard_state_store.go` 使用 `logsanitize.SanitizeKey`。
  - [x] Step 8: 重跑 architecture focused test，确认通过。
- Validation: `cd code/backend && go test ./tests/architecture -run TestNoRawSensitiveZapFields -count=1`
- Review focus: guardrail 不误伤 `node_key` / `network_key` / `phase_key` 等非凭据字段。
- Done criteria: Guardrail 能拦截高风险字段，当前代码无 violation。

### Slice 3: Slice validation and docs state

- Goal: 跑完整当前 slice 验证并更新 plan checklist。
- Dependencies: Slice 1-2
- Files:
  - Create:
  - Modify:
    - `docs/plan/impl-plan/2026-06-13-backend-sensitive-log-sanitizer-implementation-plan.md`
  - Review:
  - Test:
- Steps:
  - [x] Step 9: 运行 `cd code/backend && go test ./internal/platform/logsanitize ./tests/architecture -run 'Test(Sanitize|NoRawSensitiveZapFields)' -count=1`。
  - [x] Step 10: 运行 `git diff --check -- code/backend/internal/platform/logsanitize/sanitize.go code/backend/internal/platform/logsanitize/sanitize_test.go code/backend/internal/module/ops/infrastructure/dashboard_state_store.go code/backend/tests/architecture/test_architecture_test.go docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md docs/plan/impl-plan/2026-06-13-backend-sensitive-log-sanitizer-implementation-plan.md`。
  - [x] Step 11: 运行 `bash scripts/check-startup-gate.sh`。
  - [x] Step 12: 准备 independent review handoff。
- Validation: 见上。
- Review focus: 安全收益、误伤范围、后续 slice 依赖是否清楚。
- Done criteria: focused tests 和 diff check 通过，plan checklist 与状态同步。

## Impact And Compatibility

- API / DTO:
  - 无变更。
- Data / migration:
  - 无变更。
- State / cache / queue / event:
  - 不改变 Redis key 或 cache 行为，只改变日志输出值。
- Runtime / config:
  - 无变更。
- Frontend route / state / UX:
  - 无变更。
- Docs / contracts:
  - 新增 task group index，当前 slice plan 更新；不更新架构事实源。

## Plan Review / Architecture Fit

- Target owner boundary:
  - `internal/platform/logsanitize` 是共享日志值清洗 owner；业务模块仍显式决定何时调用。
- Reuse points / landing zones:
  - 后续 `context-logging`、`redis-error-boundary`、`resource-close-and-sleep-cleanup` slice 都复用该包。
- Known structural debt touched:
  - 不触达已记录的大 service / command boundary debt。
- How this plan avoids behavior-only convergence:
  - 不只迁移一个日志点，同时增加 guardrail 防止同类明文字段回流。
- Hidden second-redesign risk:
  - 低。后续若引入 context logger，仍可内部复用 `logsanitize`，不会废弃本 slice。
- Decision after review:
  - `ready-for-implementation`。

## Documentation Owner

- Current fact sources to read:
  - `docs/文档规范.md`
  - `docs/plan/impl-plan/README.md`
- Fact sources to update after implementation:
  - 本 slice 只更新 implementation plan 和 task group index。
  - 真正的日志规范文档放到后续 `backend-error-runbook-docs`，在更多能力落地后统一吸收。
- Plan-only notes that must not become architecture source:
  - 原计划中的 `pkg/logging/sanitizer.go` 示例路径不是当前仓库事实。
- Archive condition:
  - 当前 slice 合并并通过 independent review 后归档当前 slice plan；task group index 保持活动直到全部 slice 完成。

## Validation Plan

- Per-slice commands:
  - `cd code/backend && go test ./internal/platform/logsanitize -run TestSanitize -count=1`
  - `cd code/backend && go test ./tests/architecture -run TestNoRawSensitiveZapFields -count=1`
- Integration commands:
  - `cd code/backend && go test ./internal/platform/logsanitize ./tests/architecture -run 'Test(Sanitize|NoRawSensitiveZapFields)' -count=1`
  - `bash scripts/check-startup-gate.sh`
- Manual checks:
  - Review diff 中不出现 `zap.String("password", ...)`、`zap.String("*token*", raw)`、`zap.String("*secret*", raw)`。
- Commands intentionally skipped and why:
  - 不跑全量后端测试；当前 slice 只新增小包和源码 guardrail，focused tests 足够覆盖行为。

## Validation Evidence

- Command: `cd code/backend && go test ./internal/platform/logsanitize -run TestSanitize -count=1`
  - Result: PASS
  - Notes: Red step first failed with undefined `Sanitize*` / `RedactedValue`; green step passed after adding `internal/platform/logsanitize`.
- Command: `cd code/backend && go test ./tests/architecture -run TestNoRawSensitiveZapFields -count=1`
  - Result: PASS
  - Notes: Red step first failed on `dashboard_state_store.go` raw `zap.String("key", keys[index])`; green step passed after `logsanitize.SanitizeKey`.
- Command: `cd code/backend && go test ./internal/platform/logsanitize ./tests/architecture -run 'Test(Sanitize|NoRawSensitiveZapFields)' -count=1`
  - Result: PASS
  - Notes: Combined focused validation.
- Command: `cd code/backend && go test ./internal/module/ops/infrastructure -count=1`
  - Result: PASS
  - Notes: Confirms the migrated ops infrastructure package still compiles and tests pass.
- Command: `git diff --check -- code/backend/internal/platform/logsanitize/sanitize.go code/backend/internal/platform/logsanitize/sanitize_test.go code/backend/internal/module/ops/infrastructure/dashboard_state_store.go code/backend/tests/architecture/test_architecture_test.go docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md docs/plan/impl-plan/2026-06-13-backend-sensitive-log-sanitizer-implementation-plan.md`
  - Result: PASS
  - Notes: No whitespace errors.
- Command: `bash scripts/check-startup-gate.sh`
  - Result: PASS
  - Notes: No startup-gated changes in diff.
- Command: `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`
  - Result: PASS
  - Notes: Frontend test guard, startup gate, quick architecture and backend test architecture guard passed.
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS
  - Notes: Code change contracts, backend architecture, frontend architecture no-op check and test architecture guard passed.
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`
  - Result: PASS
  - Notes: Workflow governance checks passed; open todos were surfaced but not part of this slice.
- Command: `timeout 300s codex exec ...`
  - Result: TIMED OUT
  - Notes: Independent reviewer did not return a verdict before timeout. Static review output identified two hardening points: preserve multi-segment Redis key namespace and widen zap sensitive-field guardrail coverage.
- Command: `cd code/backend && go test ./internal/platform/logsanitize ./tests/architecture -run 'Test(Sanitize|NoRawSensitiveZapFields)' -count=1`
  - Result: PASS
  - Notes: Re-run after hardening `SanitizeKey` and widening zap guardrail coverage.
- Command: `cd code/backend && go test ./internal/module/ops/infrastructure -count=1`
  - Result: PASS
  - Notes: Re-run after hardening changes.
- Command: `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`
  - Result: PASS
  - Notes: Re-run after hardening changes.
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: PASS
  - Notes: Re-run after hardening changes.
- Command: `timeout 420s codex exec --sandbox read-only ... -o /tmp/ctf-sensitive-log-sanitizer-review.md`
  - Result: PASS
  - Notes: Independent `code-reviewer` gate review returned `pass` with no material findings. Archived at `docs/reviews/security/2026-06-13-backend-review-sensitive-log-sanitizer.md`.

## Independent Review Handoff

- Review target:
  - Commit `e643861e353957eb21ce73b805ed0241eee68312` / diff `HEAD~1..HEAD`。
- Validation evidence summary:
  - `go test ./internal/platform/logsanitize ./tests/architecture -run 'Test(Sanitize|NoRawSensitiveZapFields)' -count=1`: PASS
  - `go test ./internal/module/ops/infrastructure -count=1`: PASS
- Architecture / contract inputs:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/tests/README.md`
  - `docs/plan/impl-plan/2026-06-13-backend-error-management-group/INDEX.md`
- Known risks / review focus:
  - Guardrail 误伤普通业务 key。
  - `SanitizeKey` 输出仍需足够定位 namespace，但不能保留完整 token。
- Project-local checks to consider:
  - `bash scripts/check-startup-gate.sh`
  - `cd code/backend && go test ./internal/platform/logsanitize ./tests/architecture -run 'Test(Sanitize|NoRawSensitiveZapFields)' -count=1`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`

## Review Gate Status

- Independent review gate: `pass`
- Review archive: `docs/reviews/security/2026-06-13-backend-review-sensitive-log-sanitizer.md`
- Notes:
  - First independent `codex exec` review attempt timed out before verdict.
  - The partial review output exposed two hardening points; both were fixed and re-validated.
  - Second independent `code-reviewer` gate review returned `pass` with no material findings.
- Required next action: 进入后续 slice；当前 slice 可在合并后按 `code-workflow` 归档。

## Rollback / Recovery

- Safe revert boundary:
  - 可整体 revert 当前 slice；不会影响数据库、配置或 API contract。
- Data / config / runtime recovery notes:
  - 无。
- Irreversible operations:
  - 无。

## Residual Risks

- Risk:
  - 只迁移一个已发现 session key 日志点，其他锁 key 是否需要脱敏留给后续 logging contract slice。
- Why acceptable:
  - 当前 slice 目标是安全基础能力和 guardrail；大规模迁移会扩大 review 面。
- Follow-up owner, if any:
  - `2026-06-13-backend-context-logging-contract`
