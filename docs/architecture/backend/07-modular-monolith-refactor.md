# CTF 平台后端重构方案：面向服务边界的模块化单体

> 版本：v1.0 | 日期：2026-03-21 | 状态：重构设计稿

---

## 1. 文档目标

本文档用于回答两个问题：

1. 对当前 CTF 项目，毕业设计和后续维护最合适的架构形态是什么。
2. 在不引入过高分布式复杂度的前提下，如何把现有代码重构为“服务边界清晰、分层严格、可逐步演进”的后端结构。

本文档结论：

- **运行形态**：采用模块化单体（Modular Monolith），保持单体部署。
- **设计目标**：按服务边界拆分业务域，在代码层面做到接近微服务的边界隔离。
- **演进策略**：先完成模块边界、分层、数据 ownership 和组合根治理；后续若确有容量或组织需求，再按既定边界拆分独立服务。

---

## 2. 为什么不做“真微服务”

## 2.1 毕业设计约束

结合已有开题与需求文档，当前项目的硬约束是：

- 面向**校园级**训练与竞赛平台，而非互联网级开放系统。
- 目标规模是**单机可部署、日常并发约 200 人**。
- 平台重点是**容器隔离、动态 Flag、竞赛计分、能力评估**，不是分布式平台治理。
- 开题报告已经明确采用**模块化单体架构**。

以上约束决定了本课题的优先级应是：

1. 做出完整、稳定、可演示、可答辩的业务闭环。
2. 把核心创新点放在靶场、隔离、计分和评估机制上。
3. 用清晰的架构边界体现工程能力，而不是为了“看起来高级”过早引入微服务。

## 2.2 真微服务的代价

如果当前阶段直接拆成多个独立部署服务，将额外引入以下复杂度：

- 服务注册与发现
- 网关路由与鉴权透传
- 跨服务调用失败与重试治理
- 分布式事务与最终一致性
- 服务级日志、监控、链路追踪
- 多服务本地开发和联调成本
- 多库或跨库数据一致性治理
- 部署编排、灰度和回滚流程

这些复杂度本身并不能直接提升毕业设计的核心业务价值，反而会压缩真正关键的实现与论文表达空间。

## 2.3 推荐结论

对于本项目，最佳目标形态是：

**单体部署，服务化分域；运行上仍是单体，代码上按服务边界治理。**

这意味着：

- 部署时仍是一个后端 API 进程 + PostgreSQL + Redis + Docker Engine。
- 代码中按服务边界拆模块。
- 模块内部严格分层。
- 模块之间通过接口、端口和事件交互，而不是直接穿透实现。
- 后续如需拆服务，直接按现有模块边界抽离，而不是重新设计。

---

## 3. 当前代码的主要问题

当前项目已经具备模块化单体雏形，但还没有真正达到“服务边界清晰、分层严格”的目标，主要问题如下。

## 3.1 组合根过重

`internal/app/router.go` 当前承担了过多职责：

- 创建几乎所有 repository、service、handler
- 串接跨模块依赖
- 兼顾 HTTP 路由装配和后台任务依赖准备

问题在于：

- API 层成为所有模块耦合点
- 任意模块改动都容易波及根装配文件
- 后台任务与 HTTP 运行时边界不清晰

## 3.2 分层规则执行不一致

当前代码中同时存在两种风格：

- `auth` 模块偏接口依赖，边界较清晰
- 其他模块大量直接依赖具体实现、具体 repository、具体 service

导致：

- Handler、Service、Repository 的职责不稳定
- 替换实现或单测 mock 成本不一致
- 模块间依赖方向难以约束

## 3.3 跨模块直接访问数据

多个模块直接查询其他模块关注的数据表，典型表现为：

- `teacher` 模块跨 `users / submissions / challenges / instances / skill_profiles` 聚合
- `practice` 模块直接读 `contest / contest_registration / team`
- `system` 模块依赖 `container` 的具体 repository
- `contest` 模块直接依赖 `challenge` 的具体 flag 校验实现

