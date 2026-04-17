# AWD 对抗引擎迁移设计

> 状态：已采用

## 0. 当前阶段落地边界（2026-04-17）

当前仓库已经落地了 AWD 显式服务模型的前两层，并开始把运行态读路径逐步切到赛事服务层：

- 已落地：独立 `awd_service_templates` 题库
- 已落地：服务模板后台 CRUD 与基础管理页
- 已落地：`contest_awd_services` 赛事服务关联存储
- 已落地：管理员赛事题目列表返回 `awd_service_id / awd_template_id / awd_service_display_name`
- 已切换：`ListServiceDefinitionsByContest` 优先读取 `contest_awd_services.runtime_config + score_config`
- 已切换：`ListReadinessChallengesByContest` 优先读取 `contest_awd_services.runtime_config`，验证状态仍兼容复用 `contest_challenges`
- 未切换：workspace 编排、flag 注入目标标识、攻击提交流程仍以 `challenge_id + contest_challenges` 兼容键运转

也就是说，当前已经形成：

1. 模板层：`awd_service_templates`
2. 赛事服务层：`contest_awd_services`
3. 兼容运行层：`contest_challenges.awd_*`

其中第 1、2 层已经开始进入运行态读路径，但第 3 层仍承担兼容字段、验证状态和挑战关系承接职责，后续 runtime cutover 继续沿着“先 service 定义，再 target 身份”的顺序推进。

## 1. 背景

当前平台已经有一套可运行的 AWD 基础能力：

- 有 `awd_rounds` 轮次调度与当前轮缓存
- 有每轮 flag 生成、Redis 缓存与容器内注入
- 有 `awd_attack_logs` 攻击提交流程与重复得分去重
- 有 `awd_team_services` 服务状态记录
- 有 `awd_traffic_events` 运行时代理流量采集
- 有管理端轮次态势、攻击日志、流量摘要与排行榜面板

但这套实现仍然更接近“教学平台中的 AWD 运营功能”，还不是标准的 Attack-Defense 比赛引擎。当前最核心的缺口有三个：

1. 服务检查仍以 HTTP 探活为主，不能证明 `putflag/getflag/havoc` 语义成立
2. 赛事中的被防守对象还只是“contest challenge + team instance”的隐式组合，没有正式的 service 配置模型
3. 计分模型偏向“攻击分 + 防守分”，还没有把 SLA 作为独立得分维度固化下来

毕业设计课题要求的平台边界是：

- 既要支持攻防对抗、团队赛、实时排行榜和攻击流量监控
- 也要保留教学实验、集中管理、实训数据沉淀和复盘能力

因此目标不是把现有平台替换成一个独立赛事引擎，而是在现有平台内核上把 AWD 升级成更接近专业 A/D 平台的运行方式。

## 2. 目标

本轮迁移目标是把现有 AWD 从“轮次 + 探活 + flag 提交”升级为“以 tick 为核心的最小可用 A/D 对抗引擎”，并保持与当前平台的题目、实例、流量和教学数据体系兼容。

本轮必须达成：

1. 把每轮服务巡检升级成标准 checker 执行链路
2. 为 AWD 赛事中的每个 service 引入显式配置
3. 让每支队伍、每个 service、每一轮都有可追溯的 checker 结果
4. 把总分拆成 `sla_score / attack_score / defense_score / total_score`
5. 保持现有 flag 轮换、提交判定、流量监控和管理面板继续工作
6. 满足毕业设计对“团队对抗、自动监控攻击流量、统计防守成功次数、生成实时排行榜”的要求

## 3. 非目标

本轮不做下面这些内容：

- 外置独立 A/D 引擎服务
- checker 沙箱执行集群
- checker 镜像仓库、签名、版本分发
- 跨主机赛场网络自动编排
- 自动裁判处罚、申诉仲裁、复杂反作弊策略
- 学员侧 AWD 比赛页的整页终态重构

这些能力都可以作为后续扩展，但不属于这轮“把平台升级为可工作的 A/D 内核”的必要条件。

## 4. 方案比较

### 4.1 方案 A：继续增强现有 AWD 功能页

做法：

- 保留当前轮次、探活、flag 提交和运营面板
- 只补少量状态说明和统计细节

优点：

- 代码改动最小
- 上线最快

缺点：

- 服务状态仍然不是标准 checker 结果
- 不能提供 `putflag/getflag/havoc` 级别的语义校验
- 本质上仍是教学平台功能增强，不是 A/D 引擎

不推荐。

### 4.2 方案 B：在现有平台内嵌 A/D 引擎

做法：

- 继续复用现有 `contest / runtime / instance / traffic` 体系
- 补 `AWD service 配置 + checker runner + tick result + 三段分`
- 让现有 AWD 数据结构向标准 A/D 语义演进

