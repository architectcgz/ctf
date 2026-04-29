# AWD 开赛就绪门禁设计

## 目标

在现有后台 `ContestManage -> AWD 运维视图` 这条链路上，把 `phase7` 已经具备的 checker 校验状态升级成真正可执行的“开赛前门禁”。

本轮要解决的问题：

- `pending / failed / stale` 不再只是展示信息，而是接入关键管理动作
- 管理员在创建轮次、执行当前轮巡检、把赛事切到 `running` 前，能先看到一份可操作的就绪摘要
- 系统默认阻止“带病开赛”，但保留一次性、可审计的强制放行能力
- 前后端继续复用现有 AWD 运维页、现有命令接口和现有审计模型，不再另起一套管理流程

## 非目标

- 不改正式轮次巡检与计分语义
- 不把强制放行视为 checker 校验通过
- 不新增独立的 AWD readiness 页面
- 不做自动修复、自动回滚或批量重试 checker
- 不扩展到 Jeopardy 赛事

## 当前背景

phase2-7 已经完成这些能力：

- AWD 题目支持 `legacy_probe / http_standard`
- 管理端支持结构化 checker 配置编辑与 preview 试跑
- `contest_awd_services` 已持久化：
  - `awd_checker_validation_state`
  - `awd_checker_last_preview_at`
  - `awd_checker_last_preview_result`
- AWD service 配置列表已经可以展示：
  - `未验证`
  - `最近通过`
  - `最近失败`
  - `待重新验证`

但现在还有一个明显缺口：

- 当前轮巡检、创建轮次、赛事切到 `running` 仍然不会因为 readiness 不通过而被拦住
- 也没有一份面向管理员的“这场 AWD 现在到底能不能开”的摘要
- `phase7` 的状态更多还是“被动展示”，还没有进入运营闭环

这会让 service 配置能力和赛事运维能力之间继续断开，不符合课题里“集中监控与管理”“自动监控攻击流量、统计防守成功次数”的整体口径。

## 方案比较

### 方案 A：只提示，不拦截

做法：

- 前端新增 readiness 看板
- 关键动作仍然允许直接执行

优点：

- 改动最小
- 对现有命令链影响最小

缺点：

- `phase7` 仍然停留在展示层
- 无法真正阻止“带病开赛”
- 管理员依旧可能忽略提示直接执行关键动作

不采用。

### 方案 B：默认门禁，允许带审计的强制放行

做法：

- 对 `创建轮次 / 当前轮巡检 / 赛事切到 running` 三类关键动作增加 readiness gate
- 默认阻止存在阻塞项的执行
- 管理员可以填写原因后强制继续
- 后端把本次放行原因与阻塞快照写入现有 `audit_logs`

优点：

- 真正把 checker 校验状态接入运营动作
- 兼顾安全和现场运维灵活性
- 不需要新增持久化表

缺点：

- 要补一层共享 gate 服务
- 前端要补“阻塞确认 + 原因输入”交互

推荐采用。

### 方案 C：全硬门禁，不允许放行

做法：

- readiness 不通过时，关键动作永远不允许执行

优点：

- 规则最简单
- 一致性最强

缺点：

- 只要一题状态异常，就会把整场 AWD 运维完全卡死
- 不适合教学、演示和现场应急操作

本轮不采用。

## 决策

采用方案 B。

phase8 把 `phase7` 已有的 validation state 接入关键管理动作，形成“先汇总 readiness，再决定是否执行”的闭环：

1. 系统汇总当前 AWD 赛事的 readiness 摘要
2. 关键动作执行前先走 readiness gate
3. 若存在阻塞项，默认返回阻止结果
4. 管理员可填写原因后强制执行一次
5. 强制执行会写入审计，但不会改变题目本身的 validation state

## 门禁范围

### 拦截动作

本轮拦截以下动作：

- `创建轮次`
- `立即巡检当前轮`
- 赛事状态切到 `running`
- 定时状态任务自动把 AWD 赛事从 `registration` 推进到 `running`

原因：

- 这些动作直接决定 AWD 是否进入正式运行态
- 它们都属于“开赛前 / 开赛中关键操作”
- 它们已经有明确的后端命令入口，适合补统一 gate
- 自动状态推进没有管理员交互上下文，不能走强制放行；readiness 不通过或检查失败时应继续停留在 `registration`

### 不拦截动作

- 指定历史轮次重跑
- 人工补录服务检查
- 人工补录攻击日志
- 题目配置编辑与 checker preview

原因：

- 这些动作本身就是排障或运维手段
- 如果也被 gate 拦住，会妨碍管理员修复 readiness 问题

## readiness 判定规则

### 通过条件

AWD 赛事下，每个已关联 service 满足以下条件才视为通过：

- 已配置 `checker_type`
- `awd_checker_validation_state = passed`

### 阻塞条件

以下任一情况都视为阻塞：

