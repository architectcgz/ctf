# 事件总线契约设计

> 状态：Current
> 事实源：`code/backend/internal/platform/events/`、各模块事件发布代码
> 替代：无

## 本文档范围

| 负责 | 不负责 |
|------|--------|
| 定义事件 payload 结构规范与演进策略 | 定义双轨事件发布策略（强事件 / 弱事件）→ `docs/architecture/features/事件发布与降级策略.md` |
| 说明消费方注册机制与订阅规则 | Event Outbox 持久化实现细节 → `docs/architecture/backend/02-database-design.md` |
| 规定事件审计要求与日志规范 | 具体业务事件的语义定义 → 各模块文档 |
| 维护事件命名约定与版本演进规则 | 消息队列部署配置 → `docs/operations/` |

## 当前设计

### 组件边界

| 本设计负责 | 本设计不负责（见其他文档） |
|-----------|-------------------------|
| 事件 Payload 结构规范和命名约定 | 事件发布策略（强事件 vs 弱事件）→ `事件发布与降级策略.md` |
| 事件版本演进策略和兼容性原则 | 具体事件的业务语义 → 各模块文档和 `docs/contracts/` |
| 消费方注册机制和 Handler 规范 | 消息队列的部署和配置 → `docs/operations/` |
| 事件审计要求和日志记录 | - |

### 事件总线架构

**代码位置**：`code/backend/internal/platform/events/bus.go`

#### 核心接口

```go
type Event struct {
    Name    string
    Payload any
}

type Handler func(ctx context.Context, evt Event) error

type Bus interface {
    Subscribe(name string, fn Handler)
    Publish(ctx context.Context, evt Event) error
}
```

#### 实现特性

- **内存实现**：当前使用 `inMemoryBus`，支持同一进程内模块间通信
- **同步调用**：`Publish` 顺序调用所有订阅者，收集错误并通过 `errors.Join` 返回
- **多消费者支持**：同一事件名可注册多个 Handler，按注册顺序执行
- **线程安全**：使用 `sync.RWMutex` 保护订阅者列表

### 事件 Payload 结构规范

#### 1. 命名约定

**事件名格式**：`<module>.<entity>.<action>`

**示例**：
- `contest.submission.created` - 竞赛提交创建
- `challenge.flag.solved` - 题目解出
- `awd.round.started` - AWD 轮次开始
- `practice.progress.updated` - 刷题进度更新

**命名规则**：
- 使用小写字母和点号分隔
- 模块名在前，实体类型次之，动作在后
- 动作使用过去时态（如 `created`、`updated`、`deleted`）
- 禁止使用缩写或模糊名称

#### 2. Payload 结构

**基本原则**：
- Payload 使用 `any` 类型，由发布方和消费方约定具体结构
- 优先使用结构体而非 map，提供类型安全
- Payload 应包含足够信息，避免消费方再次查库

**推荐结构**：
```go
type SubmissionCreatedPayload struct {
    SubmissionID int64
    ContestID    int64
    UserID       int64
    ChallengeID  int64
    IsCorrect    bool
    SubmittedAt  time.Time
}
```

**禁止模式**：
- 只传递 ID，强迫消费方查库
- 传递整个 ORM 实体（耦合数据库层）
- 传递未导出字段的结构体

#### 3. 必填字段约定

所有事件 Payload 应包含以下基础字段（通过结构体嵌入或显式定义）：

| 字段 | 类型 | 说明 |
|------|------|------|
| 实体 ID | `int64` | 触发事件的主实体 ID（如 `SubmissionID`） |
| 时间戳 | `time.Time` | 事件发生时间，必须使用 UTC |
| 操作者 ID | `int64`（可选） | 触发事件的用户 ID（系统事件可省略） |

#### 4. 时间戳规范

- 所有时间字段必须使用 `time.Time` 类型
- 赋值时必须使用 `time.Now().UTC()`
- 禁止使用字符串或时间戳秒数表示时间
- 参考：`docs/architecture/backend/design/contest-status-state-machine.md` 中的时间口径

### 事件版本演进策略

#### 1. Payload 兼容性原则