这类写法在单体里能运行，但会导致：

- 数据 ownership 模糊
- 模块边界名义存在、实际被 SQL 和 repository 绕开
- 后续如需拆服务，代价极高

## 3.4 模块命名与职责存在偏差

当前有两个模块边界尤其不理想：

- `container` 实际上不只是“容器模块”，而是运行时资源管理域
- `teacher` 不是业务域，而是角色视角下的聚合查询与展示服务

这会让代码结构更像“按页面/角色分组”，而不是“按领域能力分组”。

## 3.5 文档与实现存在落差

当前架构文档已经提出：

- 模块化单体
- 分层架构
- 接口依赖
- 事件总线

但代码尚未完全落地这些原则。重构的一个目标就是让**文档口径、代码结构、答辩表述**三者一致。

---

## 4. 重构目标

本次架构重构不以“拆成多个可独立部署服务”为目标，而以以下目标为准。

## 4.1 目标一：保留单体部署

部署拓扑保持如下：

- `frontend`：Vue 3 SPA
- `backend-api`：单个 Go API 进程
- `postgresql`
- `redis`
- `docker-engine`

不引入：

- 服务注册中心
- RPC 框架
- MQ 作为核心必需依赖
- 多个独立业务进程
- 多库拆分

## 4.2 目标二：按服务边界治理代码

代码层面将后端拆成多个**服务域模块**，每个模块对外只暴露：

- application 接口
- 少量对外 contract
- 必要的事件定义

禁止对外暴露：

- GORM repository 实现
- Redis key 细节
- Docker 适配细节
- 模块内部状态机细节

## 4.3 目标三：强制分层

每个模块统一为四层：

- `api`：HTTP、WebSocket、请求校验、响应映射
- `application`：用例编排、事务边界、权限判断、跨模块调用、事件发布
- `domain`：实体、领域规则、领域服务、端口定义
- `infrastructure`：GORM、Redis、Docker、定时任务、文件导出等适配器

## 4.4 目标四：定义清晰的数据 ownership

每张核心表必须有明确 owner 模块。

原则：

- 写操作只能由 owner 模块负责
- 非 owner 模块读取时，优先通过 owner 的 application/query 接口
- 确有聚合查询需求时，通过专门的 read model/query 模块承接，而不是让任意业务模块直接跨表写 SQL

## 4.5 目标五：支持后续平滑拆服务

重构完成后，应满足：

- 任一模块都可以在未来被抽离为独立进程
- 模块外部只依赖其 contract 和 application 接口
- 模块内部基础设施替换不会影响外部模块

---

## 5. 目标服务域划分

建议将当前后端收敛为以下 7 个核心域模块 + 1 个读模型模块。

## 5.1 identity

由当前 `auth + adminuser` 收敛而来。

职责：

- 用户注册、登录、刷新、登出
- CAS 对接
- 用户资料
- 用户状态与登录锁定
- 角色与权限
- 管理端用户管理

拥有数据：

- `users`
- `roles`
- `user_roles`
- 刷新会话相关缓存 key

对外能力：

- 当前用户解析
- Token 校验
- 用户查询与管理
- 角色判断

## 5.2 challenge

保留为题目域。

职责：

- 题目元数据管理
- 提示、题解、附件
- 题目拓扑模板
- 题目标签、分类、难度
- Flag 配置规则

拥有数据：

- `challenges`
- `challenge_hints`
- `challenge_hint_unlocks`
- `challenge_writeups`
- `challenge_topologies`
- `challenge_templates`

对外能力：

- 题目发布查询
- Flag 规则读取与校验接口
- 题目拓扑查询

## 5.3 runtime

由当前 `container` 与部分镜像/实例能力重命名并收敛而来。

职责：

- Docker 镜像登记与检查
- 实例生命周期管理
- 端口分配
- 网络创建与隔离
- 运行时资源清理
- 容器运行时指标

拥有数据：

