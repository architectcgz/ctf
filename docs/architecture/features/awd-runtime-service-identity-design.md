# AWD 运行态 `service_id` 主身份设计

## 目标

把 AWD 运行态目标身份从隐式 `challenge_id` 切换为显式 `service_id = contest_awd_services.id`，让实例启动、轮次结果、攻击去重、流量归因和学生工作台都以赛事服务为主键。

这份文档承接 `docs/superpowers/plans/2026-04-18-awd-runtime-target-identity-phase3.md` 的已实现结果。

## 当前状态

- `service_id` 已进入 AWD 运行态主链路。
- `awd_team_services`、`awd_attack_logs`、`awd_traffic_events` 已显式持久化或归因 `service_id`。
- 学生 AWD 实例启动入口使用 `/contests/:id/awd/services/:sid/instances`。
- `challenge_id` 保留为题目元数据和兼容展示字段。

## 核心边界

### 1. `service_id`

`service_id` 表示“某场 AWD 赛事里的一个 service”。

必须使用 `service_id` 的场景：

- 队伍服务实例启动
- checker 目标定位
- 轮次 service 结果 upsert
- 攻击日志去重
- 受害服务影响写回
- 流量事件归因
- 学生 AWD 工作台服务目录
- 已落库 service 的 checker preview 绑定

### 2. `challenge_id`

`challenge_id` 表示题目资产身份。

允许使用 `challenge_id` 的场景：

- 读取题目标题、分类、flag 前缀、镜像和题库资产
- 对外展示题目元数据
- 历史接口和导出中的兼容字段
- 通过 `contest_awd_services.challenge_id` 派生展示字段

不允许使用 `challenge_id` 的场景：

- AWD 运行态唯一键
- 实例启动主入口
- 攻击去重主条件
- 流量主归因
- 已存在 service 的 checker preview 主身份

## Redis round flag 桥接

迁移中曾通过双字段方式桥接 round flag：

- service scoped 字段
- legacy challenge scoped 字段

当前事实源已经转为 `service_id`。保留兼容字段的目的只是避免历史数据在迁移窗口内失效，不允许新链路继续把 `challenge_id` 当主键。

## 事实表约定

### `awd_team_services`

- 唯一键以 `round_id + team_id + service_id` 为准。
- `challenge_id` 只作为展示和兼容字段。
- checker 结果按 service 写入。

### `awd_attack_logs`

- 成功攻击去重以 `round_id + attacker_team_id + victim_team_id + service_id` 为准。
- `challenge_id` 由 service 关联题目派生。

### `awd_traffic_events`

- 代理采集应尽量带上 `service_id`。
- 归因优先通过 `instance.service_id -> contest_awd_services.challenge_id` 补齐展示字段。

## 前后端接口约定

学生端：

- AWD 服务实例启动使用 `/api/v1/contests/:id/awd/services/:sid/instances`。
- 学生 AWD 工作台只渲染带有有效 `service_id` 的 AWD service。

管理端：

- AWD inspector、投屏页、流量筛选和导出优先展示 service 维度。
- 如果同时展示 `service_id` 与 `challenge_id`，文案必须区分“服务”和“题目”。

## 风险与约束

- 不要把 `service_id` 命名为 `challenge_id` 或 `awd_service_id` 后再混用。
- 如果新表表达运行态事实，应优先持久化 `service_id`。
- 如果新接口需要接收已存在的 AWD service，参数应使用 `service_id`，不是 `challenge_id`。
- `challenge_id` 保留不代表可以继续承担运行态身份。

## 验收标准

- AWD 学生实例启动入口只使用 `service_id`。
- 轮次结果、攻击日志和流量归因以 `service_id` 为主。
- 前端服务匹配存在 `service_id` 时不回退到 `challenge_id`。
- 新写路径不再依赖 `runtime_config.challenge_id`。
