<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# 后端错误管理改进 Task Group

**Task Group Slug:** `2026-06-13-backend-error-management-group`

**Goal:** 将 `2026-06-12-backend-error-management-improvement-plan.md` 拆成可独立启动、实现、验证和 review 的 code-workflow slice，逐步收口错误边界、日志安全、异步错误传播、外部依赖错误分类和可观测性。

**Status:** `in-progress`（Slice 1 已合入；Slice 2 已通过 review / governance，等待合并；其余 slice 未开始）

**Created:** `2026-06-13T00:00:00Z`

---

## Overview

- Background: 原计划覆盖错误边界、日志、goroutine、熔断、超时、监控、敏感信息和大规模错误迁移，不能作为一个可审阅 task 直接实现。
- Motivation: 每个 slice 必须有独立 `task-slug`、implementation plan、startup gate、worktree、验证证据和独立 review gate。
- Scope: 后端 Go 代码、后端测试、架构/运维文档、项目 workflow 检查。
- Non-Goals: 不在单个 slice 中完成全仓库 `ErrInternal` 迁移、全量 `Close()` 检查或真实生产环境熔断演练。

## Inputs

- Parent plan: `docs/plan/impl-plan/2026-06-12-backend-error-management-improvement-plan.md`
- Backend architecture: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Backend tests guide: `code/backend/tests/README.md`
- Project workflow: `AGENTS.md`

## Slices

### Slice 1: 敏感日志脱敏基础能力

- Task Slug: `2026-06-13-backend-sensitive-log-sanitizer`
- Status: `completed`
- Plan: [implementation-plan](../../archive/impl-plan/2026-06/2026-06-13-backend-sensitive-log-sanitizer-implementation-plan.md)
- Review: [security review](../../../reviews/security/2026-06-13-backend-review-sensitive-log-sanitizer.md)
- Depends On: 无
- Goal: 新增共享日志脱敏能力，先拦住 password/token/secret/key 这类高风险日志字段。
- Validation: `go test ./internal/platform/logsanitize ./tests/architecture -run 'Test(Sanitize|NoRawSensitiveZapFields)' -count=1`
- Review Focus: 脱敏工具落点不能制造新的 module 反向依赖，architecture guardrail 不能误伤普通业务 key。
- Notes: 已合入 `main`；plan、security review 和 governance evidence 已归档。

### Slice 2: 错误边界基线与 architecture guardrail

- Task Slug: `2026-06-13-backend-error-contract-baseline`
- Status: `ready-to-merge`
- Plan: [implementation-plan](../../archive/impl-plan/2026-06/2026-06-13-backend-error-contract-baseline-implementation-plan.md)
- Review: [backend review](../../../reviews/backend/2026-06-13-backend-review-error-contract-baseline.md)
- Depends On: 无
- Goal: 建立 application / ports / infrastructure 错误边界基线，先加 guardrail 和少量试点，不做全量业务迁移。
- Validation: `go test ./internal/apperror ./tests/architecture ./internal/app -run 'Error|Architecture' -count=1`
- Review Focus: handler 只消费 public app error；application 不直接分支 GORM/Redis/Docker sentinel。
- Notes: challenge contract repository not-found 映射、transport sentinel guardrail 和 contest adapter 兼容已完成；`completion-full`、独立 backend review、`workflow-governance` 和 plan 归档完成，当前 task 分支等待合并。

### Slice 3: Context-aware 错误日志契约

- Task Slug: `2026-06-13-backend-context-logging-contract`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-context-logging-contract` 生成
- Depends On: `2026-06-13-backend-sensitive-log-sanitizer`
- Goal: 定义 context-aware logging helper 和新增错误日志约束，先迁移容器/实例/认证等少量关键路径。
- Validation: logger helper tests、受影响模块 tests、architecture guardrail。
- Review Focus: 保留 request context，不把 application 层绑定到 `internal/infrastructure/logger` builder。

### Slice 4: Goroutine SafeGo 基础能力

- Task Slug: `2026-06-13-backend-async-safego`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-async-safego` 生成
- Depends On: `2026-06-13-backend-context-logging-contract`
- Goal: 新增 SafeGo 包装器，捕获 panic/error 并记录带上下文的结构化日志，迁移最高风险后台任务。
- Validation: SafeGo 单测、受影响 jobs tests、裸 `go func` guardrail。
- Review Focus: 区分 request ctx 与 lifecycle ctx，不让后台任务保存短生命周期 request ctx。

### Slice 5: Redis 错误语义收口

- Task Slug: `2026-06-13-backend-redis-error-boundary`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-redis-error-boundary` 生成
- Depends On: `2026-06-13-backend-error-contract-baseline`
- Goal: 把 Redis miss/timeout/unavailable 收口为模块内或 public app error 语义，先覆盖缓存与锁关键路径。
- Validation: Redis wrapper/store tests、相关模块 tests。
- Review Focus: `redis.Nil` 不泄漏到 application/handler；缓存未命中不误报为系统错误。

### Slice 6: Container runtime / Docker 错误边界试点

- Task Slug: `2026-06-13-backend-container-runtime-error-boundary`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-container-runtime-error-boundary` 生成
- Depends On: `2026-06-13-backend-error-contract-baseline`
- Goal: 以 `container_runtime` 为试点把 Docker/runtime 错误映射到 module ports semantic，再由 application 映射到 public app error。
- Validation: `go test ./internal/module/container_runtime/... -count=1`
- Review Focus: infrastructure sentinel 不直接成为 HTTP contract。