- `images`
- `instances`
- `port_allocations`
- 运行时相关缓存和锁

说明：

`runtime` 比 `container` 更准确，因为这里管理的是整个靶场运行时，而不是单纯包装 Docker API。

## 5.4 practice

保留为训练域。

职责：

- 学员开始训练
- 提交 Flag
- 训练进度
- 提示解锁
- 训练时间线
- 个人积分与练习排行

拥有数据：

- 训练视角下的提交记录与进度缓存

说明：

当前 `submissions` 表物理上同时服务训练与竞赛，短期可以保留；但在逻辑边界上，训练提交与竞赛提交必须分别由 `practice` 与 `contest` 负责。

## 5.5 contest

保留为竞赛域。

职责：

- 竞赛创建、配置、状态机
- 报名审核
- 队伍管理
- 竞赛题目池
- 实时排行榜
- AWD 轮次与状态推进
- 竞赛提交与计分

拥有数据：

- `contests`
- `contest_challenges`
- `contest_registrations`
- `teams`
- `team_members`
- AWD 相关表
- 竞赛排行缓存

## 5.6 assessment

保留为评估域。

职责：

- 能力画像
- 薄弱项识别
- 靶场推荐
- 个人报告
- 班级报告

拥有数据：

- `skill_profiles`
- `reports`
- 推荐缓存

## 5.7 ops

由当前 `system` 收敛而来。

职责：

- 审计日志
- 站内通知
- 风险检测
- 系统仪表盘

拥有数据：

- `audit_logs`
- `notifications`

说明：

`ops` 比 `system` 更具体，能更好体现其“运营、审计、通知、风控”的定位。

## 5.8 teaching_readmodel

由当前 `teacher` 模块演进而来，但不建议继续当作“核心业务域服务”。

职责：

- 教师视角下的班级、学生、画像、时间线聚合查询
- 只读统计与分析视图

定位：

- 只读模块
- 不拥有核心写模型
- 不负责业务状态变更

原因：

`teacher` 本质是角色视图，不是领域本身。将其转成 `teaching_readmodel` 或 `teaching_query`，更符合职责。

---

## 6. 目标分层模型

建议每个模块统一采用如下目录形态：

```text
internal/module/<module>/
├── api/
│   ├── http/
│   └── websocket/
├── application/
│   ├── service/
│   ├── command/
│   ├── query/
│   └── event/
├── domain/
│   ├── entity/
│   ├── valueobject/
│   ├── policy/
│   ├── repository/
│   └── service/
└── infrastructure/
    ├── persistence/
    ├── cache/
    ├── runtime/
    ├── schedule/
    └── export/
```

对于当前项目，不要求一次性做到最细分目录，但必须至少先落地到：

```text
internal/module/<module>/
├── api/
├── application/
├── domain/
└── infrastructure/
```

## 6.1 api 层

职责：

- Gin handler
- WebSocket handler
- 参数绑定和校验
- application 请求对象映射
- 统一响应封装

禁止：

- 直接访问 repository
- 直接操作 GORM
- 写复杂业务判断
- 写 Redis / Docker 逻辑

## 6.2 application 层

职责：

- 用例编排
- 事务边界
- 权限判断
- 跨模块调用
- 事件发布
- DTO 转换

允许依赖：

- 本模块 domain port
- 其他模块对外 application interface
- 共享事务、日志、时钟、ID 生成等基础能力

禁止：

- 直接依赖其他模块 infrastructure
- 直接操作其他模块 repository 实现

## 6.3 domain 层

职责：

- 实体与聚合根
- 状态机
- 领域规则
- 领域服务
- repository port 定义

要求：

- 不依赖 Gin、GORM、Redis、Docker SDK
- 尽量不依赖外部框架

## 6.4 infrastructure 层

职责：

- GORM repository 实现
- Redis cache 实现
- Docker runtime adapter
- 定时任务
- 报告导出
- WebSocket 推送适配

要求：

