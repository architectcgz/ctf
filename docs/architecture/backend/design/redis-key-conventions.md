# Redis 键命名规范与 TTL 策略

> 状态：Current
> 事实源：`code/backend/internal/module/*/infrastructure/cachekeys/`
> 替代：无

## 本文档范围

| 负责 | 不负责 |
|------|--------|
| 定义 Redis 键命名规范（前缀、分隔符、业务标识） | 规定 Redis 集群部署架构或主从复制策略 |
| 说明 TTL 设置策略与键过期管理 | 定义 Redis 持久化配置或内存淘汰策略 |
| 列举现有 Redis 键清单与代码位置 | 说明 Redis 连接池配置或客户端选型 |
| 规定键冲突避免机制与命名空间规则 | 承诺缓存一致性或分布式锁的容错保证 |

## 当前设计

### 命名规范

所有 Redis 键通过各模块的 `infrastructure/cachekeys` 包统一管理，遵循以下规范：

#### 1. 命名空间

- 全局命名空间前缀：`ctf:`（`code/backend/internal/module/contest/infrastructure/cachekeys/redis_keys.go` 中的 `redisNamespace`）
- 不带命名空间的键仅用于模块内局部缓存（如 `challenge:solved_count:`）

#### 2. 键结构

```
[namespace:]<business_context>:<entity_type>:<id_or_qualifier>[:<sub_qualifier>]
```

**示例**：
- `ctf:contest:detail:123` - 竞赛详情缓存
- `ctf:rank:contest:456:user` - 竞赛用户排行
- `challenge:solved_count:789` - 题目解出人数（模块内）

#### 3. 分隔符约定

- 使用冒号 `:` 分隔层级
- 业务上下文在前，实体类型次之，ID 或限定符在后
- 禁止使用 `-` 或 `_` 作为层级分隔符（仅用于单层内部的词组连接，如 `solved_count`）

#### 4. 命名规则

- 使用小写字母和下划线
- 业务上下文和实体类型使用单数形式
- ID 使用数字或唯一标识符
- 对于复合键（如 team + service），使用 field 而非嵌套键：`awd:123:round:5:flags` + field `456:s:789`

### 键分类与 TTL 策略

#### 1. 业务数据缓存

**Contest 模块**：
- `ctf:contest:detail:<contest_id>` - 竞赛详情缓存
  - TTL: 5 分钟（频繁读取，低变更频率）
  - 代码位置：`code/backend/internal/module/contest/infrastructure/cachekeys/redis_keys.go` → `ContestDetailKey`

- `ctf:contest:challenges:<contest_id>` - 竞赛题目列表
  - TTL: 5 分钟
  - 代码位置：`ContestChallengesKey`

**Challenge 模块**：
- `challenge:solved_count:<challenge_id>` - 题目解出人数
  - TTL: 由调用方指定（通常 5-10 分钟）
  - 代码位置：`code/backend/internal/module/challenge/infrastructure/solved_count_cache.go` → `StoreSolvedCount`

**Practice 模块**：
- `practice:progress:<user_id>` - 用户刷题进度
  - TTL: 由调用方指定
  - 代码位置：`code/backend/internal/module/practice/infrastructure/progress_cache.go`

**Assessment 模块**：
- `assessment:dimension:total:<class_id>` - 班级维度统计汇总
  - TTL: 由调用方指定
  - 代码位置：`code/backend/internal/module/assessment/application/commands/dimension_total_cache_invalidation_service.go`

#### 2. 排行榜数据（ZSET）

- `ctf:rank:global` - 全局排行榜
  - TTL: 永久（定期重建）
  - 代码位置：`code/backend/internal/module/contest/infrastructure/cachekeys/redis_keys.go` → `RankGlobalKey`

- `ctf:rank:contest:<contest_id>:user` - 竞赛个人排行
  - TTL: 永久（随竞赛生命周期）

- `ctf:rank:contest:<contest_id>:team` - 竞赛团队排行
  - TTL: 永久
  - 代码位置：`code/backend/internal/module/contest/infrastructure/awd_scoreboard_cache.go` → `RebuildContestScoreboardCache`

- `ctf:rank:contest:<contest_id>:frozen` - 封榜快照
  - TTL: 永久（手动删除或竞赛结束后清理）

#### 3. 分布式锁

- `ctf:contest:status_updater:lock` - 竞赛状态更新器锁
  - TTL: 30 秒（续租保持）
  - 代码位置：`ContestStatusUpdateLockKey`

- `ctf:awd:scheduler:lock` - AWD 调度器锁
  - TTL: 30 秒（续租保持）
  - 代码位置：`AWDSchedulerLockKey`

- `ctf:awd:round:lock:<contest_id>:<round_number>` - AWD 轮次锁
  - TTL: 30 秒（续租保持）
  - 代码位置：`AWDRoundLockKey`

#### 4. AWD 运行态数据