### Slice 7: Timeout 策略基线

- Task Slug: `2026-06-13-backend-timeout-policy-baseline`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-timeout-policy-baseline` 生成
- Depends On: 无
- Goal: 定义统一 timeout owner 和层级验证，先迁移 Redis/Docker/HTTP 最关键路径，不全仓库一次性替换。
- Validation: config tests、受影响模块 tests。
- Review Focus: 依赖超时不能超过请求总超时，不引入不可取消 sleep。

### Slice 8: Circuit breaker 基础设施试点

- Task Slug: `2026-06-13-backend-circuit-breaker-foundation`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-circuit-breaker-foundation` 生成
- Depends On: `2026-06-13-backend-redis-error-boundary`
- Goal: 引入熔断器基础能力并接入一个外部依赖试点。
- Validation: breaker 单测、试点模块 tests。
- Review Focus: 熔断打开时的降级语义明确，不把所有依赖失败都变成相同 500。

### Slice 9: 错误率 metrics

- Task Slug: `2026-06-13-backend-error-rate-metrics`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-error-rate-metrics` 生成
- Depends On: `2026-06-13-backend-context-logging-contract`
- Goal: 在 HTTP/middleware 边界采集成功/失败率指标，避免每个 handler 手写。
- Validation: middleware/metrics tests。
- Review Focus: Prometheus label 基数受控，不使用带 ID 的原始 URL path。

### Slice 10: 错误处理文档与 Runbook

- Task Slug: `2026-06-13-backend-error-runbook-docs`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-error-runbook-docs` 生成
- Depends On: Slice 1-9 中已落地的事实
- Goal: 把已实现的错误分类、日志脱敏、熔断和 metrics 写入架构/运维事实源。
- Validation: `bash scripts/check-workflow-governance.sh`、`git diff --check -- <touched-docs>`
- Review Focus: 文档只能写已经落地的事实，不提前声明未实现能力。

### Slice 11: 核心链路 application 错误迁移

- Task Slug: `2026-06-13-backend-application-error-migration-core`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-application-error-migration-core` 生成
- Depends On: `2026-06-13-backend-error-contract-baseline`
- Goal: 迁移认证、Flag 提交、容器生命周期三条核心路径中的直接 `return err` / `ErrInternal` 滥用。
- Validation: 相关模块 tests、HTTP system critical scenarios。
- Review Focus: 用户可见错误稳定，internal cause 不进入响应正文。

### Slice 12: 事务错误包装规范

- Task Slug: `2026-06-13-backend-transaction-error-wrapping`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-transaction-error-wrapping` 生成
- Depends On: `2026-06-13-backend-error-contract-baseline`
- Goal: 对 GORM transaction 内错误补充步骤上下文和 guardrail。
- Validation: source guardrail、受影响模块 tests。
- Review Focus: 包装信息便于定位，但不改变业务分支语义。

### Slice 13: 资源关闭与 sleep 清理

- Task Slug: `2026-06-13-backend-resource-close-and-sleep-cleanup`
- Status: `not-started`
- Plan: 待 `scripts/start-implementation.sh backend-resource-close-and-sleep-cleanup` 生成
- Depends On: `2026-06-13-backend-context-logging-contract`
- Goal: 按 owner 分批清理高风险 `Close()` 错误忽略和不可取消 `time.Sleep`。
- Validation: source guardrail、受影响 tests。
- Review Focus: 不为了检查 `Close` 错误引入无意义日志或伪错误处理。

## Dependency Graph

```text
Slice 1 sensitive log sanitizer
  └─> Slice 3 context logging
      ├─> Slice 4 SafeGo
      ├─> Slice 9 metrics
      └─> Slice 13 close/sleep cleanup

Slice 2 error contract baseline
  ├─> Slice 5 Redis error boundary
  │   └─> Slice 8 circuit breaker foundation
  ├─> Slice 6 container runtime error boundary
  ├─> Slice 11 application error migration
  └─> Slice 12 transaction error wrapping

Slice 7 timeout baseline can run independently after Slice 1.
Slice 10 docs/runbook absorbs only landed facts from prior slices.
```

## Integration Validation

- [ ] 所有 completed slice 的 implementation plan 已归档或状态已更新。
- [ ] `bash scripts/check-workflow-governance.sh` 通过。
- [ ] 后端关键模块 focused tests 通过。
- [ ] 独立 review findings 已处理或明确阻塞。
- [ ] 已落地事实同步到 `docs/architecture/backend/` 或 `docs/operations/`。

## Completion Criteria

- [ ] 每个 slice 有独立 task slug、startup gate、implementation plan 和 review evidence。
- [ ] 高风险安全项完成：敏感日志不记录 password/token/secret 明文。
- [ ] 错误边界 guardrail 已建立并至少覆盖一个外部依赖试点。
- [ ] 日志、异步、metrics 和运维文档只描述已落地能力。
- [ ] 无 blocker 级 residual risk。
