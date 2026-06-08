# Go 结果时间戳收尾适合用局部 finish 闭包

## 问题描述

在 `DockerCheckerRunner.RunChecker` 里，多个 early return 都需要同时写入 `FinishedAt` 和 `Duration`。如果每个分支都手写：

- 容易漏掉某个分支的完成时间或耗时。
- 容易让 `StartedAt` / `FinishedAt` 的 UTC 业务时间要求和 `Duration` 的 monotonic 计时要求混在一起。
- 后续新增分支时，重复代码会增加 review 成本。

这类场景不是简单地把所有 `time.Now()` 改成 `time.Now().UTC()` 就结束。`Duration` 需要保留 Go `time.Time` 的 monotonic 信息，而对外结果时间戳需要按项目契约归一为 UTC。

## 原因分析

`time.Now().UTC()` 会返回 UTC 业务时间，但转换后的值不应再作为耗时计算的主输入。对于运行时耗时，应保留原始 `startedAt := time.Now()` 和 `finishedAt := time.Now()` 的配对，再用 `finishedAt.Sub(startedAt)` 计算。

同时，结果对象中的 `StartedAt` / `FinishedAt` 可能进入审计、JSON 或跨请求状态，它们应该写成 UTC。把这两种语义放在每个 early return 分支里手动维护，很容易出现某个分支只满足其中一半。

## 解决方案

在函数内部使用局部 `finish()` 闭包集中完成态收尾：

```go
startedAt := time.Now()
result := CheckerRunResult{
	StartedAt: startedAt.UTC(),
}

finish := func() {
	finishedAt := time.Now()
	result.FinishedAt = finishedAt.UTC()
	result.Duration = finishedAt.Sub(startedAt)
}
```

每个 early return 前只调用 `finish()`。这样：

- `Duration` 始终来自原始 monotonic 时间配对。
- 对外 `StartedAt` / `FinishedAt` 始终是 UTC。
- 新增 early return 时，review 只需要检查是否调用了 `finish()`。

## 收获

- “集中收尾”比“每个分支重复写字段”更符合 Go 后端里低抽象但高一致性的 good taste。
- UTC 业务时间和 monotonic 运行时计时不能混为一谈；好的实现应该让两者在代码形态上分开。
- 对多个 early return 都要写相同完成态字段的函数，局部闭包通常比提取跨包 helper 更合适，因为它能捕获本函数的局部语义，不会制造泛化工具包。

## 交叉链接

- `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
- `docs/reviews/backend/2026-06-08-worktree-review-docker-checker-runner-utc-timestamps.md`

## 沉淀状态

- 状态：仅项目保留
- Owner：CTF backend harness feedback
- 链接：`feedback/2026-06-08-go-finish-closure-for-result-timestamps.md`
