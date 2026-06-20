# Goroutine Panic Owner Boundary

## 问题描述

项目曾尝试用 `internal/shared/safego` 统一包装后台 goroutine，默认捕获 panic、记录日志并让 goroutine 结束。这个方向会把“goroutine 怎么启动”和“panic 后系统应该怎么处理”混成一个共享 helper 决策。

对 CTF 后端来说，goroutine 的 panic 后果不能由通用 helper 默认决定。不同 owner 的正确行为不同：

- HTTP 请求边界可以 recover，返回 500 并记录 request 上下文。
- 关键 root background job panic 通常应该暴露为严重进程问题，至少需要记录后重新 panic，不能静默退出。
- cron / reconcile 单轮任务可以由 owner 本地 recover，但语义应是“本轮失败，下一轮可重试”。
- 有业务状态的异步任务应把 panic 转成业务失败状态，例如报告生成任务标记失败。
- 仅用于等待 `WaitGroup.Wait()` 的辅助 goroutine 不应该挂 panic recover 语义。
- 短生命周期 fan-out 更适合 `WaitGroup` / `errgroup` 这类明确 error propagation 的机制。

## 原因分析

共享 `SafeGo` 风格 helper 的主要风险不是 recover 代码本身，而是它会隐藏 owner 决策：

- recover 后只打一条日志并退出，可能让关键后台能力死掉但进程继续对外服务。
- `nil ctx` / `nil logger` 兜底会把调用方接线错误变成静默降级。
- `WaitGroup.Add` 和 panic 语义藏进 helper 后，生命周期边界反而不如显式 goroutine 清楚。
- 测试容易退化成“必须导入某个 helper”的源码 guard，而不是验证业务 owner 的真实失败语义。

成熟 Go 项目通常也不会把所有 goroutine 套进全局 recover helper。它们更倾向于在明确边界处理 panic：HTTP handler、controller worker、stopper / supervisor、业务任务 owner，或者让 panic 作为 bug 直接暴露。

## 解决方案

CTF 后端 goroutine / worker / poller / reconcile loop 统一采用以下判断顺序：

1. 先写清 owner：谁启动 goroutine，谁负责取消、等待、错误处理、panic 后果。
2. 再判断 panic 语义：暴露为进程 bug、记录后 re-panic、转成业务失败、还是本轮任务失败后等待下一轮。
3. recover 只放在明确 owner 边界内，不新增跨项目通用 `SafeGo` / `safe goroutine` 默认包装。
4. 若一个 goroutine 没有明确 logger、ctx、task name 或失败状态 owner，先修 owner 接线，不用 no-op / inert fallback 掩盖问题。
5. `WaitGroup.Wait()` 的等待辅助 goroutine 保持最小显式写法，不包装 recover。
6. 架构测试优先锁行为边界，例如“报告 panic 后标记失败”“root job 不静默吞 panic”，不要锁“必须使用某个 helper”。

## 代码范例

root 级关键后台任务由 root owner 本地记录后重新抛出，避免后台能力静默死亡：

```go
// owner：root lifecycle。该 goroutine 承载关键后台能力，panic = 严重进程级 bug。
wg.Add(1) // 由 root owner 登记生命周期，便于 Stop 时统一等待
go func() {
	defer wg.Done() // 无论正常结束还是 panic，都先让 WaitGroup 计数归零，避免 Wait 永久阻塞
	defer func() {
		// recover 在这里的目的不是“吞掉错误让进程活下去”，而是“先保留现场再让进程崩”
		if recovered := recover(); recovered != nil {
			logger.Error("background_job_panicked",
				zap.Any("panic", recovered),         // 原始 panic 值，定位根因
				zap.String("task_name", name),       // 标明是哪个后台任务死的
				zap.ByteString("stack", debug.Stack()), // 完整栈，panic 现场只此一次
			)
			panic(recovered) // 关键：re-panic，让进程真正崩溃/被重启接管，杜绝核心能力静默假死
		}
	}()
	run(runCtx) // 实际后台逻辑；用 root 传入的 ctx，取消语义跟随 root lifecycle
}()
```

只用于等待 `WaitGroup.Wait()` 的 goroutine 保持最小裸等待，不挂 recover 语义：

```go
// owner：Stop 流程。这个 goroutine 不跑业务逻辑，只负责“等所有 worker 退出后发信号”。
done := make(chan struct{}) // 用于把“全部退出”这一事件通知给 Stop 的等待方
go func() {
	wg.Wait()    // 纯阻塞等待，自身不产生 panic，没有需要 recover 的失败语义
	close(done)  // 全部退出后关闭 channel，Stop 据此知道可以安全返回
}()
```

有业务失败状态的异步任务在业务 owner 内 recover，并写入业务失败结果：

```go
// owner：业务任务（如报告生成）。panic 不该崩进程影响其他用户，而该收敛成“这一个任务失败”。
go func() {
	defer tasks.Done() // 任务计数归零，owner 才能正确统计在途任务
	defer func() {
		// recover 在这里把“技术 panic”翻译成“业务失败状态”，用户能看到明确失败而非永久卡住
		if recovered := recover(); recovered != nil {
			service.markFailed(taskCtx, id, fmt.Errorf("report task panicked: %v", recovered))
		}
	}()
	// 正常返回的 error 走同一个失败出口：panic 和 error 最终都落到 markFailed，业务状态统一
	if err := run(ctx); err != nil {
		service.markFailed(ctx, id, err)
	}
}()
```

禁止用共享 helper 默认吞掉所有后台 panic，尤其不要用 no-op logger / inert ctx 兜底掩盖接线错误：

```go
// 反例：共享 helper 替所有调用方统一“recover + 打日志 + 安静退出”。
// 问题 1：它替 owner 做了失败决策——上面三种本应不同的语义被压成同一种“吞掉继续跑”。
// 问题 2：root 关键任务用它会静默假死（进程还在，但核心后台能力已死，无人发现）。
// 问题 3：nil ctx / nil logger 之类兜底会把“接线错误”降级成静默运行，bug 被藏起来。
safego.Go(&wg, ctx, logger, "task_name", func(ctx context.Context) {
	run(ctx)
})
```

## 收获

- panic recovery 是失败语义，不是启动语法糖。
- “后台任务不能因为 panic 拖垮全部服务”和“关键后台任务不能静默死掉”都可能成立，必须由 owner 决定。
- 共享 helper 只有在它不替 owner 做失败语义决策时才安全；否则会把技术债伪装成统一抽象。

## 沉淀状态

- 状态：archived
- Owner：`.agents/skills/ctf-backend-patterns` 的 Concurrency & Durable State 索引
- 链接：
  - `.agents/skills/ctf-backend-patterns/SKILL.md`
  - `feedback/2026-06-20-goroutine-panic-owner-boundary.md`

## 后续处理

- 撤回现有 `internal/shared/safego` 试点时，删除 helper 与只验证 helper 导入的 architecture guard。
- 三处试点应分别回到对应 owner：
  - root background job：显式 goroutine；若保留 logger 版本，采用 log + re-panic 或明确的 root lifecycle 失败策略。
  - practice async task：由 practice owner 决定是否本地 recover 并记录业务上下文。
  - runtime cleaner：由 cleaner owner 决定单轮任务失败语义，`Stop` 的 wait goroutine 保持裸等待。
