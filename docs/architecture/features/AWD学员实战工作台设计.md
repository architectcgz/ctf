# AWD 学员实战工作台架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`frontend`、`backend`、`contracts`
- 关联模块：
  - `code/frontend/src/views/contests`
  - `code/frontend/src/components/contests`
  - `code/frontend/src/features/contest-awd-workspace`
  - `code/frontend/src/api/contest.ts`
  - `code/backend/internal/module/contest/application/queries`
  - `code/backend/internal/app/router_routes.go`
- 过程追溯：旧稿 `AWD学员实战工作台设计.md`
- 最后更新：`2026-05-07`

## 1. 背景与问题

旧稿把学生 AWD 面写成了“如何从通用竞赛页演进出战场工作台”的方案，但当前代码已经给出正式组织方式。这里需要记录的是：

- 学生 AWD 仍然挂在 `/contests/:id`，没有拆成独立路由
- 学生工作台依赖一条专用 AWD read model 和一组动作接口
- 防守入口已经固定为 SSH 连接，而不是浏览器文件工作台

## 2. 架构结论

- 学生 AWD 赛事仍使用现有竞赛详情页 `/contests/:id`。
- `ContestDetail.vue` 在 `contest.mode === 'awd'` 时，把原 `challenges` 面板替换为 `ContestAWDWorkspacePanel`，标题显示为“攻防战场”。
- 学生战场 read model 固定为 `GET /contests/:id/awd/workspace`，返回 `current_round`、`my_team`、`services`、`targets`、`recent_events`。
- 学生动作当前固定为五类：
  - 启动本队服务：`POST /contests/:id/awd/services/:sid/instances`
  - 重启本队服务：`POST /contests/:id/awd/services/:sid/instances/restart`
  - 提交 stolen flag：`POST /contests/:id/awd/services/:sid/submissions`
  - 打开目标访问：`POST /contests/:id/awd/services/:sid/targets/:team_id/access`
  - 生成防守 SSH：`POST /contests/:id/awd/services/:sid/defense/ssh`
- 当前工作台会在赛事 `running / frozen` 时每 15 秒刷新 workspace 和 scoreboard。
- 深度防守入口是 SSH ticket；学生当前没有独立 `/contests/:id/awd` 页面，也没有浏览器文件编辑器。

## 3. 模块边界与职责

### 3.1 模块清单

- `ContestDetail.vue`
  - 负责：按 `contest.mode` 决定题目面板展示普通刷题工作台还是 AWD 战场
  - 不负责：编排战场内部轮询和动作流

- `ContestAWDWorkspacePanel`
  - 负责：渲染战场 HUD、本队服务、目标目录、最近战报、榜单摘要
  - 不负责：定义后端 read model

- `useContestAWDWorkspace`
  - 负责：加载 `awd/workspace` 和 scoreboard、控制自动刷新
  - 不负责：直接处理具体动作请求

- `useAwdWorkspaceServiceActions`
  - 负责：启动和重启本队服务
  - 不负责：攻击提交流程

- `useAwdWorkspaceAttackSubmission`
  - 负责：提交 stolen flag 并刷新战场视图
  - 不负责：打开目标或 SSH

- `useAwdWorkspaceAccessActions`
  - 负责：打开本队服务、打开目标、生成 SSH 连接
  - 不负责：维护战场 read model

- `AWDWorkspaceQuery`
  - 负责：组装学生战场 read model
  - 不负责：暴露管理员运维字段

### 3.2 事实源与所有权

- 页面模式切换事实源：`ContestDetail.vue`
- 战场 read model 事实源：`ContestAWDWorkspaceResp`
- 分数榜事实源：`GET /contests/:id/scoreboard` 与 `ScoreboardRealtimeBridge`
- 学生动作 HTTP 面事实源：`router_routes.go`

## 4. 关键模型与不变量

### 4.1 核心实体

- `ContestAWDWorkspaceResp`
  - 顶层字段：`contest_id`、`current_round`、`my_team`、`services`、`targets`、`recent_events`