优点：

- 能最大化复用现有代码和数据
- 最符合毕业设计的平台化目标
- 技术风险和改造范围可控

缺点：

- 需要引入新配置字段和新的 checker 运行抽象
- 需要同步调整排行榜和后台摘要结构

推荐采用。

### 4.3 方案 C：独立 A/D 引擎，通过接口对接当前平台

做法：

- 新起独立服务负责 tick、checker、flag store 和计分
- 当前平台只保留赛事、用户、可视化和教学沉淀

优点：

- 架构边界最清晰
- 更接近大型专业赛事实现

缺点：

- 系统一致性、部署和联调复杂度最高
- 对当前项目和毕业设计周期来说风险过大

不采用。

## 5. 总体设计

### 5.1 设计概览

采用方案 B，在现有平台中加入标准化的 AWD 领域闭环：

- `AWD Contest`：现有 `contests(mode=awd)`，负责比赛窗口、轮次和权重
- `AWD Service Template`：`awd_service_templates`，负责独立 AWD 服务模板题库
- `Contest AWD Service`：`contest_awd_services`，负责模板到赛事题目的显式服务映射
- `Legacy Contest Challenge Config`：`contest_challenges.awd_*`，当前仍承担兼容运行态配置
- `AWD Team Target`：现有 `team + challenge + instance` 关系，表示每队每服务的目标实例
- `AWD Checker`：每轮对目标实例执行 `putflag / getflag / havoc`
- `AWD Tick Result`：写入 `awd_team_services`，记录本轮服务结果和明细
- `AWD Attack Submission`：现有 stolen flag 提交流程
- `AWD Score Projection`：从 checker 和攻击结果统一重算总分

### 5.2 与现有实现的关系

保留不动：

- `awd_rounds`
- `awd_attack_logs`
- `awd_traffic_events`
- flag 轮换与容器注入机制
- runtime proxy 流量采集链路

升级语义：

- `awd_team_services` 从“服务探活状态表”升级为“每轮 service 结果表”
- `AWDRoundUpdater` 从“轮次调度 + HTTP 巡检”升级为“轮次调度 + checker 执行”
- 排行榜从“攻击分 + 防守分”升级为“攻击分 + 防守分 + SLA 分”

## 6. 数据模型

### 6.1 `awd_service_templates` 作为模板层

当前已新增独立 `awd_service_templates`，用于沉淀与复用 AWD 服务模板。

模板层职责：

- 管理独立 AWD 题库，不与普通 Jeopardy 题目语义混用
- 保存服务类型、部署模式、checker 草稿、访问配置、运行配置和 readiness 草稿状态
- 为后续赛事服务映射与多场复用提供来源

### 6.2 `contest_awd_services` 作为赛事服务层

当前已新增 `contest_awd_services`，负责把某场 AWD 赛事中的题目映射成显式服务。

核心字段：

- `contest_id`
- `challenge_id`
- `template_id`
- `display_name`
- `order`
- `is_visible`
- `score_config`
- `runtime_config`

这一层当前已经接入管理员赛事题目增删改链路，但还没有切到运行时读取。

### 6.3 `contest_challenges` 作为兼容运行层

当前运行态仍然依赖 `contest_challenges.awd_*` 作为兼容配置载体。

原因：

- `awd_round_flag_support.go`、`awd_check_run.go`、`awd_attack_submit_support.go` 仍直接依赖赛事题目关系
- 先落显式服务模型，再切运行态，可以避免在同一阶段同时改动存储层和裁判链路
- 这样可以把迁移拆成“先加法落库，再逐步切流”的更安全路径

新增字段建议：

- `awd_checker_type`
  - 类型：字符串
  - 默认值：空
  - 当前枚举：`http_standard`
- `awd_checker_config`
  - 类型：JSON 文本
  - 默认值：空对象
  - 用于存放 `putflag/getflag/havoc` 所需配置
- `awd_sla_score`
  - 类型：整数
  - 默认值：0
  - 该 service 在每轮存活时可获得的 SLA 分
- `awd_defense_score`
  - 类型：整数
  - 默认值：沿用当前轮次默认值或显式配置
  - 该 service 在防守维度使用的分值权重

当 `contest.mode != awd` 时，这些字段为空或零值，不参与普通比赛逻辑。

### 6.2 `awd_team_services` 作为每轮 service 结果表

保留现表，并扩充字段使其承载标准 checker 结果。

当前已有字段：

- `round_id`
- `team_id`
- `challenge_id`
- `service_status`
- `check_result`
- `attack_received`
- `defense_score`
- `attack_score`

建议新增：

