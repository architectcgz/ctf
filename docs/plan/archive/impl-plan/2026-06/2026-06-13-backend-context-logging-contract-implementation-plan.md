<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# 2026-06-13-backend-context-logging-contract Implementation Plan

**Goal:** 建立 request-context 日志契约，让认证、实例维护、容器运行时试点路径的结构化日志自动带上 `request_id`，并用源码 guardrail 防止试点文件回退到裸 `*zap.Logger` `Debug/Info/Warn/Error` 调用。

**Architecture:** 先把 `request_id` 从 Gin middleware 写入 `http.Request.Context()`，再提供一个不落在 `internal/infrastructure/logger` 下的共享 helper，从 `context.Context` 提取链路字段并补到 zap fields。试点只迁移少量关键路径，并用窄范围 architecture test 约束这些文件必须经 helper 打日志，不在本 slice 做全仓替换。

**Tech Stack:** Go, Gin, zap, `context.Context`, `go test`, repository-local architecture tests

---

## Task Metadata

- Task Slug: `2026-06-13-backend-context-logging-contract`
- Parent Task Group: `2026-06-13-backend-error-management-group` <!-- 独立任务写"无"；task group slice 写 parent group slug -->
- Slice Index: `3/13` <!-- 独立任务写"-"；task group slice 写 "1/5"、"2/5" -->
- Depends On: `2026-06-13-backend-sensitive-log-sanitizer` <!-- 前置依赖 task slug，多个用逗号分隔；无依赖写"无" -->
- Started At: `2026-06-13T01:41:41Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-13-backend-context-logging-contract`
- Branch: `task/2026-06-13-backend-context-logging-contract`
- Plan Type: `slice` <!-- slice | roadmap -->

## Plan Status

- Status: `archived` <!-- draft | ready-for-implementation | implemented | review-pending | review-passed | archived -->
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective:
  - 让 request 级 `request_id` 真正进入 downstream `context.Context`。
  - 新增共享 context-aware logging helper，供 application / infrastructure 关键路径复用，并覆盖试点所需的 `Debug/Info/Warn/Error`。
  - 迁移认证、实例维护、容器运行时三条试点路径，并为试点文件加回退 guardrail。
- Non-Goals:
  - 不在本 slice 做全仓 `logger.Info/Warn/Error` 替换。
  - 不把 application 层改为依赖 `internal/infrastructure/logger`。
  - 不同时实现 goroutine `SafeGo`、metrics、timeout 策略或 Redis error boundary。

## Problem Statement

- Current behavior / structure:
  - Gin `RequestID` middleware 只把 `request_id` 放进 `gin.Context` 和响应头，没有写回 `http.Request.Context()`。
  - application / infrastructure 中大量直接调用 `*zap.Logger`，即便显式传了 `ctx`，日志也看不到 request-scoped 字段。
  - 现有 `internal/module/practice/application/commands/error_logging.go` 只处理 wrapped cause，不负责 request context。
- Target behavior / structure:
  - `request_id` 进入 `http.Request.Context()` 后，可被共享 helper 统一提取。
  - 共享 helper 只负责把 context 中的 request fields 追加到 zap 日志，不承担日志配置构建职责。
  - 试点文件中的 error/warn/info 日志统一经 helper 发出，并由源码级测试禁止回退。
- Why this task is needed now:
  - Slice 1 已经建立敏感字段 guardrail，Slice 3 现在需要把“能不能安全打日志”推进到“能不能带链路字段打日志”。
  - Slice 4 `SafeGo`、Slice 9 error-rate metrics、Slice 13 close/sleep cleanup 都依赖一条可复用的 context-aware logging 路径。

## Inputs