- 只实现 port
- 不承载业务流程
- 不对外暴露内部实现

---

## 7. 模块通信规则

## 7.1 同步调用

模块间同步调用统一通过 application 接口完成。

示例：

- `practice` 调用 `challenge` 获取题目与 Flag 校验策略
- `contest` 调用 `runtime` 创建竞赛实例
- `assessment` 调用 `practice` 获取训练统计
- `ops` 调用 `runtime` 获取容器运行指标

禁止：

- `practice` 直接拿 `challenge` 的 GORM repository
- `system/ops` 直接持有 `container/runtime` 的具体 repository
- `contest` 直接依赖 `challenge.FlagService` 具体实现

## 7.2 异步事件

当前阶段建议保留**进程内事件总线**，不引入 MQ。

建议落地的事件包括：

- `practice.flag_accepted`
- `practice.hint_unlocked`
- `runtime.instance_started`
- `runtime.instance_expired`
- `contest.score_changed`
- `identity.user_logged_in`

用途：

- 更新推荐缓存
- 更新画像
- 记录审计
- 推送通知
- 刷新排行榜或统计视图

要求：

- 事件总线必须真的实现，不再只存在于文档中
- 事件消费者失败不得破坏主交易流程，除非业务要求强一致

## 7.3 查询聚合

对于教师视图、仪表盘、复盘统计等跨域查询，统一采用两种方式之一：

1. application query service 聚合
2. 专门的 read model/query 模块聚合

禁止在任意业务模块中随意跨域直连 SQL。

---

## 8. 数据 ownership 规则

建议明确如下 owner。

| 数据对象 | owner 模块 | 说明 |
|----------|------------|------|
| 用户、角色、认证状态 | `identity` | 统一用户主数据 owner |
| 题目、提示、题解、拓扑模板 | `challenge` | 统一题目主数据 owner |
| 镜像、实例、端口、运行时 | `runtime` | 统一运行时主数据 owner |
| 训练提交、训练进度、训练排行 | `practice` | 训练视角数据 owner |
| 竞赛、队伍、报名、排行、AWD | `contest` | 竞赛视角数据 owner |
| 画像、推荐、报告 | `assessment` | 评估数据 owner |
| 审计、通知、风控 | `ops` | 运营与风控数据 owner |
| 教师看板与教学复盘聚合 | `teaching_readmodel` | 只读，不拥有主数据 |

### 8.1 关于 `submissions`

`submissions` 当前是跨训练与竞赛共享表，属于现阶段边界最模糊的数据之一。

建议处理策略：

- **阶段一**：物理表不拆，但逻辑写入口分开
  - 训练提交只能由 `practice` 写
  - 竞赛提交只能由 `contest` 写
- **阶段二**：若继续演进，可拆为
  - `practice_submissions`
  - `contest_submissions`

对于毕业设计周期，建议优先完成**逻辑 owner 明确化**，不强制做物理拆表。

---

## 9. 目标目录建议

建议在当前项目基础上做最小破坏式重构，不重写整个仓库。

## 9.1 保留

- `cmd/`
- `configs/`
- `migrations/`
- `internal/app/`
- `pkg/`

## 9.2 收敛

### 当前到目标模块映射

| 当前模块 | 目标模块 |
|----------|----------|
| `auth` | `identity` |
| `adminuser` | 并入 `identity` |
| `challenge` | `challenge` |
| `container` | `runtime` |
| `practice` | `practice` |
| `contest` | `contest` |
| `assessment` | `assessment` |
| `system` | `ops` |
| `teacher` | `teaching_readmodel` |

## 9.3 公共代码治理

建议把当前散落的公共代码重新分为两类：

- `internal/platform/`
  - 配置、数据库、缓存、日志、事务、时钟、事件总线
- `internal/shared/`
  - 通用值对象、分页、错误码、响应封装

目标是让“平台能力”和“业务模块”分离，而不是把所有公共代码都堆在 `pkg` 或 `internal/pkg`。