- `sla_score`
  - 本轮因服务存活而获得的 SLA 分
- `checker_type`
  - 本轮实际执行的 checker 类型
- `checker_version`
  - 可选，便于后续兼容 checker 迭代

`check_result` 继续保留为 JSON，但其结构升级为标准 checker 执行结果。第一版统一包含：

- `check_source`
- `checker_type`
- `put_flag`
- `get_flag`
- `havoc`
- `status_reason`
- `latency_ms`
- `targets`
- `error_code`
- `error`

### 6.3 `teams`

现有 `teams.total_score` 继续作为官方总分缓存，不新增专门总分表。

细分得分不写回 `teams` 的独立列，优先在查询侧通过聚合或摘要响应返回，避免这轮把队伍模型改得过重。

## 7. Checker 契约

### 7.1 目标

本轮 checker 不追求完整兼容外部 A/D 引擎，而是定义一版能直接落在当前仓库内的最小标准契约。

每轮每个队伍服务都执行三个动作：

- `putflag`
- `getflag`
- `havoc`

### 7.2 动作语义

`putflag`

- 把当前轮 flag 写入目标服务
- 成功才说明该服务具备合法 flag 存储路径
- 失败说明服务不能正确接受本轮 flag

`getflag`

- 从目标服务取回并校验当前轮 flag
- 在宽限时间内允许上一轮 flag 继续有效
- 成功才说明服务真正可用且 flag 未丢失

`havoc`

- 做轻量语义探测
- 用来避免“HTTP 返回了 200 但业务实际不可用”的假阳性
- 第一版限定为轻量 HTTP 行为验证，不做复杂破坏流量

### 7.3 第一版 checker 类型

本轮只实现一个 checker 类型：

- `http_standard`

它通过 `awd_checker_config` 声明以下字段：

- `put_flag.method`
- `put_flag.path`
- `put_flag.headers`
- `put_flag.body_template`
- `put_flag.expected_status`
- `get_flag.method`
- `get_flag.path`
- `get_flag.headers`
- `get_flag.expected_status`
- `get_flag.expected_substring`
- `havoc.method`
- `havoc.path`
- `havoc.expected_status`

模板中允许使用：

- `{{FLAG}}`
- `{{ROUND}}`
- `{{TEAM_ID}}`
- `{{CHALLENGE_ID}}`

这版设计明显偏 HTTP，但和当前 runtime proxy、access URL、服务实例模型完全兼容，足以把现有 probe 升级成真正的 service checker。

## 8. 服务状态规则

本轮 `service_status` 继续沿用现有枚举：

- `up`
- `down`
- `compromised`

但判断规则改成：

- `up`
  - `putflag` 成功
  - `getflag` 成功
  - 若配置了 `havoc`，则 `havoc` 也成功
- `down`
  - 目标不可达
  - checker 请求超时
  - `putflag` 或 `getflag` 发生关键失败
- `compromised`
  - 服务可访问，但 `getflag` 取回值不正确
  - 或服务在本轮成功被攻击后，被写回为失陷状态

这样做的重点是把 `compromised` 从“人工感觉像被打了”变成“checker 或攻击结果明确证明服务处于失陷态”。

## 9. 计分模型

### 9.1 总体规则

官方总分拆成：

- `sla_score`
- `attack_score`
- `defense_score`
- `total_score`

其中：

- `sla_score` 反映服务是否持续在线、是否满足 checker 要求
- `attack_score` 反映偷取有效 flag 的攻击成果
- `defense_score` 反映防守方在攻击面上的守住程度

### 9.2 SLA 分

每轮每个 service：

- 状态为 `up` 时，获得该 service 配置的 `awd_sla_score`
- 状态为 `down` 或 `compromised` 时，SLA 分为 0

### 9.3 Attack 分

继续沿用现有 stolen flag 提交流程：

- 攻击方提交目标队伍的有效 flag
- 每轮同一 `attacker -> victim -> challenge` 首次成功计分
- 分值沿用当前轮次的 `round.attack_score`

### 9.4 Defense 分

第一版保持与现有实现兼容，不引入过度复杂的扣分模型。

建议规则：

- 本轮某队某 service 未被成功攻击时，记该 service 的防守分
- 一旦本轮该 service 被成功攻击，本轮防守分为 0

因此 `ApplyAttackImpactToVictimService` 的语义需要从“更新 attack_received / compromised”扩展为同时清空该轮该服务的防守得分。

### 9.5 总分写回

总分重算改为：

- `total_score = sum(sla_score) + sum(defense_score) + sum(attack_score)`

现有 `RecalculateAWDContestTeamScores` 只需从 `awd_team_services` 额外读取 `sla_score` 并并入总分，不需要整体推倒。