- 未配置 checker
- 已配置 checker，但当前保存的 `checker_config` 缺少该 checker 的必需字段，或无法被后端正常解析
- `awd_checker_validation_state = pending`
- `awd_checker_validation_state = failed`
- `awd_checker_validation_state = stale`

### 边界约定

- 没有关联题目的 AWD 赛事视为 readiness 不通过
- readiness 只根据当前保存态计算，不读取对话框临时草稿
- 强制放行不会把 `failed / stale / pending` 改写成 `passed`
- phase8 不新增 `invalid_config` 独立统计桶
- 对于“已配置 checker，但保存态配置缺失或无法解析”的脏数据，统一并入 `missing_checker_challenges`

## 数据设计

### readiness 摘要 DTO

新增一份只读摘要结构，至少包含：

- `contest_id`
- `ready`
- `total_challenges`
- `passed_challenges`
- `pending_challenges`
- `failed_challenges`
- `stale_challenges`
- `missing_checker_challenges`
- `blocking_count`
- `blocking_actions`
- `global_blocking_reasons`
- `items`

其中 `blocking_actions` 是一个字符串数组，用于告诉前端这份 readiness 会阻塞哪些动作，例如：

- `create_round`
- `run_current_round_check`
- `start_contest`

`global_blocking_reasons` 用于承载不属于单个 service 明细的系统级阻塞原因。本轮至少约定：

- `no_challenges`

零 service 场景下的固定返回口径如下；当前接口字段和原因码仍保留 `challenges/no_challenges` 命名以兼容前端契约：

- `ready = false`
- `total_challenges = 0`
- `blocking_count = 1`
- `global_blocking_reasons = ["no_challenges"]`
- `items = []`
- `blocking_actions = ["create_round", "run_current_round_check", "start_contest"]`

### readiness 明细项

每个阻塞项至少返回：

- `service_id`
- `challenge_id`
- `title`
- `checker_type`
- `validation_state`
- `last_preview_at`
- `last_access_url`
- `blocking_reason`

补充约定：

- `last_access_url` 是可选字段
- 若最近一次 preview 快照里存在 `preview_context.access_url`，则按该值返回
- 若历史快照没有这个字段，或结果无法安全解析，则返回空值，前端按“无目标地址”降级展示
- `blocking_reason` 至少支持：
  - `missing_checker`
  - `invalid_checker_config`
  - `pending_validation`
  - `last_preview_failed`
  - `validation_stale`

## 后端接口设计

### 只读 readiness 接口

新增：

- `GET /api/v1/admin/contests/:id/awd/readiness`

职责：

- 汇总当前 AWD 赛事 readiness
- 给前端运维页和赛事状态流转预检查复用

### 创建轮次请求补充

继续复用现有创建轮次接口，只新增可选字段：

- `force_override`
- `override_reason`

### 当前轮巡检请求补充

继续复用现有巡检接口，但把原来的“空 POST”扩成可选 JSON body：

- `force_override`
- `override_reason`

指定历史轮次重跑不走 gate，也不需要这两个字段。

### 更新赛事状态补充

继续复用 `UpdateContest`。

只有在下面条件同时满足时，才触发 readiness gate：

- 当前赛事模式为 `awd`
- 请求里有 `status`
- 目标状态为 `running`

请求同样新增可选字段：

- `force_override`
- `override_reason`

### `override_reason` 约束

- 只在 `force_override = true` 时生效
- 后端按 `trim` 后结果做校验
- `trim` 后为空时视为非法请求，不进入放行
- 长度上限本轮定为 500 字符
- 未强制时即使传入该字段，也不改变普通 gate 行为

## 后端服务设计

### 共享 readiness query

新增一层共享 readiness 查询服务，供下面场景复用：

- AWD 运维页摘要展示
- 创建轮次 gate
- 当前轮巡检 gate
- 赛事状态切到 `running` 的 gate

这层服务只依赖当前已保存的 `contest_awd_services` 数据，不依赖临时草稿。

### 共享 readiness gate

新增一个共享 gate helper，输入：

- `contest_id`
- `action`
- `force_override`
- `override_reason`

输出：

- readiness 摘要
- 是否允许继续执行

行为规则：

- 无阻塞项：直接放行
- 有阻塞项且未强制：返回专用 gate 冲突错误
- 有阻塞项且强制：
  - 校验 `override_reason` 在 trim 后非空且长度合法
  - 返回本次放行使用的阻塞快照
  - 允许本次命令继续

共享 gate helper 只负责给出放行决策与阻塞快照，不直接写“执行成功”审计。强制放行相关审计由实际命令调用方在业务动作返回后写入，这样才能带上最终执行结果。

### 错误语义

本轮不单独新增“巨大的错误详情协议”。

推荐保持：

