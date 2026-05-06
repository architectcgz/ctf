# AWD 学员实战工作台设计

在现有 AWD 引擎 `phase2-9` 的基础上，把学生参赛页从通用 Jeopardy 视图升级为真正可用于 A/D 对抗的工作台，让学员能在同一页内完成服务启动、目标查看、偷旗提交和实时分数反馈。

## 非目标

- 不重做学员侧整套路由结构
- 不新增独立的学生 AWD 大屏或旁路工作台
- 不开放管理员级流量明细、全量服务状态矩阵或 readiness 运维动作给学生
- 不改现有 AWD 计分、checker、轮次调度和教师复盘语义
- 不做复杂反作弊、批量攻击脚本托管或自动化 exploit 编排
- 不提供浏览器文件树、在线文件编辑或平台圈定修补文件列表

## 背景

phase2-8 已经完成了 AWD 引擎的核心运行能力：

- 标准 `http_standard` checker、`sla / attack / defense / total` 三段分
- AWD 题目配置、结构化编辑、preview、validation state、readiness gate
- 学生攻击提交接口 `POST /api/v1/contests/:id/awd/services/:sid/submissions`
- AWD 服务实例启动接口 `POST /api/v1/contests/:id/awd/services/:sid/instances`
- 跨队目标代理入口 `POST /api/v1/contests/:id/awd/services/:sid/targets/:team_id/access`
- 本队防守 SSH 入口 `POST /api/v1/contests/:id/awd/services/:sid/defense/ssh`
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
7. 为本队服务生成防守 SSH 连接信息，便于学生进入独立 defense workspace 自行侦察和修补

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
- 当前用户所在队伍各服务的 defense workspace 入口状态
- 其他队伍的目标目录（只暴露题目对应访问地址，不暴露管理端运维字段）
- 当前轮与本队相关的最近攻防事件

### 防守重启与攻击访问反馈

学生工作台里的“重启本队服务”表示防守方在 defense workspace 完成修补后，请求平台重新拉起本队对应 service 的比赛容器。它不是“打开实例”或“复用已有实例”，也不表示服务已被攻破。

后端应提供显式重启动作：

- `POST /api/v1/contests/:id/awd/services/:sid/instances/restart`

重启动作只允许当前登录用户所在队伍操作本队 service，不能操作其他队伍实例。重启时保留实例身份和比赛归因字段：

- `contest_id`
- `team_id`
- `service_id`
- `challenge_id`
- `share_scope`
- `nonce`
- `host_port`
- `expires_at`

重启时清空运行态字段并重新进入调度：

- 清理旧容器、网络和运行态 ACL
- 清空 `container_id`
- 清空 `network_id`
- 清空 `runtime_details`
- 清空 `access_url`
- 将实例状态置为 `pending`

调度器随后按现有 pending 实例流程重建容器。这样可以复用已有实例身份、端口和动态 flag nonce，避免一次防守重启被误建成新的“队伍 + service”运行实体。

攻击代理必须区分“目标应用真实返回 500”和“平台知道目标暂不可用”：

- 目标实例 `running` 且可反代：原样透传目标应用状态码，包括目标应用自己的 `500`
- 目标实例存在但处于 `pending / creating`：返回 `503 Service Unavailable`，提示目标服务正在启动或重启
- 目标实例存在但处于 `failed`：返回 `503 Service Unavailable`，提示目标服务暂不可用
- 目标实例不存在或访问方无权限：继续按权限和资源语义返回 `403 / 404`

攻击流量记录仍记录实际返回给攻击方的状态码。重启窗口内的受控不可用记录为 `503`，不记录为平台 `500`。

本阶段不做重启免罚窗口。checker 仍按真实探测结果判定服务状态：

- 重启期间 checker 失败或超时：记为 `down`
- flag 内容不符合预期：记为 `compromised`
- 防守方点击重启本身：不直接产生 `compromised`

不做免罚窗口的原因是避免队伍通过频繁重启规避攻击或规避 checker。若后续确实需要重启宽限，应单独设计频率限制、最大宽限秒数和审计记录，不能隐式放进本次重启语义。

