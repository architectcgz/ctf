# AWD Phase 10 学员实战工作台设计

在现有 AWD 引擎 `phase2-9` 的基础上，把学生参赛页从通用 Jeopardy 视图升级为真正可用于 A/D 对抗的工作台，让学员能在同一页内完成服务启动、目标查看、偷旗提交和实时分数反馈。

## 非目标

- 不重做学员侧整套路由结构
- 不新增独立的学生 AWD 大屏或旁路工作台
- 不开放管理员级流量明细、全量服务状态矩阵或 readiness 运维动作给学生
- 不改现有 AWD 计分、checker、轮次调度和教师复盘语义
- 不做复杂反作弊、批量攻击脚本托管或自动化 exploit 编排

## 背景

phase2-8 已经完成了 AWD 引擎的核心运行能力：

- 标准 `http_standard` checker、`sla / attack / defense / total` 三段分
- AWD 题目配置、结构化编辑、preview、validation state、readiness gate
- 学生攻击提交接口 `POST /api/v1/contests/:id/awd/challenges/:cid/submissions`
- 竞赛题目实例启动接口 `POST /api/v1/contests/:id/challenges/:cid/instances`
- 公开排行榜、公告实时刷新、队伍管理

phase9 又补齐了教师侧复盘、归档和报告导出。

但学生侧 `/contests/:id` 仍然是通用竞赛详情页：

- 题目页仍按 Jeopardy 的“选题 -> 提 Flag”工作流组织
- 没有自己的服务入口
- 没有目标队伍目录
- 没有 AWD 攻击提交表单
- 没有围绕当前轮次的即时反馈区

这意味着后端已经具备“学生能打”的能力，前端却还没有把这条链路组织成可用工作台。毕业设计里的“学员选择靶场后获取攻击目标，系统自动记录攻击步骤与漏洞利用过程，实时反馈得分”在 AWD 场景下还缺最后一段学生操作面。

## 方案比较

### 方案 A：继续复用当前题目页，只把提交按钮切到 AWD 接口

做法：

- 保留当前 `ContestDetail` 的题目列表和 Flag 输入框
- 如果 `contest.mode = awd`，只把提交目标换成 AWD 提交接口

优点：

- 改动最小
- 上线最快

缺点：

- 没有自己的服务入口与目标目录
- 学生仍然看不到“我要打谁、打哪个地址”
- 只是换接口，不是 A/D 工作台

不推荐。

### 方案 B：在现有竞赛详情页内嵌 AWD 战场面板

做法：

- 保留 `/contests/:id` 路由和现有 shell
- 当 `contest.mode = awd` 时，把“题目”页签升级为“战场”页签
- 新增学生侧 AWD workspace 查询接口，返回当前轮次、我的服务状态、目标目录和最近攻防反馈
- 前端在同一页里串起服务启动、目标查看、攻击提交、排行榜刷新

优点：

- 复用现有 contest 详情路由、公告、队伍和榜单能力
- 对学生来说路径最短，不需要再学一套新导航
- 风格可以继续对齐现有 student journal shell

缺点：

- 需要给 contest 模块补一条用户态 AWD read model
- `ContestDetail.vue` 需要做模式分流

推荐采用。

### 方案 C：新建独立学生 AWD 页面

做法：

- 新增 `/contests/:id/awd` 或 `/academy/awd/:id` 独立页面
- 把学生攻防动作全部迁到新页面

优点：

- 页面职责最清晰
- 后续能做更激进的布局

缺点：

- 路由、导航和用户心智都会被拆开
- 需要额外处理旧详情页跳转和状态同步
- 对当前阶段来说改动过大

不采用。

## 目标

本轮完成后，学生在 AWD 竞赛详情页内应能完成这条连续工作流：

1. 查看当前轮次和本队是否已进入战场
2. 启动或打开本队某题目的比赛实例
3. 浏览其他队伍该题目的攻击目标地址
4. 对指定队伍提交 stolen flag
5. 在同页看到成功/失败反馈、最近攻防记录和实时排行榜变化
6. 在运行中比赛里持续获取新的轮次快照，并快速识别本队异常服务

## 总体设计

采用方案 B，在现有 `ContestDetail` 内按模式切换工作台：

- `overview / announcements / team` 保持不变
- 原 `challenges` 页签在 AWD 赛事下改为 `battle`
- `battle` 页签由新的 `ContestAWDWorkspacePanel` 承接
- 后端新增 `GET /api/v1/contests/:id/awd/workspace`

### 统一事实源

学生工作台只依赖三类事实源：

- 已有公开竞赛元信息与题目列表
- 新增的用户态 AWD workspace read model
- 已有公开排行榜接口与 scoreboard realtime 事件

这样可以避免把“学生当前战场状态”拆散在多个页面专用接口里。

## 后端设计

### 新接口

新增：

- `GET /api/v1/contests/:id/awd/workspace`

返回：