- `ctf:awd:<contest_id>:current_round` - 当前轮次
  - TTL: 永久（随竞赛生命周期）
  - 代码位置：`AWDCurrentRoundKey`

- `ctf:awd:<contest_id>:round:<round_id>:flags` - 轮次 Flag（HASH）
  - TTL: 永久
  - Field 格式：`<team_id>:s:<service_id>`
  - 代码位置：`AWDRoundFlagsKey` + `AWDRoundFlagServiceField`

- `ctf:awd:<contest_id>:service_status` - 服务状态（HASH）
  - TTL: 永久
  - 代码位置：`AWDServiceStatusKey`

- `ctf:awd:<contest_id>:scoreboard` - AWD 排行榜
  - TTL: 永久
  - 代码位置：`AWDScoreboardKey`

#### 5. 临时令牌与会话

- `ctf:contest:freeze_flag:<contest_id>` - 封榜标记
  - TTL: 永久（手动删除）
  - 代码位置：`ContestFreezeFlagKey`

- `ctf:awd:<contest_id>:checker_preview:<token>` - Checker 预览令牌
  - TTL: 5 分钟
  - 代码位置：`AWDCheckerPreviewTokenKey`

### 键冲突避免机制

#### 1. 模块隔离

- Contest 模块统一使用 `ctf:` 命名空间
- Challenge、Practice、Assessment 模块使用模块名作为前缀（如 `challenge:`、`practice:`）
- 模块内键定义集中在各自的 `infrastructure/cachekeys/redis_keys.go`

#### 2. ID 唯一性保证

- 所有业务 ID（`contest_id`、`challenge_id`、`user_id` 等）由 PostgreSQL 自增主键或 UUID 保证全局唯一
- 复合键使用明确的字段名（如 `user`、`team`、`round`）区分不同维度

#### 3. 代码层约束

- 禁止硬编码 Redis 键字符串
- 所有键构造通过 `cachekeys` 包的函数生成
- 函数命名遵循 `<Entity><Context>Key` 模式（如 `ContestDetailKey`、`RankContestUserKey`）

### TTL 设置策略

#### 1. 业务缓存（5-10 分钟）

适用于：
- 竞赛详情、题目列表等低变更频率数据
- 用户进度、解题统计等准实时数据

失效策略：
- 自然过期 + 手动失效（数据变更时主动删除键）

#### 2. 排行榜（永久）

适用于：
- 全局排行榜、竞赛排行榜、封榜快照

失效策略：
- 竞赛结束后统一清理
- 或由管理员手动重建

#### 3. 分布式锁（30 秒）

适用于：
- 调度器锁、状态更新锁、轮次锁

失效策略：
- 续租保持（持锁期间每 15 秒续租）
- 进程异常时自然过期，避免死锁

#### 4. 临时令牌（5 分钟）

适用于：
- Checker 预览令牌、临时授权凭证

失效策略：
- 自然过期
- 使用后主动删除

## 边界

### 本文档负责

- 定义所有 Redis 键的命名规范和构造规则
- 说明各类键的 TTL 策略和失效机制
- 维护键清单与代码位置映射

### 本文档不负责

- Redis 部署架构（单机 / 哨兵 / 集群）
- Redis 连接配置（连接池、超时、重试）
- 缓存一致性保证（由各模块 Application 层负责）
- 分布式锁的容错和正确性保证（由 `platform/locks` 或调度器负责）

## Guardrail

### 命名规范检查

- 目前无自动化检查
- Review 时需确认：
  - 新增键是否通过 `cachekeys` 包函数生成
  - 键命名是否符合 `<business>:<entity>:<id>` 结构
  - 是否使用冒号 `:` 作为层级分隔符

### TTL 一致性检查

- 业务缓存 TTL 应在 5-10 分钟范围内
- 分布式锁 TTL 应为 30 秒且配置续租逻辑
- 永久键（排行榜、AWD 数据）应有明确的清理时机

### 代码位置验证

- 每个 Redis 键应在 `infrastructure/cachekeys/redis_keys.go` 中有对应函数
- 禁止在业务代码中直接拼接 Redis 键字符串

## 已知限制

### 1. 键清单不完整

- 部分历史代码可能存在硬编码的 Redis 键
- 新增模块的键可能未及时补充到本文档

### 2. TTL 配置分散

- 部分 TTL 由调用方指定，缺乏统一默认值
- 建议未来在 `cachekeys` 包中定义常量（如 `DefaultBusinessCacheTTL = 5 * time.Minute`）

### 3. 缺乏自动化检查

- 无 CI 检查确保所有 Redis 键调用都通过 `cachekeys` 包
- 建议添加架构测试：扫描代码中的 `redis.Get/Set` 调用，确保参数来自 `cachekeys` 函数

### 4. 模块命名空间不统一

- Contest 模块使用 `ctf:` 命名空间
- 其他模块使用模块名作为前缀（无全局命名空间）
- 未来可考虑统一为 `ctf:<module>:...` 格式