---

## 10. 运行时装配重构

## 10.1 当前问题

当前 `buildRouterRuntime` 是超大装配点，负责：

- 创建所有模块依赖
- 串接全部跨模块调用
- 给 HTTP 和后台任务同时供依赖

## 10.2 目标

改为：

- `internal/app/composition/` 负责统一依赖注入
- 每个模块提供自己的 `Module` 构造器
- 每个模块暴露：
  - `Contracts`
  - `HTTPRegistrar`
  - `BackgroundJobs`

建议结构：

```text
internal/app/composition/
├── root.go
├── identity.go
├── challenge.go
├── runtime.go
├── practice.go
├── contest.go
├── assessment.go
├── ops.go
└── teaching_readmodel.go
```

每个模块装配器只负责：

- 构造本模块 infrastructure
- 构造本模块 application service
- 返回本模块对外 contract

组合根只负责：

- 连接模块 contract
- 注册路由
- 注册后台任务

---

## 11. 分阶段迁移方案

本重构必须采用渐进式迁移，不能一次性大翻新。

## 阶段 0：建立重构约束

目标：

- 冻结目标架构口径
- 明确模块 ownership
- 明确禁止事项

产出：

- 本文档
- 代码评审检查表
- 模块依赖规则

## 阶段 1：先收敛组合根

目标：

- 把超大 `router.go` 拆为模块装配
- 每个模块有独立注册入口

优先级：最高

原因：

- 这是后续重构的入口控制点
- 不先收敛组合根，边界治理会持续失效

## 阶段 2：先做命名与模块边界收敛

目标：

- `auth + adminuser -> identity`
- `container -> runtime`
- `system -> ops`
- `teacher -> teaching_readmodel`

优先级：高

原因：

- 不先修正模块语义，后续层次拆分会持续混乱

## 阶段 3：统一分层模板

目标：

- 在每个模块中引入 `api/application/domain/infrastructure`
- 新代码只能进入目标结构
- 旧代码按模块逐步搬迁

优先级：高

## 阶段 4：收敛跨模块依赖

目标：

- 删除跨模块具体 repository 依赖
- 删除跨模块具体 service 依赖
- 统一改成 application interface 或 port

重点模块：

- `practice`
- `contest`
- `ops`
- `teaching_readmodel`

## 阶段 5：治理数据 ownership

目标：

- 明确每张核心表的 owner
- 跨域查询改为 query service / read model
- 不再允许任意模块直接跨域查表

## 阶段 6：落地进程内事件总线

目标：

- 将推荐更新、画像刷新、通知推送、审计记录等弱一致逻辑事件化
- 减少 application service 直接串联的副作用逻辑

## 阶段 7：治理报告、任务和导出

目标：

- 异步任务统一纳入模块 background jobs
- 生命周期统一纳入 app lifecycle
- 报告导出从业务流程中显式隔离

## 阶段 8：视情况再考虑独立服务抽离

只有在以下条件同时满足时，才考虑拆独立服务：

- 单体内模块边界已经稳定
- 模块对外 contract 已经清晰
- 单机部署已成为明确瓶颈
- 答辩和交付已经不受影响

---

## 12. 每个模块的重构重点

## 12.1 identity

重点：

- 合并 `auth` 与 `adminuser`
- 对外统一导出用户查询与鉴权接口
- 隐藏用户 GORM 细节

完成标准：

- 其他模块不再直接查询 `users`
- 用户相关写操作只经过 `identity`

## 12.2 challenge

重点：

- 将题目、提示、题解、拓扑、Flag 规则收敛到同一域内
- 对外导出题目查询与 Flag 校验 contract

完成标准：

- `contest` 与 `practice` 不再依赖 `challenge` 的具体 repository 或具体 service

## 12.3 runtime

重点：

- 镜像、实例、端口、网络隔离统一归口
- Docker 适配与业务状态机分离

完成标准：

- 其他模块只通过 runtime application 接口获取实例能力
- 不再从外部模块直接碰实例表与 Docker 封装

