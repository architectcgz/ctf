# Event Wakeup Keeps Durable State Owner

## 问题描述

手动关闭实例时，直接把事件当成唯一清理触发源会让运行时资源回收依赖进程内投递是否成功。当前项目的 `platform/events` 是进程内事件总线，不是跨进程消息队列；如果用它替代 `instances.status = stopping` 扫描，多副本、重启或事件处理失败时都可能留下未清理实例。

## 原因分析

实例生命周期的可靠事实源是数据库状态，运行时清理是后台维护能力。事件适合缩短等待时间，但不适合承载“必须执行”的持久状态转移。

如果只做事件清理，会出现这些问题：

- 事件在进程内同步分发，不能跨 API 节点可靠传播。
- HTTP 请求上下文不应该承载容器删除、网络删除和端口释放这类长耗时副作用。
- 清理失败后仍需要靠数据库状态恢复，否则同一实例可能长期停在半清理状态。

## 解决方案

本次优化采用“事件唤醒 + 数据库状态兜底”：

- `DestroyInstance` 成功把实例标记为 `stopping` 后，只发布 best-effort wakeup event。
- maintenance service 订阅事件后只投递内部唤醒信号，不在事件 handler 中直接清理。
- `RunStoppingCleanupLoop` 收到信号后复用原来的 stopping cleanup dispatch。
- 轮询继续保留；事件丢失或进程重启后仍能从 `status = stopping` 恢复。
- 多节点部署通过 Redis lock 避免多个 API 副本同时处理同一批 `stopping` 实例。

## 收获

好的异步设计不是“把轮询换成事件”，而是先明确哪个状态是事实源，再决定事件是事实、命令还是提示。对运行时资源回收这类必须最终完成的副作用，事件更适合做低延迟唤醒，不能替代可重扫、可恢复、可审计的 durable state。

## 沉淀状态

- 状态：仅项目保留
- Owner：`feedback/` 与本次 runtime cleanup implementation plan
- 链接：`docs/plan/archive/impl-plan/2026-06/2026-06-09-runtime-stopping-cleanup-optimization-implementation-plan.md`