**向后兼容变更（允许）**：
- 新增可选字段
- 扩展枚举值
- 添加新的嵌套结构

**不兼容变更（禁止）**：
- 删除现有字段
- 修改字段类型
- 重命名字段

#### 2. 版本标识

当前不强制要求 Payload 包含版本号，但推荐以下模式：

**方案 A：事件名版本化**
```go
const (
    SubmissionCreatedV1 = "contest.submission.created.v1"
    SubmissionCreatedV2 = "contest.submission.created.v2"
)
```

**方案 B：Payload 版本字段**
```go
type SubmissionCreatedPayload struct {
    Version     int // 固定为 1 或 2
    SubmissionID int64
    // ...
}
```

**选择建议**：
- 小规模变更使用向后兼容字段（新增可选字段）
- 结构重大变更使用事件名版本化（v1 → v2）
- 避免使用 Payload 版本字段（消费方仍需 type assertion）

#### 3. 废弃流程

1. **标记废弃**：在代码注释中标记旧事件为 `Deprecated`
2. **并行发布**：同时发布旧事件和新事件（至少保持一个版本周期）
3. **消费方迁移**：逐步迁移所有消费方到新事件
4. **删除旧事件**：确认无消费方后删除旧事件发布代码

### 消费方注册机制

#### 1. 订阅时机

**应用启动阶段**：
- 所有订阅应在应用启动时完成（`main.go` 或模块初始化函数）
- 禁止在运行时动态注册订阅者（除测试外）

**代码位置示例**：
```go
// code/backend/cmd/api/main.go
func initEventBus(deps *Dependencies) {
    bus := deps.EventBus
    
    // Contest 模块订阅
    contestSvc := deps.ContestService
    bus.Subscribe("challenge.flag.solved", contestSvc.OnFlagSolved)
    
    // Practice 模块订阅
    practiceSvc := deps.PracticeService
    bus.Subscribe("challenge.flag.solved", practiceSvc.OnFlagSolved)
}
```

#### 2. Handler 规范

**签名要求**：
```go
func (s *Service) OnEventName(ctx context.Context, evt events.Event) error
```

**实现要求**：
- 必须接收 `context.Context`，遵守 ctx 取消信号
- 必须对 `evt.Payload` 进行类型断言并检查成功
- 断言失败应返回明确错误（不应 panic）
- 处理逻辑应幂等（同一事件重复投递不会产生副作用）

**示例实现**：
```go
func (s *ContestService) OnFlagSolved(ctx context.Context, evt events.Event) error {
    payload, ok := evt.Payload.(SubmissionCreatedPayload)
    if !ok {
        return fmt.Errorf("invalid payload type for contest.submission.created")
    }
    
    // 幂等处理逻辑
    return s.updateScoreboard(ctx, payload.ContestID, payload.UserID)
}
```

#### 3. 错误处理

**原则**：
- Handler 返回错误会通过 `errors.Join` 收集并返回给发布方
- 弱事件发布方会记录日志但不阻塞业务（参见 `docs/architecture/features/事件发布与降级策略.md`）
- 强事件发布方会阻塞事务直到所有 Handler 成功或进入 Outbox 重试

**建议**：
- 非关键消费方应捕获错误并记录日志，返回 `nil`（避免影响其他消费者）
- 关键消费方应返回错误，触发发布方重试或回滚事务

### 事件审计要求

#### 1. 发布方日志

**弱事件发布**（使用 `publishWeakEvent` 模式）：
```go
func (s *Service) publishWeakEvent(ctx context.Context, evt platformevents.Event) {
    if s == nil || s.eventBus == nil {
        return
    }
    if err := s.eventBus.Publish(ctx, evt); err != nil {
        s.log.Warn("publish_event_failed", 
            zap.String("event", evt.Name), 
            zap.Error(err))
    }
}
```

**强事件发布**：
- 通过 Outbox 持久化，Outbox Dispatcher 负责日志
- 代码位置：`code/backend/internal/platform/events/outbox_dispatcher.go`

#### 2. 消费方日志