- `ContestAWDWorkspaceServiceResp`
  - 关键字段：`instance_id`、`instance_status`、`access_url`、`service_status`、`operation_status`、`checker_type`、`attack_received`、`sla_score`、`defense_score`、`attack_score`、`defense_connection`

- `ContestAWDWorkspaceTargetServiceResp`
  - 关键字段：`service_id`、`awd_challenge_id`、`reachable`

### 4.2 不变量

- AWD 学生面不新增独立路由，仍嵌在 `ContestDetail` 的 `challenges` tab 中。
- `targets` 只暴露目标可达性和服务映射，不暴露管理员运维字段。
- `defense_connection` 只提供 SSH 入口状态，不提供浏览器文件工作台入口。
- 自动刷新只在赛事 `running / frozen` 时开启，间隔固定为 `15_000 ms`。
- 普通 Jeopardy 赛事继续使用 `ContestChallengeWorkspacePanel`，不会共享 AWD 战场 read model。

## 5. 关键链路

### 5.1 页面加载链路

1. 学生进入 `/contests/:id`。
2. `ContestDetail.vue` 根据 `contest.mode` 判断是否启用 AWD 战场。
3. 若为 AWD，`ContestAWDWorkspacePanel` 挂载并调用 `useContestAWDWorkspace`。
4. 前端并行请求 `GET /contests/:id/awd/workspace` 和 `GET /contests/:id/scoreboard`。
5. 若赛事处于 `running / frozen`，前端启动 15 秒轮询；榜单实时事件由 `ScoreboardRealtimeBridge` 补充刷新。

### 5.2 学生动作链路

1. 启动或重启动作通过 `useAwdWorkspaceServiceActions` 调用对应接口。
2. 攻击方通过 `useAwdWorkspaceAccessActions.openTarget` 申请目标访问，再通过 `useAwdWorkspaceAttackSubmission` 提交 stolen flag。
3. 防守方通过 `openDefenseSSH` 申请 SSH 连接信息，进入独立防守工作区。
4. 每次成功动作后，工作台重新刷新 workspace 数据。

## 6. 接口与契约

### 6.1 读取接口

- `GET /contests/:id/awd/workspace`
- `GET /contests/:id/scoreboard`

### 6.2 动作接口

- `POST /contests/:id/awd/services/:sid/instances`
- `POST /contests/:id/awd/services/:sid/instances/restart`
- `POST /contests/:id/awd/services/:sid/submissions`
- `POST /contests/:id/awd/services/:sid/targets/:team_id/access`
- `POST /contests/:id/awd/services/:sid/defense/ssh`

## 7. 兼容与迁移

- 当前实现没有独立学生 AWD 页面，旧稿里的 `/contests/:id/awd` 路线不属于现状。
- 学生战场当前是“竞赛详情页内嵌 AWD 面板”，不是另起一套竞赛 shell。
- 防守入口当前已经固定为 SSH，旧稿中的浏览器文件树和在线编辑不再是学生当前能力。
- 普通题目刷题工作台保留原有 Jeopardy 流程，不受 AWD 战场 read model 影响。

## 8. 代码落点

- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/useContestAWDWorkspace.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAccessActions.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackSubmission.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceServiceActions.ts`
- `code/frontend/src/api/contest.ts`
- `code/backend/internal/dto/contest_awd_workspace.go`
- `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- `code/backend/internal/app/router_routes.go`

## 9. 验证标准

- AWD 赛事进入 `/contests/:id` 时，“题目”tab 会切换为“攻防战场”面板。
- 工作台能读取 `current_round`、`services`、`targets`、`recent_events` 和榜单摘要。
- 启动、重启、攻击提交、目标访问和 SSH 连接五类动作都能从同一页面发起。
- 赛事 `running / frozen` 时会每 15 秒刷新；其他状态不会保持自动轮询。
- 学生页面不会出现独立 AWD 路由或浏览器文件工作台入口。