## 10. 轮次执行流程

每个调度 tick 的执行顺序固定为：

1. 解析当前 active round
2. 物化轮次记录
3. 生成当前轮 flag 并注入目标实例
4. 加载参赛队伍、AWD service 配置、目标实例
5. 对每个 `team + challenge` 执行 checker
6. 写入 `awd_team_services`
7. 重算队伍官方总分
8. 刷新 scoreboard cache

相比当前实现，变化点主要在第 5 步：

- 当前是 `HTTP health check`
- 迁移后变成 `checker run`

`AWDRoundUpdater` 继续作为统一入口，不新起第二套调度器。

## 11. API 影响

### 11.1 赛事题目管理

给管理端 `contest_challenges` 相关接口增加 AWD service 配置字段：

- `awd_checker_type`
- `awd_checker_config`
- `awd_sla_score`
- `awd_defense_score`

这些字段仅在 `contest.mode = awd` 时允许设置。

### 11.2 AWD 轮次摘要

`AWDRoundSummaryResp` 和相关查询需要补充：

- `sla_score`
- `defense_score`
- `attack_score`
- `total_score`
- checker 关键结果摘要

### 11.3 后台 AWD 面板

后台面板保留当前轮次、流量、攻击日志结构，但服务结果区要从“探活视图”升级成“checker 结果视图”：

- 展示 `putflag / getflag / havoc` 成功与否
- 展示状态原因和错误码
- 展示 SLA / attack / defense 三段分

本轮不要求重做整页布局，只要求接口字段和摘要意义更新。

## 12. 迁移策略

### 12.1 数据迁移

第一步：

- 给 `contest_challenges` 增加 AWD checker 配置字段
- 给 `awd_team_services` 增加 `sla_score`、`checker_type`

第二步：

- 旧 AWD 赛事若没有显式配置 checker，则使用默认 `http_standard`
- 默认 `awd_checker_config` 从现有全局 `CheckerHealthPath` 推导出最小 `getflag/havoc` 占位配置

### 12.2 运行兼容

为了避免一次切换导致现有赛事无法运行：

- 若某 AWD 题目未配置 `awd_checker_type`，仍允许回退到当前简化探活逻辑
- 但新建或更新 AWD service 时，管理端应优先要求配置 checker

这意味着迁移期会同时存在两类 service：

- `legacy_probe`
- `http_standard`

迁移完成后再考虑移除 `legacy_probe`。

### 12.3 前端兼容

本轮只做最小必要同步：

- 让后台能配置 AWD service checker
- 让 AWD 面板能看懂新的 service 结果结构

学员侧完整比赛页不是本轮必要项。

## 13. 测试策略

至少覆盖下面几类测试：

1. 数据与配置
   - AWD 赛事题目可保存 checker 配置
   - 非 AWD 赛事拒绝保存 AWD checker 配置
2. 轮次执行
   - `http_standard` checker 成功时写入 `up + sla_score`
   - `putflag` 或 `getflag` 失败时写入 `down` 或 `compromised`
3. 攻击影响
   - 成功攻击后，目标 service 的本轮 `defense_score` 被清空
   - 同一攻击方对同一目标同一题在同轮重复提交不重复计分
4. 排行榜
   - 总分由 `sla + defense + attack` 组成
   - 仅官方来源的 checker 与提交结果计入总分
5. 查询与展示
   - AWD 轮次摘要返回三段分和 checker 摘要
   - 后台面板可读取新字段

## 14. 与毕业设计要求的对应

课题里的核心要求与本设计的对应关系如下：

- “使用 Docker 技术快速生成靶机”
  - 继续复用现有 runtime / instance / Docker 机制
- “支持团队对抗模式”
  - 继续使用 `contest(mode=awd) + team`
- “自动监控攻击流量”
  - 保留并继续使用 `awd_traffic_events + runtime proxy`
- “统计防守成功次数，生成实时排行榜”
  - 通过 `awd_team_services + score recalculation` 实现
- “导出实训报告供教学复盘”
  - 现有攻击日志、流量事件和 checker 结果都能作为复盘证据源

也就是说，这次迁移不是偏离毕业设计，而是在保留教学平台边界的前提下，把 AWD 内核做实。

## 15. 当前落地结果

本设计采用以下具体落地原则：

- 不新起独立 A/D 引擎服务
- `contest_challenges` 承担 AWD service 配置
- `awd_team_services` 承担每轮 checker 结果
- 第一版只实现 `http_standard` checker
- 排行榜升级为 `sla / attack / defense / total`
- 保留现有 flag 轮换、攻击提交流程、流量监控和后台面板框架
- 用兼容迁移方式逐步替换旧的 probe-only 逻辑
