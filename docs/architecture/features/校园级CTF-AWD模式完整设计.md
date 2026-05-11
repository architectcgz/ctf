# 校园级 CTF AWD 模式架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`backend`、`frontend`、`runtime`、`assessment`
- 关联模块：
  - `code/backend/internal/model`
  - `code/backend/internal/module/challenge`
  - `code/backend/internal/module/contest`
  - `code/backend/internal/module/runtime`
  - `code/backend/internal/module/assessment`
  - `code/frontend/src/views/contests`
  - `code/frontend/src/views/teacher`
  - `code/frontend/src/router/routes`
- 关联文档：
  - `docs/architecture/features/AWD学员实战工作台设计.md`
  - `docs/architecture/features/AWD防守工作区与边界设计.md`
  - `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`
  - `docs/architecture/features/教学复盘优化设计.md`
- 过程追溯：旧稿 `校园级CTF-AWD模式完整设计.md`
- 最后更新：`2026-05-09`

## 1. 背景与问题

旧稿把这篇写成了“参考公开平台后，准备构建的最终 AWD 方案”。当前仓库已经有可运行的 AWD 子系统，因此这里需要给出的不是目标蓝图，而是当前代码真实形成的系统边界：

- AWD 与 Jeopardy 已经在资源层和赛事层分开建模
- 比赛编排、运行时、计分、学生入口和教师复盘都已经落在主平台内
- 深度防守入口已经收敛为 SSH，不是浏览器 IDE
- 学生战场、SSH 防守边界、教师赛事复盘和个人教学复盘都已经各自有 owning 专题；本文只保留六层总览和跨层主身份

## 2. 架构结论

- 当前平台的 AWD 是主平台内部的一套比赛子系统，不是独立外置 gameserver。
- AWD 资源池与 Jeopardy 资源池已经分离：
  - Jeopardy 题目对象：`Challenge`
  - AWD 题目对象：`AWDChallenge`
- 赛事内的 AWD 服务编排对象是 `ContestAWDService`，而不是直接复用 `contest_challenges`。
- AWD 运行态统一以 `service_id = contest_awd_services.id` 作为主身份；`awd_challenge_id` 退回为题目资产与展示字段。
- 运行态配置、分值配置和 readiness 校验统一从 `contest_awd_services` 读取；`contest_challenges` 只保留题目关系和编排字段。
- 轮次与 checker 是当前官方裁判内核，核心运行事实落在：
  - `awd_rounds`
  - `awd_team_services`
  - `awd_attack_logs`
  - `awd_traffic_events`
- 当前学生攻防入口是“内嵌战场工作台 + target proxy + defense SSH”，不是浏览器源码编辑器。
- 当前教师入口是“AWD 赛事复盘 + 学生证据读模型 + 评估导出”三条并行链路，而不是单一大而全读模型。
- 学生战场、SSH 防守边界、AWD 赛事 archive 和个人教学复盘细节分别由对应 owning 专题文档继续展开，本文不重复列第二套接口和页面行为说明。
- 当前已支持的 checker 类型是：
  - `legacy_probe`
  - `http_standard`
  - `tcp_standard`
  - `script_checker`

## 3. 模块边界与职责

### 3.1 六层结构

- 资源与题包层
  - 负责：`AWDChallenge`、题包解析、checker/flag/runtime 配置入库
  - owner：`internal/module/challenge`

- 赛事配置层
  - 负责：把 `AWDChallenge` 关联到具体赛事，形成 `ContestAWDService`，并做 readiness 审计
  - owner：`internal/module/contest`

- 运行编排层
  - 负责：实例生命周期、服务操作、defense workspace、目标代理和 SSH 接入
  - owner：`internal/module/runtime` + `internal/module/contest`

- 裁判计分层
  - 负责：轮次调度、checker 执行、服务状态写回、攻击记录与榜单刷新
  - owner：`AWDRoundUpdater`

- 学生战场层
  - 负责：读取当前轮次、本队服务、攻击目标、最近事件，并发起启动/重启/攻击/SSH 动作
  - owner：`ContestDetail` + `ContestAWDWorkspacePanel`

- 教师复盘与评估层
  - 负责：AWD 赛事复盘、学生证据阅读、报告和归档导出
  - owner：`assessment` + `teaching_readmodel`

### 3.2 事实源与所有权

- AWD 题库事实源：`awd_challenges`
- AWD 赛事服务事实源：`contest_awd_services`
- AWD 运行态配置与 readiness 事实源：`contest_awd_services.runtime_config`、`score_config`、`awd_checker_validation_state`
- 赛事题目关系与编排事实源：`contest_challenges`
- 轮次裁判事实源：`awd_rounds`、`awd_team_services`
- 攻击与流量事实源：`awd_attack_logs`、`awd_traffic_events`
- 防守工作区事实源：`awd_defense_workspaces`
- 教师赛事复盘事实源：`TeacherAWDReviewService`
- 学生证据与攻击会话事实源：`teaching_readmodel`

## 4. 关键模型与不变量

### 4.1 核心实体

- `AWDChallenge`
  - 关键字段：`service_type`、`deployment_mode`、`checker_type`、`flag_mode`、`defense_entry_mode`、`runtime_config`、`readiness_status`

- `ContestAWDService`
  - 关键字段：`id`、`contest_id`、`awd_challenge_id`、`display_name`、`score_config`、`runtime_config`、`validation_state`

- `AWDRound`
  - 关键字段：`contest_id`、`round_number`、`status`、`attack_score`、`defense_score`

- `AWDTeamService`
  - 关键字段：`round_id`、`team_id`、`service_id`、`service_status`、`checker_type`、`sla_score`、`defense_score`、`attack_score`