## 12.4 practice

重点：

- 只关注训练场景
- 与 `contest` 的共享逻辑改成可复用 domain service 或 contract，而不是共享 repository

完成标准：

- 训练流程不直接读竞赛域 persistence
- 训练副作用通过事件驱动评估与推荐更新

## 12.5 contest

重点：

- 报名、队伍、题池、排行、AWD 聚合到竞赛域
- 竞赛提交与训练提交明确区分

完成标准：

- `contest` 不再直接依赖 `challenge.FlagService` 具体实现
- 排行逻辑与提交流程边界清晰

## 12.6 assessment

重点：

- 画像、推荐、报告生成职责分开
- 训练数据读取通过 query contract 获取

完成标准：

- 推荐与报告不直接依赖其他模块具体实现

## 12.7 ops

重点：

- 审计、通知、风险、仪表盘从“杂项系统模块”变成明确运营域
- 仪表盘改依赖 query/service contract，而非外部 repository

完成标准：

- `ops` 不再直接持有 `runtime` 的 persistence 实现

## 12.8 teaching_readmodel

重点：

- 明确为只读聚合模块
- 只做教学查询与复盘，不做主业务写入

完成标准：

- 不再作为核心业务域承载跨域写逻辑

---

## 13. 代码规范补充

为保证重构真正落地，建议新增以下规则。

## 13.1 新增硬规则

- Handler 只能依赖本模块 application interface
- Application 只能依赖 domain port、平台能力和其他模块对外 contract
- Infrastructure 只实现 port，不被模块外直接引用
- 跨模块禁止直接引用 `*Repository`
- 跨模块禁止直接引用 `*Service` 具体实现
- 跨模块禁止直接访问对方模块表，除非经过显式 read model/query 模块

## 13.2 Review 检查项

- 是否出现跨模块 concrete dependency
- 是否出现跨模块直接查表
- 是否出现 handler 写业务逻辑
- 是否出现 service 直接操作 GORM / Docker / Redis 细节
- 是否出现模块 owner 不清晰的数据写入

---

## 14. 验收标准

当以下条件全部满足时，认为“服务划分与分层处理基本完成”。

### 14.1 结构验收

- 所有核心模块都具有 `api/application/domain/infrastructure` 结构
- 组合根不再集中拼装全部细节
- 教师视图被重构为只读聚合模块

### 14.2 依赖验收

- 模块间无具体 repository 依赖
- 模块间无具体 service 依赖
- 模块间通过 interface / contract / event 交互

### 14.3 数据验收

- 核心表 owner 明确
- 写路径唯一
- 聚合查询集中治理

### 14.4 文档验收

- 架构文档、开题口径、代码结构三者一致
- 答辩时可以清楚说明：
  - 为什么采用模块化单体
  - 如何按服务边界治理
  - 后续如何演进为独立服务

---

## 15. 毕业设计答辩建议口径

建议在论文与答辩中统一表述为：

> 本系统采用面向服务边界的模块化单体架构。考虑到校园级部署规模、毕业设计周期以及运维可控性，系统在运行时保持单体部署，但在代码结构上按身份、题目、运行时、训练、竞赛、评估、运营等服务域拆分模块，并在每个模块内部实施 API、应用、领域、基础设施四层分离。该设计既降低了分布式系统的运维成本，又为后续按模块独立拆分服务预留了清晰边界。

这套口径具备三个优点：

- 与开题报告一致
- 与当前校园级单机约束一致
- 能体现工程设计能力，而不是停留在功能堆叠

---

## 16. 最终结论

本项目最适合的架构目标不是“现在就拆成多个独立微服务”，而是：

**以模块化单体为运行形态，以服务边界为设计目标，以严格分层为重构抓手。**

简化表达就是：

**部署保持单体，代码按服务治理。**

这是对毕业设计周期、校园级部署约束、系统复杂度和后续可维护性的综合最优解。