- 前端优先主动请求 readiness
- 后端命令仍做二次校验
- 若被 gate 拦截，返回专用 `ErrAWDReadinessBlocked`，HTTP 状态仍为 `409 conflict`
- 前端根据响应 `Envelope.code` 区分“readiness 门禁拦截”和其他普通冲突
- `409` 响应不内联完整 readiness 详情，前端收到专用错误码后再主动重拉 readiness

这样可以避免把当前错误处理链改得过重，同时保留前端稳定分流能力。

## 审计设计

### 复用现有审计模型

直接复用现有：

- `internal/auditlog.Entry`
- `audit_logs`

不新增审计表。

### 审计内容

强制放行时记录：

- `action = admin_op`
- `resource_type = contest`
- `resource_id = contest_id`
- `detail`

`detail` 至少包含：

- `module = awd_readiness_gate`
- `gate_action`
- `override_reason`
- `blocking_count`
- `global_blocking_reasons`
- `blocking_items`
- `execution_outcome`
- `execution_error`

补充约定：

- 审计只在 `force_override = true` 且 gate 允许放行时写入
- 审计写入时机放在实际业务动作返回之后
- `execution_outcome` 至少区分：
  - `succeeded`
  - `failed`
- 当后续业务动作失败时，仍然写这条审计，并在 `execution_error` 中保留精简错误摘要
- `blocking_items` 可以只保留精简快照，不必完整复制 readiness 全响应
- 零题目场景下 `blocking_items = []`，阻塞原因通过 `global_blocking_reasons = ["no_challenges"]` 表达

## 前端交互设计

### AWD 运维页公共 readiness 区块

不新增独立页面，直接放在 `AWDOperationsPanel` 的赛事选择与 tab rail 之间。

区块包含两层：

1. 概况卡片
   - 已通过
   - 未验证
   - 最近失败
   - 待重新验证
   - 未配 Checker
2. 阻塞题目短名单
   - 题目名
   - 当前状态
   - 最近校验时间
   - 目标地址
   - `编辑配置` 入口

如果存在 `global_blocking_reasons`，则在短名单上方显示系统级阻塞摘要；例如零题目时显示“当前赛事还没有关联题目，无法执行开赛关键动作”。

视觉继续沿用现有 `metric-panel`、`workspace-directory-section` 和 flat row 语言。

### 被拦动作的交互

对于：

- 创建轮次
- 立即巡检当前轮
- AWD 赛事状态切到 `running`

如果 readiness 不通过：

- 不只弹一个普通 toast
- 而是打开一个阻塞确认层

确认层包含：

- 当前动作名称
- 阻塞摘要
- 系统级阻塞原因（若存在）
- 阻塞 service 列表
- 原因输入框
- `取消`
- `强制继续`

### 指定轮次重跑

不经过 readiness gate，继续保持现状。

原因：

- 这是排障手段，不应该被 readiness 阻止

## 文案与一致性约束

- UI 文案继续保持工具型表达
- 不在页面渲染“方案说明”“实现解释”等说明性文字
- `强制继续` 的提示语要明确：
  - 本次只是跳过门禁
  - 不代表 checker 已验证通过

## 测试策略

### 后端

至少覆盖：

1. readiness 摘要查询
   - 统计 `passed / pending / failed / stale / missing_checker`
   - 零题目赛事返回 `global_blocking_reasons = ["no_challenges"]`
   - 已配置 checker 但配置无法解析时，并入 `missing_checker`
2. 创建轮次 gate
   - 默认拦截
   - 强制放行可通过
   - `override_reason` 纯空白会被拒绝
3. 当前轮巡检 gate
   - 默认拦截
   - 强制放行可通过
4. `UpdateContest -> running`
   - AWD 赛事会走 gate
   - 非 AWD 赛事不受影响
5. 审计
   - 强制放行成功会写 `audit_logs`
   - 强制放行后业务动作失败也会写 `audit_logs`

### 前端

至少覆盖：

1. AWD 运维页渲染 readiness 摘要
   - 零 service 时显示系统级阻塞摘要
2. 阻塞 service 列表展示
3. 创建轮次被拦后出现原因输入层
4. 当前轮巡检被拦后出现原因输入层
5. AWD 赛事状态切到 `running` 时触发同样门禁
6. 填写原因后，请求携带：
   - `force_override`
   - `override_reason`
7. 收到专用 gate 错误码后，会自动重拉 readiness，而不是把所有 `409` 都当成门禁拦截

## 验收标准

- AWD 运维页能直接看到当前赛事 readiness 摘要
- 零题目赛事会被明确标记为不可开赛，而不是展示为空白正常态
- 当存在 `pending / failed / stale / missing_checker` 题目时：
  - 创建轮次默认被拦
  - 当前轮巡检默认被拦
  - 赛事切到 `running` 默认被拦
- 管理员填写原因后可以单次强制继续
- 强制继续会写审计
- 强制继续不会改写题目原有 validation state
- 指定历史轮次重跑不受 gate 影响
