# AWD 运行态服务读路径切换设计

## 目标

把 AWD 运行态的 service definition 和 readiness 读取入口从旧的 `contest_challenges.awd_*` 字段切换到 `contest_awd_services`，让赛事服务层成为 checker 配置、分值配置和 readiness 状态的主事实源。

这份文档承接 `docs/superpowers/plans/2026-04-17-awd-runtime-cutover-phase2.md` 的已实现结果。

## 当前状态

- `ListServiceDefinitionsByContest` 优先读取 `contest_awd_services.runtime_config + score_config`。
- `ListReadinessChallengesByContest` 优先读取 `contest_awd_services.runtime_config + validation`。
- `contest_challenges` 只承担赛事题目关系和编排字段。
- `contest_challenges.awd_*` 已不再作为运行态读契约。

## 设计决策

### 1. `contest_awd_services` 是运行配置事实源

凡是描述“某场 AWD 赛事中的某个 service 如何运行和如何检查”的配置，都归属 `contest_awd_services`。

包括：

- checker 类型
- checker 配置
- SLA 分
- 防守分
- preview validation 状态
- service 展示名和运行配置

### 2. `contest_challenges` 降级为关系层

`contest_challenges` 仍保留必要职责：

- 赛事与题目的关系
- 题目顺序
- 可见性
- Jeopardy 题目分值等通用编排字段

它不再承载 AWD checker、SLA、防守分、readiness validation 等 service runtime 配置。

### 3. 读路径集中在 repository 边界

运行态切换不分散到 jobs、commands 或前端临时判断，而是收口到 contest 模块的 AWD repository 查询入口。

这样 `AWDRoundUpdater`、readiness query 和管理端视图可以通过同一组 service definition 读取结果工作，避免不同调用方各自理解旧字段和新字段。

## 数据流

当前读路径：

1. 管理端保存 AWD service 配置。
2. 配置写入 `contest_awd_services.runtime_config`、`score_config` 和 validation 字段。
3. readiness 查询读取赛事 service 列表和校验状态。
4. 轮次执行读取 service definition。
5. checker 结果写入运行态事实表。

## 兼容边界

当前迁移已经结束 `contest_challenges.awd_*` 的主事实源地位。若历史数据仍存在旧字段，只能作为迁移核对或回填来源，不允许新功能继续依赖。

新代码约定：

- 不新增 `contest_challenges.awd_*` 读写。
- 不从 `runtime_config.challenge_id` 反推 service 与题目关系。
- 不在前端用 challenge relation 合成 service runtime 配置。

## 风险与约束

- readiness 和 round updater 必须使用同一个配置口径，否则会出现“赛前检查通过但轮次执行失败”的不一致。
- `contest_challenges` 仍被通用竞赛题目列表使用，不能直接删除关系层。
- 若后续扩展 service 模板能力，应继续写入 `contest_awd_services`，不要回流到 relation 表。

## 验收标准

- 运行态 service definition 来源于 `contest_awd_services`。
- readiness 来源于 `contest_awd_services`。
- 旧 relation 字段不再作为新运行态配置契约。
- 管理端、readiness、checker runner 使用一致的 service 配置口径。
