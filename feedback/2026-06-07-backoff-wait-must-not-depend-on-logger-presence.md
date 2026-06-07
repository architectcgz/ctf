# backoff wait must not depend on logger presence

## 问题描述

review 指出一类容易被忽略的控制流问题：失败后的 sleep/backoff 不能写成 `err != nil && logger != nil` 这样的合并条件。这样一来，只要 `logger == nil`，失败分支就既不会打日志，也不会等待，循环会立即重试，形成无日志的忙等。

## 原因分析

- “是否等待” 属于控制流和稳定性语义，决定失败后系统会不会进入快速重试。
- “是否打日志” 属于观测语义，只影响可观测性，不应改变重试节奏。
- 把两者绑在同一个 `if` 上，会让一个本来只是降级观测能力的空 logger，意外改变运行时行为。

## 解决方案

- 把失败分支拆成两段：先判断 `err != nil` 决定是否进入退避等待，再在分支内单独判断 `logger != nil` 决定是否输出 warn。
- 对长循环 consumer / worker / poller 的 review，单独检查“错误后的等待”“日志输出”“退出条件”是不是各自独立 owner。

## 收获

- 控制流不能依赖观测条件；logger、metric、trace 缺失时，系统行为应尽量保持一致。
- review worker / consumer 时，除了看业务语义和幂等，还要看失败路径会不会因为 `nil logger`、关闭 debug 开关等非功能性条件退化成忙等。

## 沉淀状态

- 状态：仅项目保留
- Owner：当前先沉淀在项目 `feedback/`，作为 harness 复盘条目；若后续再次出现同类问题，再上收到共享 review / backend skill
- 链接：
  - `/home/azhi/workspace/projects/ctf/code/backend/internal/module/ops/runtime/module.go`
  - `/home/azhi/workspace/projects/ctf/feedback/2026-06-07-backoff-wait-must-not-depend-on-logger-presence.md`

## 证据

- file:
  - `/home/azhi/workspace/projects/ctf/code/backend/internal/module/ops/runtime/module.go`
- command:
  - `go test ./internal/module/ops/application/commands -count=1`
  - `go test ./internal/module/contest/application/commands ./internal/module/ops/application/commands ./internal/module/contest/runtime -count=1`
- behavior:
  - `ConsumeOnce` 返回错误时，即使 `logger == nil`，consumer 仍会执行 `250ms` 退避等待，而不是立即重试。

## Decision Log

- 2026-06-07: review 指出 `err != nil && logger != nil` 会把“是否等待”和“是否记录”错误耦合。
- 2026-06-07: 当前仓库代码已采用“日志判断分离，等待始终执行”的写法，并把经验记录到项目 `feedback/`。
