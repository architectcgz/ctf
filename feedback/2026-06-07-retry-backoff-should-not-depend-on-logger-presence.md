# retry backoff should not depend on logger presence

## 问题描述

在接 `contest realtime stream consumer` 时，循环里一度出现了这种写法：

- 只有 `err != nil && logger != nil` 时，才进入失败后的 `250ms` 退避分支。

这会把“是否退避”错误地绑定到“是否能打日志”上。`logger` 本来只是可选观测能力，不应该决定运行时节奏；一旦 `err != nil` 但 `logger == nil`，循环就会跳过等待，退化成快速重试甚至热循环。

## 原因分析

- 失败后的退避属于运行时正确性问题，owner 是消费循环本身，而不是日志设施。
- `logger` 为 `nil` 往往只是测试、降级路径或局部接线形态；这些场景下更需要避免失败自旋，而不是放宽约束。
- 把“是否记录日志”和“是否进入 backoff”写在同一个条件里，表面上省一层缩进，实际上把两个独立语义耦合了，review 时也更容易漏看。

## 解决方案

- 先按失败语义决定是否退避：`if err != nil { ... backoff ... }`
- 日志能力单独作为可选分支处理：`if logger != nil { logger.Warn(...) }`
- 对所有后台 consumer / dispatcher / poller / reconcile loop，统一套用这条检查：失败后是否退避，不能依赖 logger、metrics、trace 或其他观测组件是否存在。

## 收获

- retry / backoff 是控制流语义，不是观测副作用。
- 可选依赖只能影响“记录什么”，不能影响“失败后怎么跑”。
- review 背景循环时，要专门检查条件表达式里有没有把 pacing、cancel、retry 语义偷偷挂到 logger / metric / debug flag 上。

## 沉淀状态

- 状态：implemented
- Owner：`feedback/` 先保留这条 harness 经验；后续若同类问题重复出现，再上收到全局 backend / runtime skill
- 链接：
  - `/home/azhi/workspace/projects/ctf/code/backend/internal/module/ops/runtime/module.go`
  - `/home/azhi/workspace/projects/ctf/feedback/2026-06-07-retry-backoff-should-not-depend-on-logger-presence.md`

## 证据

- file:
  - `/home/azhi/workspace/projects/ctf/code/backend/internal/module/ops/runtime/module.go`
- command:
  - `go test ./internal/app -run 'TestBuildOpsModuleDelegatesToContainerRuntime|TestBuildContestModuleDelegatesToRuntime' -count=1`
  - `go test ./internal/module/ops/application/commands -run 'TestContestRealtime(Service|OutboxDispatcher)' -count=1`
- behavior:
  - `logger == nil` 时不再跳过失败后的 `250ms` 等待；日志是否存在只影响是否 `Warn`，不影响 retry pacing。

## Decision Log

- 2026-06-07: 发现 `err != nil && logger != nil` 会把失败退避错误地绑定到日志能力上。
- 2026-06-07: 代码已修正为“失败决定退避，logger 只决定是否记录”，并同步记录到项目 harness 的 `feedback/`。
