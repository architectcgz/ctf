# CTF Backend 架构风格决策

## 1. 决策结论

CTF 后端采用：

- `Modular Monolith`
- `Clean-ish`
- `DDD-lite`
- `CQRS-lite`

更准确地说，当前后端的目标形态是：

**`模块化单体 + Clean-ish + DDD-lite + CQRS-lite`**

这份约束遵守 workspace 共享 Go 规范 `/home/azhi/workspace/docs/go-code-style-guide.md`，但不会把 CTF 生搬硬套成“多服务优先”目录。

## 2. 为什么这样选

CTF 后端已经具备三个明确特征：

1. 有稳定的模块边界
   - `auth`
   - `challenge`
   - `practice`
   - `contest`
   - `system`
   - `assessment`

2. 有强状态与一致性要求
   - 练习实例生命周期
   - 竞赛状态流转
   - 审计、通知、积分、提示解锁

3. 读写关注点已经开始分化
   - 命令侧更关注事务、幂等、状态推进
   - 查询侧更关注聚合、投影、分页和只读拼装

如果继续把所有逻辑堆进“大 handler + 大 service + 大 repository”，后续可维护性会继续恶化；但如果直接上重型 DDD、重型 CQRS、CommandBus / QueryBus，又会引入不必要的工程噪音。

## 3. 与共享 Go 规范的映射关系

workspace 共享 Go 规范中的“owner / service boundary”在 CTF 里应映射成“模块边界”，不是物理微服务边界。

对应关系如下：

- `service boundary` -> `module boundary`
- `internal/services/<service>` -> `internal/module/<module>`
- `transport/http -> app -> domain/ports -> infra` -> `api/http -> application -> domain/ports -> infrastructure`
- `internal/platform` -> `internal/bootstrap`、`internal/infrastructure`、`internal/platform`、`internal/middleware`
- 克制共享 `pkg/` -> 克制共享 `internal/pkg` 与 `pkg/`

## 4. 目录与分层约束

### 4.1 保持模块化单体，不改成全局大分层

CTF 不采用以下全局目录重构：

- `internal/domain`
- `internal/app`
- `internal/infra`
- `internal/transport`

原因：

- 这会把已经拆开的模块重新揉回一个大应用
- 会削弱 `challenge`、`practice`、`contest` 等模块边界

### 4.2 新模块与重构模块的推荐布局

新增模块或进行较大重构时，优先采用：

```text
internal/module/<name>/
  contracts.go
  api/http/
  application/commands/
  application/queries/
  domain/
  ports/
  infrastructure/
  runtime/
```

说明：

- `contracts.go`：对外暴露的消费契约
- `api/http`：Gin handler，仅做协议适配
- `application`：use case 编排、事务边界、规则拼装
- `domain`：领域对象、状态机、纯业务规则
- `ports`：由消费方定义的最小依赖接口
- `infrastructure`：GORM / Redis / 外部依赖适配
- `runtime`：模块内唯一 wiring 层；可选但推荐

不再把根包兼容壳视为目标结构的一部分。迁移过程中若临时保留转发文件，应视为待删除遗留，而不是长期门面层。

现有平铺模块暂时允许保留，但新增读模型、查询模块、复杂重构不应继续默认落成 `handler.go + service.go + repository.go` 的单包大文件模式。

### 4.3 composition 是唯一装配入口

模块依赖装配统一收敛到：

- `internal/app/composition`

禁止在 handler、repository、测试辅助代码里偷偷拼装跨模块实现依赖，再把模块边界绕开。

## 5. 依赖方向

推荐依赖方向如下：

```text
api/http -> application -> domain/ports
                            |
                       infrastructure

runtime -> api/http + application + domain + ports + infrastructure
```

必须遵守：

- `api/http` 不直接写 SQL，不直接编排复杂业务
- `application` 负责 use case 编排、事务边界、错误收敛
- `domain` 只承载领域规则，不感知 HTTP、SQL 和 SDK
- `ports` 只定义依赖抽象，不依赖实现
- `infrastructure` 负责数据库、缓存、外部系统适配
- 跨模块调用优先依赖 `contracts.go` 暴露的契约，不依赖对方内部子包实现

## 6. CQRS-lite 的适用原则

满足以下任一条件时，应优先考虑拆出 readmodel / query module：

- 查询和写入关注点明显不同
- 查询会聚合多个来源的数据
- 写路径需要事务与状态推进，而读路径只需要投影
- 模块内部的 query 逻辑已经开始膨胀

已经符合这条规则的典型模块：

- `teaching_readmodel`
- `practice_readmodel`

## 7. 明确不采用的做法

CTF 当前阶段不采用：

- Event Sourcing
- 全局 CommandBus / QueryBus
- 全局 Generic Repository
- “所有东西都抽接口”
- 以“服务化”为名把模块边界再打散成 RPC 泥球

## 8. 增量落地规则

从现在开始，后端按以下顺序收敛：

1. 新增模块、读模型模块，直接按分层结构落地
2. 当前正在重构的模块，优先把查询侧先抽离
3. 旧平铺模块在修改到一定复杂度时，再做局部拆分
4. 不做无边界、全仓一次性“大换血”

## 9. 评审准则

Code review 默认按以下问题判断是否偏离规范：

1. 是否把业务规则塞进 handler
2. 是否把事务边界散落在 repository 或 HTTP 层
3. 是否让模块直接依赖别的模块内部实现
4. 是否把复杂查询和复杂命令继续塞进同一个大 service
5. 是否为了“抽象感”制造无意义接口