- 当前运行轮次
- 当前用户所属队伍
- 当前用户所在队伍的服务状态与已有访问地址
- 其他队伍的目标目录（只暴露题目对应访问地址，不暴露管理端运维字段）
- 当前轮与本队相关的最近攻防事件

### 数据来源

复用现有 `contestports.AWDRepository` 能力：

- `FindContestTeamByMember`
- `ListChallengesByContest`
- `ListServiceInstancesByContest`
- `FindRunningRound`
- `ListServicesByRound`
- `ListAttackLogsByRound`
- `FindTeamsByContest`

不新建专门的学生 AWD 表。

### 权限边界

- 只有登录用户可访问
- 非 AWD 赛事直接拒绝
- 未加入队伍时，返回当前轮次和空的目标/服务数据，不暴露他队目标地址
- 已加入队伍后，才返回目标目录和本队相关事件

### DTO 设计

新增用户态 DTO：

- `ContestAWDWorkspaceResp`
- `ContestAWDWorkspaceTeamResp`
- `ContestAWDWorkspaceServiceResp`
- `ContestAWDWorkspaceTargetTeamResp`
- `ContestAWDWorkspaceTargetServiceResp`
- `ContestAWDWorkspaceRecentEventResp`

其中：

- `services` 只表示“我的队伍”
- `targets` 只表示“其他队伍”
- `targets.services` 只返回 `reachable` 等代理可达状态，不向学生端暴露对方实例的原始 `access_url`
- `services` 可返回本队运行实例的 `instance_id`，学生打开本队服务时也通过 `/instances/:id/access` 获取平台代理地址，不直接展示容器内网 `access_url`
- `recent_events` 只保留与当前队伍有关的当前轮事件，并区分 `attack_out / attack_in`

## 前端设计

### 页面结构

`ContestDetail.vue` 在 `contest.mode === 'awd'` 时：

- 页签文案从“题目”切到“战场”
- 内容区域渲染 `ContestAWDWorkspacePanel`

`ContestAWDWorkspacePanel` 采用一个主工作区 + 一个右侧上下文栏：

- 主区：
  - 防守告警摘要
  - 我的服务目录
  - 目标队伍目录
  - stolen flag 提交区
- 右栏：
  - 当前轮次摘要
  - 自动刷新状态
  - 实时排行榜 Top 区
  - 最近攻防反馈

### 交互规则

- 若当前用户尚未加入队伍：
  - 主区不展示目标目录与攻击表单
  - 显示“先在队伍页创建或加入队伍”的明确提示
- 启动服务时：
  - 调用现有 `POST /contests/:id/challenges/:cid/instances`
  - 成功后刷新 workspace
- 提交 stolen flag 时：
  - 调用现有 `POST /contests/:id/awd/challenges/:cid/submissions`
  - 成功后刷新 workspace 与排行榜
- 比赛处于 `running / frozen` 时：
  - 工作台每 15 秒自动刷新一次 workspace 与排行榜
  - 面板显示最近一次同步时间，避免学生误判数据是否已更新
- 我的服务存在 `down / compromised / attack_received > 0` 时：
  - 在“我的服务”前方给出防守告警摘要
  - 直接点名异常题目，减少学生在列表里逐条排查
- 目标目录提供队伍关键字筛选，并支持只看当前有可用地址的目标
- 排行榜继续复用现有 `getScoreboard` + websocket 刷新

### UI 约束

- 继续使用现有 `journal-shell-user`、`workspace-page-title`、`metric-panel-*`、`top-tabs`
- 不给学生侧引入管理员运维色板或 teacher dark surface
- 不把解释实现的说明性文字写进 UI

## 测试策略

至少覆盖：

1. 后端 query
   - 已入队学生能拿到当前轮次、我的服务、目标目录和相关事件
   - 未入队学生不会看到他队目标地址
2. 路由
   - 新用户态 workspace 路由已注册到 contest AWD handler
3. 前端 API
   - workspace 查询、比赛实例启动、AWD 攻击提交的请求与返回归一化
4. 前端页面
   - AWD 赛事显示“战场”页签而不是通用题目页
   - 启动服务、提交 stolen flag、刷新最近反馈与排行榜
   - 运行中的战场每 15 秒自动刷新
   - 异常服务会进入防守告警摘要
   - 目标目录支持队伍筛选与可用地址过滤
   - 未入队时显示受限空态

## 与毕业设计要求的对应

- “学员选择靶场后获取攻击目标”
  - 学生工作台直接展示按题目分组的目标队伍访问地址
- “系统自动记录攻击步骤与漏洞利用过程”
  - stolen flag 提交继续写 `awd_attack_logs`，运行代理继续写 `awd_traffic_events`
- “实时反馈得分”
  - 提交返回即时分数变化，排行榜区继续实时刷新
- “支持团队对抗模式”
  - 工作台按队伍视角组织本队服务和他队目标，不再沿用个人 Jeopardy 提交流程

## 结论

phase10 不再继续补管理员或教师视角，而是把学生真实参赛链路接上。这样 AWD 迁移才算从“引擎可跑、后台可管、教师可复盘”走到“学生可直接参赛”。