- Source docs:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-backend-error-management-improvement-plan.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/internal/app/backend_context_architecture_test.go`
  - `code/backend/tests/architecture/test_architecture_test.go`
- Related prior work:
  - `code/backend/internal/platform/logsanitize/*`
  - `code/backend/internal/module/practice/application/commands/error_logging.go`
  - `code/backend/internal/auditlog/context.go`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 触达 middleware、共享平台能力、多个 backend module 试点和 architecture guardrail。
  - 需要独立 worktree、implementation plan、focused validation 和独立 review gate。

## Files

- Create:
  - `code/backend/internal/platform/requestctx/request_id.go`
  - `code/backend/internal/platform/requestctx/request_id_test.go`
  - `code/backend/internal/platform/logctx/logger.go`
  - `code/backend/internal/platform/logctx/logger_test.go`
- Modify:
  - `code/backend/internal/middleware/request_id.go`
  - `code/backend/internal/middleware/request_id_test.go`
  - `code/backend/internal/module/auth/application/commands/service.go`
  - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
  - `code/backend/internal/module/container_runtime/application/commands/provisioning_service.go`
  - `code/backend/tests/architecture/test_architecture_test.go`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
- Review:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/internal/module/practice/application/commands/error_logging.go`
- Test:
  - `code/backend/internal/middleware/request_id_test.go`
  - `code/backend/internal/platform/requestctx/request_id_test.go`
  - `code/backend/internal/platform/logctx/logger_test.go`
  - `code/backend/internal/module/auth/application/commands/*test.go`
  - `code/backend/internal/module/instance/application/commands/*test.go`
  - `code/backend/internal/module/container_runtime/application/commands/*test.go`
  - `code/backend/tests/architecture/test_architecture_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `internal/auditlog/context.go`：已有 context value owner 模式。
  - `internal/module/practice/application/commands/error_logging.go`：已有局部日志 helper，但只处理 cause field。
  - `internal/middleware/request_id.go`：已有 request id 生成与 header 处理。
- Reuse / extend / split / create-new decision:
  - 复用 `RequestID` middleware 作为 request id owner。
  - 新建 `internal/platform/requestctx` 作为 request-scoped field owner。
  - 新建 `internal/platform/logctx` 作为共享日志 helper；试点先覆盖 `Debug/Info/Warn/Error`，不扩展 `internal/infrastructure/logger`。
  - 保留 `wrappedErrorCauseField` 这类局部 helper，改为与 `logctx` 组合使用。
- Owner boundary:
  - `middleware/request_id.go` 负责把 request id 写入 gin/http 双上下文。
  - `platform/requestctx` 只负责 context value 存取。
  - `platform/logctx` 只负责从 context 抽字段并委托到 zap。
  - 各业务 service / infrastructure 只决定“何时记录日志”和业务字段，不拥有 request-field 拼装规则。
- Why this is the narrowest safe surface:
  - 先补 context owner 和共享 helper，再迁移三条试点路径，能解锁后续 slice，又不会把全仓日志重写绑进同一次提交。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这不是单点 bug，而是共享日志契约和 owner 边界设计，需要先确认 context owner、helper 落点和试点范围。
- grill-with-docs findings:
  - 当前 `c.Request.Context()` 已广泛向下游传递，但 `request_id` 没有进 request context，必须先补 owner。
  - application 层普遍持有 `*zap.Logger`；如果 helper 落在 `internal/infrastructure/logger` 会形成错误依赖方向。
  - 现有 architecture tests 已接受窄范围源码 guardrail，适合先锁试点文件，不宜直接扫全仓。
- Plan adjustments after challenge:
  - 从“helper 直接放 logger 包”改为“`requestctx` + `logctx` 双包”。
  - 试点收敛到 `auth/application/commands/service.go`、`instance/application/commands/maintenance_service.go`、`container_runtime/application/commands/provisioning_service.go`。
  - task group index 在本 slice 中同步把 Slice 2 标成 `completed`，把 Slice 3 标成 `in-progress`。

## Execution Slices

### Slice 1: Context owner 与 helper 基线

- Goal:
  - 让 request id 进入 request context，并建立共享 helper 的最小 API。
- Dependencies:
  - 无
- Files:
  - Create:
    - `code/backend/internal/platform/requestctx/request_id.go`
    - `code/backend/internal/platform/requestctx/request_id_test.go`
    - `code/backend/internal/platform/logctx/logger.go`
    - `code/backend/internal/platform/logctx/logger_test.go`
  - Modify:
    - `code/backend/internal/middleware/request_id.go`
    - `code/backend/internal/middleware/request_id_test.go`
  - Review:
    - `code/backend/internal/auditlog/context.go`
  - Test:
    - `code/backend/internal/platform/requestctx/request_id_test.go`
    - `code/backend/internal/platform/logctx/logger_test.go`
    - `code/backend/internal/middleware/request_id_test.go`
- Steps:
  - [x] Step 1: 先写 `requestctx` / `logctx` / middleware 的 failing tests，证明 request id 目前不会进入 request context，也不会自动进入日志字段。
  - [x] Step 2: 跑最小测试，确认红灯失败原因正确。
  - [x] Step 3: 实现 `requestctx.WithRequestID/RequestIDFromContext` 与 `logctx.Info/Warn/Error`。
  - [x] Step 4: 修改 `middleware.RequestID()`，把 `request_id` 写入 `c.Request.Context()`。
  - [x] Step 5: 重跑该组测试，确认转绿。
- Validation:
  - `go test ./internal/middleware ./internal/platform/requestctx ./internal/platform/logctx -count=1`
- Review focus:
  - helper 落点不能反向依赖 module / gin。
  - `Debug` 不能继续作为例外留在裸 logger。
  - nil context / 空 request id / nil logger 的行为必须可预期。
- Done criteria:
  - request id 进入 request context。
  - helper 在日志中附加 request id；无 request id 时不产生脏字段。

### Slice 2: 试点迁移、guardrail 与 task group 同步

- Goal:
  - 迁移认证、实例维护、容器运行时试点文件，并为这些文件增加回退 guardrail。
- Dependencies:
  - Slice 1
- Files:
  - Create:
    - 无
  - Modify:
    - `code/backend/internal/module/auth/application/commands/service.go`
    - `code/backend/internal/module/instance/application/commands/maintenance_service.go`
    - `code/backend/internal/module/container_runtime/application/commands/provisioning_service.go`
    - `code/backend/tests/architecture/test_architecture_test.go`
    - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
  - Review:
    - `code/backend/internal/module/auth/api/http/handler.go`
    - `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
  - Test:
    - 对应 package 现有单测 + 新增 observer / source guardrail tests
- Steps:
  - [x] Step 1: 先为 auth / instance / container_runtime 试点日志写 failing tests，确认旧实现不会带 `request_id`。
  - [x] Step 2: 跑 targeted tests，确认红灯。
  - [x] Step 3: 用 `logctx` 迁移试点文件中的 `Info/Warn/Error` 调用。
  - [x] Step 4: 在 architecture test 中为试点文件添加“不得直接调用裸 logger”的 guardrail。
  - [x] Step 5: 同步 task group index 当前事实。
  - [x] Step 6: 跑 focused tests 与 architecture guardrail，确认转绿。
- Validation:
  - `go test ./internal/module/auth/application/commands ./internal/module/instance/application/commands ./internal/module/container_runtime/application/commands -count=1`
  - `go test ./tests/architecture -run 'Test(NoRawSensitiveZapFields|ContextAwareLoggingContractPilots)' -count=1`
- Review focus:
  - 只改日志契约，不顺带改变业务错误分支。
  - guardrail 先锁试点文件，不扩大到全仓；`Debug` 也不能成为回退口。
- Done criteria:
  - 试点路径日志在 context 带 request id 时能自动落字段。
  - 试点文件没有裸 `logger.Debug/Info/Warn/Error` 回退。
  - task group index 与当前实际状态一致。

## Impact And Compatibility

- API / DTO:
  - 无公开 API 结构变化。
- Data / migration:
  - 无。
- State / cache / queue / event:
  - 无持久化协议变化；仅补 request-scoped 日志字段。
- Runtime / config:
  - 无新增配置项；request id 仍沿用现有 header / fallback 生成规则。
- Frontend route / state / UX:
  - 无。
- Docs / contracts:
  - task group index 状态更新；如 helper owner 明显影响架构事实，再补 backend architecture 文档。

## Plan Review / Architecture Fit

- Target owner boundary:
  - request field owner 在 `platform/requestctx`，日志拼装 owner 在 `platform/logctx`，业务 service 只保留业务字段和日志时机 owner。
- Reuse points / landing zones:
  - 复用现有 middleware request-id 生成。
  - 复用 `auditlog/context.go` 的 context value pattern。
  - 复用现有 module tests 的 context propagation style 和 `zaptest/observer`。
- Known structural debt touched:
  - `practice/application/commands/error_logging.go` 仍然是局部 helper，本 slice 不统一替换，但后续可决定是否与 `logctx` 收口。
  - `auth/api/http/handler.go` 仍有 handler 自己的日志与 audit detail 拼装，不在本 slice 清理。
- How this plan avoids behavior-only convergence:
  - 不是只在三个文件手动补 `zap.String("request_id", ...)`，而是先建立共享 owner，再迁移试点文件。
- Hidden second-redesign risk:
  - 若后续需要 trace/user/session 等更多链路字段，`requestctx` / `logctx` 需要继续扩展；本次 API 应避免只为 request_id 写死不可扩展结构。
- Decision after review:
  - 通过；按“共享 owner + 小范围试点 + 窄 guardrail”执行。

## Documentation Owner

- Current fact sources to read:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
- Fact sources to update after implementation:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
  - 如 helper owner 需要沉淀，再更新 `docs/architecture/backend/04-api-design.md` 或相关 backend 架构事实源。
- Plan-only notes that must not become architecture source:
  - “先只试点三条路径”的 rollout 说明属于实施策略，不是长期架构事实。
- Archive condition:
  - focused tests、`completion-full`、独立 backend review 与 `workflow-governance` 通过后归档到 `docs/plan/archive/impl-plan/2026-06/`。

## Validation Plan

- Per-slice commands:
  - Slice 1: `go test ./internal/middleware ./internal/platform/requestctx ./internal/platform/logctx -count=1`
  - Slice 2: `go test ./internal/module/auth/application/commands ./internal/module/instance/application/commands ./internal/module/container_runtime/application/commands -count=1`
  - Slice 2: `go test ./tests/architecture -run 'Test(NoRawSensitiveZapFields|ContextAwareLoggingContractPilots)' -count=1`
- Integration commands:
  - `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Manual checks:
  - 无额外手工操作；日志字段用 observer tests 证明。
- Commands intentionally skipped and why:
  - 全仓 `go test ./...`：本 slice 只改共享 helper 与三条试点路径，先跑最小充分范围。

## Validation Evidence

- Command: `timeout 240s go test ./internal/middleware ./internal/platform/requestctx ./internal/platform/logctx -count=1`
  - Result: `PASS`
  - Notes: 覆盖 request id context owner 与 shared helper 基线。
- Command: `timeout 240s go test ./internal/module/auth/application/commands ./internal/module/instance/application/commands ./internal/module/container_runtime/application/commands -count=1`
  - Result: `PASS`
  - Notes: 覆盖三条试点路径日志迁移。
- Command: `timeout 240s go test ./tests/architecture -run 'Test(NoRawSensitiveZapFields|ContextAwareLoggingContractPilots)' -count=1`
  - Result: `PASS`
  - Notes: 覆盖敏感字段 guardrail 与试点文件 logctx 约束。
- Command: `timeout 240s go test ./internal/module/auth/application/commands ./internal/module/instance/application/commands ./internal/module/container_runtime/application/commands -count=1`
  - Result: `PASS`（review blocker 修复后复跑）
  - Notes: 确认 request-id 归一化修复没有破坏三条试点路径。
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: `PASS`（review blocker 修复后复跑）
  - Notes: backend architecture / tests architecture gate 均通过。
- Command: `timeout 240s go test ./internal/middleware ./internal/platform/requestctx ./internal/platform/logctx -count=1`
  - Result: `PASS`（补充 `Debug` context-aware logging 后复跑）
  - Notes: 确认 `logctx.Debug` 与 request-id 注入没有破坏 middleware/requestctx/logctx 基线。
- Command: `timeout 240s go test ./internal/module/auth/application/commands ./internal/module/instance/application/commands ./internal/module/container_runtime/application/commands -count=1`
  - Result: `PASS`（补充 `Debug` context-aware logging 后复跑）
  - Notes: 覆盖实例维护试点 `Debug` 迁移以及三条试点路径回归。
- Command: `timeout 240s go test ./tests/architecture -run 'Test(NoRawSensitiveZapFields|ContextAwareLoggingContractPilots)' -count=1`
  - Result: `PASS`（补充 `Debug` context-aware logging 后复跑）
  - Notes: 确认试点 guardrail 已把裸 `Debug` 也纳入约束。
- Command: `timeout 180s bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - Result: `PASS`（补充 `Debug` context-aware logging 后复跑）
  - Notes: 确认扩面后 backend architecture/test architecture 阶段仍通过。

## Independent Review Handoff

- Review target:
  - context owner、shared helper 落点、试点文件 guardrail、是否意外引入 module 反向依赖
- Validation evidence summary:
  - focused middleware/platform tests
  - auth / instance / container_runtime focused tests
  - architecture guardrail
  - `completion-full`
 - Architecture / contract inputs:
  - `AGENTS.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-13-backend-error-management-group/INDEX.md`
 - Known risks / review focus:
  - request_id 是否真正进入 downstream `context.Context`
  - helper 是否误绑到 `internal/infrastructure/logger`
  - guardrail 是否只锁试点，不误伤非试点文件
  - 新增 `Debug` 覆盖后，实例维护试点是否仍保持 request-scoped logging 一致性
 - Project-local checks to consider:
  - `go test ./tests/architecture -run 'Test(NoRawSensitiveZapFields|ContextAwareLoggingContractPilots)' -count=1`
 - Independent review result:
   - `2026-06-13-backend-review-context-logging-contract-round-1.md`
   - `2026-06-13-backend-review-context-logging-contract-round-2.md`
   - Gate verdict: `pass with minor issues`
   - Round 1 follow-up 已转入当前 patch；Round 2 仅剩 `logctx.Debug` 的 nil fallback 单测缺口，未阻塞合并

## Rollback / Recovery

- Safe revert boundary:
  - 单次 task branch / merge commit 可整体回退。
- Data / config / runtime recovery notes:
  - 无数据迁移；回退后仅恢复到旧日志字段行为。
- Irreversible operations:
  - 无。

## Residual Risks

- Risk:
  - 仍有大量非试点文件直接调用裸 `*zap.Logger`，本 slice 不会一次清零。
- Why acceptable:
  - task group 已明确这是分片推进；本 slice 先建立共享 owner 和 guardrail 试点，给后续 `SafeGo` / metrics / cleanup 提供落点。
- Follow-up owner, if any:
  - 后续 slice 4、9、13，以及必要时新增日志迁移 slice。