**推荐日志点**：
- 事件接收时：记录事件名和关键 Payload 字段（脱敏）
- 处理成功时：记录 info 级别日志
- 处理失败时：记录 error 级别日志，包含错误原因

**示例**：
```go
func (s *Service) OnFlagSolved(ctx context.Context, evt events.Event) error {
    payload, ok := evt.Payload.(FlagSolvedPayload)
    if !ok {
        logctx.Error(ctx, s.log, "event_payload_type_mismatch",
            zap.String("event", evt.Name),
            zap.String("expected_type", "FlagSolvedPayload"))
        return fmt.Errorf("invalid payload type")
    }
    
    logctx.Info(ctx, s.log, "event_received",
        zap.String("event", evt.Name),
        zap.Int64("challenge_id", payload.ChallengeID),
        zap.Int64("user_id", payload.UserID))
    
    if err := s.processEvent(ctx, payload); err != nil {
        logctx.Error(ctx, s.log, "event_processing_failed",
            zap.String("event", evt.Name),
            zap.Error(err))
        return err
    }
    
    logctx.Info(ctx, s.log, "event_processed",
        zap.String("event", evt.Name))
    return nil
}
```

#### 3. 审计范围

**必须记录的事件类型**：
- 权限变更事件（如角色分配、权限授予）
- 资源分配事件（如实例创建、Flag 分发）
- 关键业务事件（如竞赛状态变更、成绩计算）

**日志字段要求**：
- 事件名
- 操作者 ID（如有）
- 目标实体 ID
- 操作时间（UTC）
- 处理结果（成功 / 失败 / 错误信息）

## 边界

### 本文档负责

- 定义事件命名约定和 Payload 结构规范
- 说明事件版本演进策略和兼容性原则
- 规定消费方注册机制和 Handler 实现规范
- 维护事件审计要求和日志规范

### 本文档不负责

- 双轨事件发布策略（强事件 / 弱事件）→ `docs/architecture/features/事件发布与降级策略.md`
- Event Outbox 的持久化机制和重试策略 → `docs/architecture/backend/02-database-design.md`
- 具体业务事件的语义定义和使用场景 → 各模块文档
- 消息队列的部署配置和运维策略 → `docs/operations/`

## Guardrail

### 事件命名检查

- 目前无自动化检查
- Review 时需确认：
  - 事件名是否符合 `<module>.<entity>.<action>` 格式
  - 动作是否使用过去时态
  - 是否避免缩写和模糊名称

### Payload 结构检查

- Payload 是否使用结构体而非 map
- 是否包含必填字段（实体 ID、时间戳）
- 时间字段是否使用 `time.Time` 和 UTC
- 是否避免传递 ORM 实体或未导出字段

### Handler 实现检查

- Handler 是否进行类型断言并检查成功
- 是否处理 ctx 取消信号
- 逻辑是否幂等（重复投递不产生副作用）
- 错误处理是否符合弱事件 / 强事件策略

### 审计日志检查

- 发布方是否记录弱事件发布失败日志
- 消费方是否记录事件接收、处理成功和处理失败日志
- 关键业务事件是否包含完整审计字段

## 已知限制

### 1. 缺乏跨进程通信支持

- 当前只支持内存事件总线（`inMemoryBus`）
- 多实例部署时，事件无法跨进程传播
- 未来可考虑引入消息队列（如 Redis Pub/Sub、RabbitMQ）

### 2. Payload 类型安全不足

- Payload 使用 `any` 类型，缺乏编译期类型检查
- 消费方必须手动类型断言，容易出错
- 建议未来引入泛型或代码生成工具

### 3. 事件版本管理缺失

- 当前无统一的事件版本标识机制
- Payload 不兼容变更缺乏迁移工具和流程
- 建议未来引入事件注册表和版本管理工具

### 4. 缺乏自动化检查

- 无 CI 检查确保事件名符合命名约定
- 无架构测试验证 Payload 结构规范
- 无工具检测 Handler 幂等性和错误处理
- 建议添加静态分析和运行时监控

### 5. 订阅关系不透明

- 当前无工具查看事件订阅关系
- 难以追踪某个事件的所有消费方
- 建议添加事件订阅关系可视化工具或文档生成器