这里还要明确 defense workspace 与 restart 的关系：

- `restart` 只重建运行容器，不清空 defense workspace
- 学生已改动的业务文件在普通 restart 后继续保留
- 只有管理员触发的 `recreate / reseed` 才会重置 defense workspace 并使旧连接失效

### 防守连接入口语义

`POST /api/v1/contests/:id/awd/services/:sid/defense/ssh` 的语义应固定为“签发 defense workspace 连接票据”，而不是“进入 service 容器 shell”。

约束如下：

- 连接目标是本队该服务对应的 `defense workspace container`
- 返回内容只包含连接所需主机、端口、用户名、短时票据和命令示例
- 不返回文件树、文件路径、`editable_paths`、`protected_paths` 或 `service_contracts`
- 若 workspace 暂未就绪，应返回受控不可用状态，而不是退回 service 容器
- battle 页可以展示连接命令与使用提示，但不再跳转浏览器防守文件页

### 数据来源

复用现有 `contestports.AWDRepository` 聚合比赛事实，并额外读取队伍服务级 `defense workspace` 状态：

- `FindContestTeamByMember`
- `FindRunningRound`
- `FindTeamsByContest`
- `ListChallengesByContest`
- `ListServiceDefinitionsByContest`
- `ListServiceInstancesByContest`
- `ListDefenseWorkspaceSummariesByContestTeam`
- `ListServicesByRound`
- `ListAttackLogsByRound`

`GET /api/v1/contests/:id/awd/workspace` 不新建学生专用 read model 表；防守工作区状态直接复用 `awd_defense_workspaces` 持久化表，并在 query 阶段与 contest/service/instance 数据聚合。

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
- `services` 不再返回 `defense_scope`，只通过 `defense_connection` 暴露防守入口摘要
- `services.defense_connection` 只返回 `entry_mode`、`workspace_status`、`workspace_revision` 三个字段，用于判断 SSH 入口是否可用、当前 workspace 是否就绪，以及 `reseed / recreate` 后连接是否已轮换；不下发 `container_id`、挂载根或文件路径
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
  - 调用现有 `POST /contests/:id/awd/services/:sid/instances`
  - 成功后刷新 workspace
- 提交 stolen flag 时：
  - 调用现有 `POST /contests/:id/awd/services/:sid/submissions`
  - 成功后刷新 workspace 与排行榜
- 比赛处于 `running / frozen` 时：
  - 工作台每 15 秒自动刷新一次 workspace 与排行榜
  - 面板显示最近一次同步时间，避免学生误判数据是否已更新
- 我的服务存在 `down / compromised / attack_received > 0` 时：
  - 在“我的服务”前方给出防守告警摘要
  - 直接点名异常题目，减少学生在列表里逐条排查
- 目标目录提供队伍关键字筛选，并支持只看当前有可用地址的目标
- 打开攻击目标时：
  - 调用 `POST /contests/:id/awd/services/:sid/targets/:team_id/access`
  - 成功后打开平台代理地址，不直接暴露目标队伍实例原始地址
- 生成防守 SSH 时：
  - 调用 `POST /contests/:id/awd/services/:sid/defense/ssh`
  - 前端临时展示连接命令和 OpenSSH 配置，并提供复制入口
  - 不再跳转浏览器文件工作台
- 防守方点击 restart 后：
  - 主区继续保留 defense workspace 入口
  - 只提示“服务正在重启”，不把工作区视为被清空
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
   - workspace 查询、比赛实例启动、AWD 攻击提交、defense SSH 的请求与返回归一化
4. 前端页面
   - AWD 赛事显示“战场”页签而不是通用题目页
   - 启动服务、提交 stolen flag、刷新最近反馈与排行榜
   - 运行中的战场每 15 秒自动刷新
   - 异常服务会进入防守告警摘要
   - 目标目录支持队伍筛选与可用地址过滤
   - 学生侧不再出现浏览器防守文件路由和相关 API 调用
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
