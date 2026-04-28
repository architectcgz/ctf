# AWD Checker 试跑设计

## 目标

在现有后台 `ContestManage -> AWD 运维视图 -> 题目配置` 的配置对话框内，补一条真实可执行的 checker 试跑链路。

本轮要解决的问题：

- 管理员在保存前可以直接验证 `legacy_probe / http_standard` 配置是否能打到目标服务
- 试跑结果复用现有 checker 结果语义，而不是前端再拼一套私有状态文案
- 试跑不写 `awd_team_services`，不影响轮次、排行榜和当前服务缓存
- UI 继续留在现有 AWD 题目配置对话框内，保持管理端工作台风格一致

## 非目标

- 不新增独立的 checker 调试页面
- 不把试跑结果写入正式轮次记录
- 不要求 preview 必须绑定真实队伍实例
- 不引入外部 checker 沙箱、异步任务队列或批量试跑

## 当前背景

phase1-5 已经完成：

- AWD 赛事题目支持 `awd_checker_type / awd_checker_config / awd_sla_score / awd_defense_score`
- 后端已有 `legacy_probe / http_standard` checker 执行链
- 后台已有结构化 checker 配置编辑器和 JSON 预览

但当前仍有一个明显缺口：

- 配置只能“填完再保存”，不能在保存前确认目标服务是否真的符合 checker 契约

这会直接削弱毕业设计里“集中监控与管理”的可操作性。管理员即使填完了配置，也只能等到正式轮次巡检后才能知道规则是否可用，排错链路太长。

## 方案比较

### 方案 A：只做前端静态校验

做法：

- 继续只校验字段是否为空、JSON 是否可解析
- 不发真实请求

优点：

- 改动最小
- 没有额外后端接口

缺点：

- 只能说明“字段格式对”，不能说明“checker 真能跑”
- 对 `http_standard` 这种依赖路径、状态码和 flag 回读的配置价值很低

不采用。

### 方案 B：管理员输入访问地址，后端用 preview 上下文真实试跑

做法：

- 在题目配置对话框里补 `access_url` 和可选 `preview_flag`
- 后端复用现有 checker runner，按 preview 上下文直接请求目标服务
- 返回结构化结果给前端展示，但不写库

优点：

- 不依赖赛事里是否已经存在真实实例
- 可以覆盖“创建新题目关联”与“编辑既有题目配置”两个场景
- 复用现有 checker 语义，后续维护成本低

缺点：

- 需要定义 preview 场景下模板变量的占位策略

推荐采用。

### 方案 C：只允许选择真实队伍实例试跑

做法：

- 试跑前先让管理员选择 `team + challenge` 对应实例
- 完全复用真实实例上下文

优点：

- 上下文最贴近正式轮次

缺点：

- 创建题目关联时不可用
- UI 和数据准备成本都更高

本轮不采用。

## 决策

采用方案 B。

这轮在 `AWDChallengeConfigDialog` 内增加“试跑 Checker”能力，管理员填写目标访问地址后，后端用 preview 上下文执行一次真实 checker，并把结构化结果返回给前端展示。

## Preview 上下文设计

### 输入

试跑请求需要这些字段：

- `challenge_id`
- `checker_type`
- `checker_config`
- `access_url`
- `preview_flag`，可选

其中：

- `challenge_id` 只用于给模板变量提供最小上下文，也用于校验题目存在
- `checker_type / checker_config` 先走与保存时一致的合法化逻辑
- `access_url` 是这次试跑唯一目标

### 模板变量占位策略

preview 场景不绑定真实轮次和真实队伍，统一使用下面的占位值：

- `FLAG = preview_flag`，未填写时默认 `flag{preview}`
- `ROUND = 0`
- `TEAM_ID = 0`
- `CHALLENGE_ID = 当前请求中的 challenge_id`

这样可以满足 `http_standard` 当前已支持的模板变量，又不会和正式轮次 flag 混淆。

### 副作用约束

试跑必须满足：

- 不调用轮次同步与写库逻辑
- 不更新 `awd_team_services`
- 不刷新 live service status cache
- 不重算队伍分数

这次执行只是返回一次临时结果。

## 后端设计

### 接口

新增后台接口：

- `POST /api/v1/admin/contests/:id/awd/checker-preview`

请求体：

- `challenge_id`
- `checker_type`
- `checker_config`
- `access_url`
- `preview_flag`

响应体：

- `checker_type`
- `service_status`
- `check_result`
- `preview_context`

其中 `preview_context` 返回：

- `access_url`
- `preview_flag`
- `round_number`
- `team_id`
- `challenge_id`

### 服务层

在 `AWDService` 上新增 preview 命令：

1. 校验赛事必须是 AWD
2. 校验题目存在
3. 复用 `validateAndNormalizeContestAWDFields`
4. 调用 `roundManager` 的 preview 执行方法
5. 把结果包装成 DTO 返回

### Checker 执行层

在 `AWDRoundUpdater` 增加一个“不写库”的 preview 入口。

执行策略：

- `legacy_probe`
  - 构造单个临时 `AWDServiceInstance{AccessURL: access_url}`
  - 走现有 probe 检查逻辑
- `http_standard`
  - 构造单个临时 target
  - 使用 preview 模板上下文执行 `put_flag / get_flag / havoc`
- `check_source` 固定写 `checker_preview`

这样可以最大化复用 phase2 的 checker runner，而不是复制一套 preview 专用逻辑。

## 前端交互设计

### 配置区新增试跑输入

在现有 checker 配置区后面增加一个试跑分区，包含：

- `目标访问地址`
- `预览 Flag`
- `试跑 Checker` 按钮

文案保持工具型表达：

- 说明只强调“会真实请求目标地址，不写入赛事数据”
- 不渲染实现说明和开发者提示

### 结果区

试跑成功后，在对话框内展示：

- 顶部状态摘要：`正常 / 下线 / 已失陷`
- checker 类型、来源、状态原因、执行时间
- 动作结果：`PUT Flag / GET Flag / Havoc`
- 目标探测摘要和原始 JSON 结果预览

试跑失败时：

- 保留最近一次结果区
- 把本次接口错误展示成独立错误提示

### UI 一致性约束

这轮继续遵守现有 CTF 管理端工作台语言：

- 使用现有暗色 surface 和扁平分区
- 主动作仍是高亮按钮，辅助动作保持中性
- 结果区优先用摘要条 + 明细列表，不堆厚卡片
- 继续复用现有 checker 结果标签语义

## 测试策略

至少覆盖：

1. 后端命令
   - `http_standard` preview 成功
   - 无效 checker 配置被拒绝
   - preview 不写 `awd_team_services`
2. 路由与 API 契约
   - 新接口路径和请求体转换正确
   - 前端能归一化 preview 响应
3. 前端对话框
   - 填写访问地址后可以触发试跑
   - 返回结果后能展示状态摘要与动作明细
   - 失败时能显示错误信息且不影响保存流程