- `AWDAttackLog`
  - 关键字段：`attacker_team_id`、`victim_team_id`、`service_id`、`attack_type`、`source`、`is_success`、`score_gained`

- `AWDDefenseWorkspace`
  - 关键字段：`contest_id`、`team_id`、`service_id`、`instance_id`、`workspace_revision`、`status`

### 4.2 不变量

- AWD 与 Jeopardy 不是同一个题库对象，不能把 `Challenge` 直接当成 AWD 资源主身份。
- `service_id` 是 AWD 运行态唯一主身份，不能再用 `awd_challenge_id` 充当轮次结果、攻击去重、实例定位和流量归因主键。
- `contest_awd_services` 是当前运行态配置与 readiness 的主事实源；`contest_challenges` 不再承载 checker、SLA、防守分和校验状态。
- 当前官方比赛结果只认平台记录的轮次、checker 和攻击提交结果。
- 学生深度防守入口当前只能是 SSH，不是浏览器 IDE。
- `AWDRoundUpdater` 是当前轮次推进和官方状态写回 owner。
- 教师赛事复盘与学生证据读模型当前并存，不是同一套聚合对象。

## 5. 关键链路

### 5.1 资源导入与赛事配置链路

1. 题包通过 `awd_package_parser` 解析 checker、runtime、defense workspace 等配置。
2. 平台把资源写入 `AWDChallenge`。
3. 管理员把 `AWDChallenge` 关联到具体赛事，形成 `ContestAWDService`，并把运行态配置、分值配置和校验状态收口到这一层。
4. readiness 审计根据 service 配置和题包能力生成赛前阻塞项。

### 5.2 运行与裁判链路

1. `AWDRoundUpdater.Start` 按调度周期推进轮次。
2. 每轮基于 `contest_awd_services` 提供的赛事服务定义、队伍、实例和 checker 生成官方状态。
3. 结果写回 `awd_team_services`，并更新 scoreboard cache。
4. 学生攻击提交流程按 `service_id` 校验当前轮和上一轮宽限 Flag，并把成功记录写入 `awd_attack_logs`。
5. 目标代理流量按 `service_id` 写入 `awd_traffic_events`，来源当前标记为 `runtime_proxy`。

### 5.3 学生战场链路

1. 学生通过 `/contests/:id` 进入内嵌 AWD 战场。
2. 工作台 read model、动作接口和前端 owner 由 `AWD学员实战工作台设计.md` 继续展开。
3. 深度防守入口与 `/workspace` SSH 边界由 `AWD防守工作区与边界设计.md` 继续展开。
4. 当前学生深度防守入口仍只有 SSH，不存在浏览器文件工作台。

### 5.4 教师复盘链路

1. 教师通过 `/academy/awd-reviews` 查看 AWD 赛事目录与详情。
2. AWD 赛事级 archive、ZIP/PDF 导出和详情页切片由 `AWD教师复盘归档与报告导出设计.md` 继续展开。
3. 教师个人教学复盘工作台、证据链、攻击会话和稳定归档快照由 `教学复盘优化设计.md` 及其关联专题继续展开。
4. 当前教师侧并行保留“赛事复盘”与“个人教学复盘”两条入口，不合并成单一聚合对象。

## 6. 接口与契约

### 6.1 学生接口

- 学生战场接口组详见：
  - `AWD学员实战工作台设计.md`
  - `AWD防守工作区与边界设计.md`

### 6.2 教师接口

- AWD 赛事复盘接口组详见 `AWD教师复盘归档与报告导出设计.md`
- 教师个人复盘接口组详见：
  - `教学复盘优化设计.md`
  - `攻击证据链与教学复盘架构.md`
  - `攻击会话读模型与复盘工作台架构.md`

### 6.3 赛事配置接口

- `GET /admin/contests/:id/awd/readiness`
- `GET /admin/contests/:id/awd/services`
- `POST /admin/contests/:id/awd/services`
- `PUT /admin/contests/:id/awd/services/:sid`

## 7. 兼容与迁移

- 当前实现没有外置的独立 gameserver、checker 集群控制面或公开赛级多机房设施。
- 当前学生端不提供浏览器内完整 shell 或源码编辑器。
- 当前 teacher 前端路由命名空间已经收敛到 `/academy/*`，但后端 API 仍保持 `/teacher/*`。
- 旧稿中的公开赛级基础设施设想和 VPN / IDE 路线，不应再被当成当前实现事实。

## 8. 代码落点

- `code/backend/internal/model/awd.go`
- `code/backend/internal/model/awd_challenge.go`
- `code/backend/internal/model/contest_awd_service.go`
- `code/backend/internal/model/awd_service_operation.go`
- `code/backend/internal/model/awd_defense_workspace.go`
- `code/backend/internal/model/contest_challenge.go`
- `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- `code/backend/internal/module/contest/application/queries/awd_summary_query.go`
- `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
- `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository.go`
- `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
- `code/backend/internal/app/router_routes.go`
- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`

## 9. 验证标准

- AWD 题库、赛事服务、轮次状态、攻击记录和防守工作区都已经有独立模型，不再依附 Jeopardy 题目关系表。
- 学生 AWD 战场能够读取 workspace、访问目标代理、提交攻击结果并生成 defense SSH。
- 当前 checker 类型至少覆盖 `legacy_probe / http_standard / tcp_standard / script_checker`。
- `AWDRoundUpdater` 负责轮次推进与榜单更新，官方状态可落到 `awd_team_services`。
- 教师能够同时访问 AWD 赛事复盘和学生证据读模型，两者不会被误写成同一套查询对象。
