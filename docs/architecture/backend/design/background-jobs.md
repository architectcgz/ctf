# 后台任务调度与注册机制

> 状态：Current
> 事实源：各模块 `runtime/module.go`、`application/jobs/`
> 替代：无

## 本文档范围

| 包含 | 不包含 |
| --- | --- |
| 后台任务注册方式 | HTTP handler 注册 |
| 任务调度器类型（cron / interval / event-driven） | 具体业务逻辑实现 |
| 并发控制、失败重试、优雅停机 | 分布式任务调度系统（如 Celery） |
| 各模块后台任务清单 | 前端轮询或 WebSocket 推送 |

## 当前设计

### BackgroundJob 注册方式

后台任务通过 `runtime/module.go` 的 `BackgroundJob` 结构体注册到模块：

```go
type BackgroundJob struct {
    Name  string
    Start func(context.Context) error  // 启动任务
    Stop  func(context.Context) error  // 停止任务（可选）
}

// 或简化版本（contest、challenge）
type BackgroundJob struct {
    Name string
    Run  func(context.Context)  // 阻塞式运行
}
```

**注册时机**：
- 模块 `Build()` 函数中，将 `BackgroundJob` 列表赋值给 `Module.BackgroundJobs`
- App composition 层收集所有模块的 `BackgroundJobs`，统一启动

### 任务调度器类型

后台任务分为三种调度类型：

| 调度类型 | 触发方式 | 适用场景 | 代表任务 |
| --- | --- | --- | --- |
| **Cron** | 定时表达式（如 `0 2 * * *`） | 定期全量重建、清理、归档 | `assessment_cleaner`（画像重建） |
| **Interval** | 固定间隔轮询（如每 10 秒） | 状态机推进、轮次更新、资源清理 | `contest_status_updater`、`awd_round_updater` |
| **Event-driven** | 订阅 platform event bus 事件 | 增量更新、缓存失效、通知推送 | `profile_service.handleFlagAcceptedEvent()` |

### 并发控制

- **Redis 锁**：高并发场景通过 Redis 分布式锁防止重复执行
  - 例：`assessment/infrastructure/profile_lock_store.go` 锁定用户画像更新
  - 例：`contest/application/jobs/awd_round_updater.go` 锁定轮次更新

- **数据库乐观锁**：通过 `version` 字段或 `updated_at` 检测冲突
  - 例：`contest` 模块状态机推进时检查 `status` 版本

- **单实例执行**：部分任务通过环境变量或配置标志限制只在一个实例运行
  - 例：`container_runtime` 模块的资源清理任务

### 失败重试

- **事件驱动任务**：通过 platform outbox 保证至少一次投递，失败后自动重试
- **Cron / Interval 任务**：下一个周期自动重试，不阻塞当前周期
- **错误日志**：所有任务失败都记录到 `zap.Logger`，包含任务名称和错误详情

### 优雅停机

- **Context 传播**：所有任务接收 `context.Context`，App 关闭时通过 `context.WithCancel()` 通知
- **Stop 方法**：部分任务实现 `Stop(ctx context.Context) error` 方法，清理资源
  - 例：`assessment_cleaner.Stop()` 停止 cron 调度器

- **超时控制**：任务启动时设置超时，避免无限阻塞
  - 例：`assessment_cleaner.Start(ctx, cron, timeout)`

## 各模块后台任务清单

### assessment 模块

| 任务名称 | 类型 | 触发方式 | 职责 |
| --- | --- | --- | --- |
| `assessment_cleaner` | Cron | 配置的 `full_rebuild_cron`（如 `0 2 * * *`） | 全量重建过期画像、清理过期报告输出 |

**代码位置**：
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/assessment/application/commands/cleaner.go`

### challenge 模块

| 任务名称 | 类型 | 触发方式 | 职责 |
| --- | --- | --- | --- |
| `challenge_artifact_gc` | Interval（可选） | 配置启用时，定期扫描 | 清理过期题包、导出包、构建缓存 |

**代码位置**：
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/infrastructure/*_gc.go`

### contest 模块

| 任务名称 | 类型 | 触发方式 | 职责 |
| --- | --- | --- | --- |
| `contest_status_updater` | Interval | 每 10 秒轮询 | 推进竞赛状态机（not_started → running → ended） |
| `awd_round_updater` | Interval | 每轮次时长轮询 | 更新 AWD 轮次、生成新 flag、注入容器 |

**代码位置**：
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/application/jobs/contest_status_updater.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`

### container_runtime 模块

| 任务名称 | 类型 | 触发方式 | 职责 |
| --- | --- | --- | --- |
| `instance_runtime_cleaner` | Interval | 每 5 分钟轮询 | 清理超时未访问实例、释放容器资源 |
| `node_health_checker` | Interval | 每 30 秒轮询 | 检查 Docker 节点健康状态、标记故障节点 |

**代码位置**：
- `code/backend/internal/module/container_runtime/runtime/module.go`
- `code/backend/internal/module/container_runtime/application/jobs/`

### ops 模块

| 任务名称 | 类型 | 触发方式 | 职责 |
| --- | --- | --- | --- |
| `notification_fanout_consumer` | Event-driven | 订阅 `notification.created` outbox 事件 | 消费通知事件，推送到 WebSocket 连接 |
| `audit_log_archiver`（可选） | Cron | 配置启用时，定期归档 | 归档过期审计日志到冷存储 |

**代码位置**：
- `code/backend/internal/module/ops/runtime/module.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`

### practice 模块

| 任务名称 | 类型 | 触发方式 | 职责 |
| --- | --- | --- | --- |
| `instance_lifecycle_manager` | Interval | 每 1 分钟轮询 | 推进训练实例生命周期（provisioning → running → stopped） |

**代码位置**：
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/application/jobs/instance_lifecycle_manager.go`

## 边界

- 后台任务只负责异步处理和定期维护，不替代同步 HTTP 请求响应。
- 任务失败不影响用户请求响应，通过日志和监控告警发现。
- 分布式任务调度（如跨多个 API 副本协调）通过 Redis 锁和 outbox 模式实现，不引入独立调度系统。

## Guardrail

- `code/backend/internal/module/*/runtime/module.go`：各模块后台任务注册入口
- `code/backend/internal/module/*/application/jobs/*_test.go`：后台任务单元测试
- `code/backend/internal/app/composition/*.go`：App 层收集和启动所有模块后台任务
- `code/backend/internal/module/assessment/application/commands/cleaner_test.go`：覆盖 cron 任务调度
- `code/backend/internal/module/contest/application/jobs/*_test.go`：覆盖 interval 任务轮询

## 已知限制

- 当前没有统一的任务监控面板，需要通过日志和 Prometheus metrics 监控任务执行状态。
- 部分任务没有实现 `Stop()` 方法，优雅停机时可能需要等待当前周期完成。
- Cron 任务调度依赖第三方库（如 `robfig/cron`），不支持跨时区调度。
