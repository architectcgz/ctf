# Single-Instance Shared Backports Implementation Plan

> 日期：2026-06-14
> 范围：`single-instance`
> 输入：`multi-instance` 自 `486677560` 之后的共享提交分析结果

## Objective

在 `single-instance` 分支上吸收 `multi-instance` 在 HA 第一步之后仍然属于共享基座的改动，优先补齐配置拆分、共享存储、错误契约、日志上下文与 SafeGo 基础能力，同时避免把多实例控制面、跨副本协调与 runtime node 相关 owner 一并带入。

## Non-Goals

- 不吸收多 API / 多实例控制面。
- 不吸收跨副本 outbox relay、runtime node 健康重建、placement / reservation / node binding。
- 不在本轮继续恢复之前已经放弃的 container runtime owner 回灌链。

## Source Inputs

- `multi-instance` 基线：`486677560`
- 已确认应吸收的提交：
  - `28ea772c1 refactor(backend): 拆分配置包职责文件`
  - `761c97589 fix(backend): 统一共享存储密钥路径约定`
  - `a0d1ccc5b fix(backend): 补回共享存储基础包`
  - `ba35ea9c9 feat(backend): 收口共享存储下的报告与附件`
  - `2fee39e9f fix(backend): 建立错误契约边界基线`
  - `0c745cb84 fix(backend): 增加敏感日志脱敏护栏`
  - `3d3abe121 fix(backend): 建立 context-aware 错误日志契约`
  - `c7c52f612 fix(backend): 扩展试点 debug context 日志`
  - `33d081e90 feat(backend): 补齐 SafeGo 试点与 guardrail`
- 已确认不吸收的提交：
  - `072e1c501`
  - `b5c5a668e`
  - runtime placement / reservation / node binding 相关提交

## Slices

### Slice 1: logctx/requestctx 基础包

- 状态：已完成
- 目标：先在 `single-instance` 落下 `requestctx` 与 `logctx` 包，作为后续 context logging 的基础。
- 代码范围：
  - `code/backend/internal/platform/logctx/`
  - `code/backend/internal/platform/requestctx/`
- 验证：
  - `go test ./internal/platform/requestctx ./internal/platform/logctx -count=1`

### Slice 2: config package split

- 目标：吸收 `internal/config` 的职责拆分，保持对外配置契约不变。
- 代码范围：
  - `code/backend/internal/config/`
  - 相关架构 / 运维事实源文档
- 验证：
  - `go test ./internal/config -count=1`

### Slice 3: shared storage convergence

- 目标：吸收共享存储基础包、报告与附件的 shared storage owner 收口，以及密钥路径约定统一。
- 代码范围：
  - `code/backend/internal/shared/storage/` 及相关调用点
  - `report` / `attachment` / challenge import 等直接依赖 shared storage 的模块
- 验证：
  - 根据实际触达包运行最小充分 `go test`

### Slice 4: error / logging shared baseline

- 目标：补齐错误契约、敏感日志脱敏、context-aware logging 试点调用点。
- 代码范围：
  - 错误契约基线包
  - `logsanitize`
  - middleware / auth / container runtime 等 context logging 试点
- 验证：
  - 根据实际触达包运行最小充分 `go test`

### Slice 5: SafeGo guardrail

- 目标：吸收共享异步安全能力与最小 guardrail。
- 代码范围：
  - `internal/platform` 下 SafeGo 相关包
  - 已落地的试点调用点与测试
- 验证：
  - 根据实际触达包运行最小充分 `go test`

## Review Focus

- 是否仅吸收 shared/common 基座，没有把多实例控制面混入 `single-instance`
- config、shared storage、错误契约和日志能力的 owner 是否仍然清晰
- 新增文档是否只更新当前仍然有效的事实源，不把 `multi-instance` 当时的 task plan 直接搬进来

## Rollback Notes

- 每个 slice 独立提交，必要时可以按提交粒度回退，不影响 HA 第一步基线。
