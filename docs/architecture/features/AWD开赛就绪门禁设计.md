# AWD 开赛就绪门禁架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`backend`、`frontend`、`contracts`
- 关联模块：
  - `internal/module/contest/domain`
  - `internal/module/contest/application/queries`
  - `internal/module/contest/application/commands`
  - `internal/module/contest/api/http`
  - `frontend/src/components/platform/contest`
- 过程追溯：`practice/superpowers-plan-index.md` 中的 `2026-04-12-awd-engine-phase8-readiness-gate`
- 最后更新：`2026-05-07`

## 1. 背景与问题

AWD readiness 现在已经是正式门禁，而不是配置列表上的被动提示。当前代码把 checker 校验状态、镜像可用性和题目关联状态接进了关键管理动作，需要明确当前真实边界：

- readiness 事实源来自赛事 service 当前保存态，而不是前端草稿
- 创建轮次、当前轮巡检、赛事启动都已经受 readiness gate 约束
- 允许一次性强制放行，但必须带原因，并落审计快照

## 2. 架构结论

- `contestdomain.BuildAWDReadiness` 是 readiness 聚合唯一事实入口。
- readiness gate 当前只拦截：
  - `create_round`
  - `run_current_round_check`
  - `start_contest`
- 强制放行只对人工触发的关键动作开放；定时自动开赛不会强制放行。
- readiness 不通过时返回 `ErrAWDReadinessBlocked`，不会偷偷放行再写 warning。
- 放行不会改写 `awd_checker_validation_state`，只表示这次动作被审计后继续执行。

## 3. 模块边界与职责

### 3.1 模块清单

- `contestdomain.BuildAWDReadiness`
  - 负责：从 challenge 记录构建摘要、统计计数和阻塞原因
  - 不负责：执行拦截

- `AWDService.GetReadiness`
  - 负责：为前端提供只读 readiness DTO
  - 不负责：修改 contest 状态或 service 配置

- `ensureAWDReadinessGate`
  - 负责：在命令执行前重新加载 summary，并决定是否允许
  - 不负责：生成前端展示文案

- `writeAWDReadinessAuditPayload`
  - 负责：在强制放行成功时写入阻塞快照与放行原因
  - 不负责：决定 gate 判定本身

- `AWDOperationsPanel` / `useAwdReadinessDecision`
  - 负责：展示 readiness 摘要、在被阻塞时弹出 override 对话框
  - 不负责：前端本地重算 readiness

### 3.2 事实源与所有权

- readiness 记录源：`contestports.AWDReadinessQuery.ListReadinessChallengesByContest`
- 阻塞原因 owner：`internal/module/contest/domain/awd_readiness.go`
- 审计快照 owner：`internal/module/contest/api/http/awd_readiness_audit.go`

## 4. 关键模型与不变量

### 4.1 核心实体

- `AWDReadinessSummary`
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

- `AWDReadinessItem`
  - `service_id`
  - `awd_challenge_id`
  - `title`
  - `checker_type`
  - `validation_state`
  - `last_preview_at`
  - `last_access_url`
  - `blocking_reason`

### 4.2 不变量

- readiness 只根据保存态 `checker_type`、`checker_config`、`validation_state` 和 runtime image 状态计算。
- 没有关联题目时一定返回 `ready = false` 且 `global_blocking_reasons = ["no_challenges"]`。
- `passed` 是唯一非阻塞校验状态；`pending`、`failed`、`stale` 都是阻塞。
- 存在 `RuntimeImageID` 且镜像状态不是 `available` 时，阻塞原因优先记为 `image_not_available`。
- 强制放行必须提供非空 `override_reason`；空白原因会被视为无效请求。

## 5. 关键链路

### 5.1 只读查询链路

1. 前端调用 `GET /admin/contests/:id/awd/readiness`。
2. `AWDService.GetReadiness` 读取 readiness challenge 记录。
3. `BuildAWDReadiness` 输出摘要、计数和阻塞明细。
4. `AWDReadinessChecklist` 与 `AWDReadinessOverrideDialog` 展示系统级阻塞和题目级阻塞短名单。

### 5.2 命令门禁链路

1. 创建轮次、当前轮巡检、赛事切到 `running` 前，命令层调用 `ensureAWDReadinessGate`。
2. gate 再次加载最新 readiness summary，不依赖前端预读结果。
3. readiness 通过则继续执行；不通过则返回 `ErrAWDReadinessBlocked`。
4. 若请求带 `force_override=true` 且提供原因，则允许继续执行，并通过 trace 把阻塞快照写入审计。

### 5.3 自动开赛链路

1. `StatusUpdater.shouldBlockAutomaticAWDStart` 在 `registration -> running` 自动推进前重算 readiness。
2. 只要 summary 不 ready 或查询异常，就阻止自动开赛并记录日志。
3. 自动任务没有 override 能力。

## 6. 接口与契约

### 6.1 阻塞原因契约

当前已实现的原因码包括：

- `image_not_available`
- `missing_checker`
- `invalid_checker_config`
- `pending_validation`
- `last_preview_failed`
- `validation_stale`
- `no_challenges`

### 6.2 前后端动作契约

- 后端门禁动作：
  - `create_round`
  - `run_current_round_check`
  - `start_contest`

- 前端 override 对话框当前只处理：
  - 创建轮次
  - 当前轮巡检

赛事启动的强制放行入口走 `UpdateContest` 请求体中的：

- `force_override`
- `override_reason`

## 7. 兼容与迁移

- 指定历史轮次重跑、人工补录服务检查、人工补录攻击日志、题目配置编辑与 checker preview 都不受 gate 限制。
- gate 只对 AWD 赛事生效；Jeopardy 赛事不会进入这条链路。
- readiness 目前不读取对话框未保存草稿，也不做后台自动重试或自动修复。

## 8. 代码落点

- `code/backend/internal/module/contest/domain/awd_readiness.go`
- `code/backend/internal/module/contest/domain/contest.go`
- `code/backend/internal/module/contest/application/queries/awd_readiness_query.go`
- `code/backend/internal/module/contest/application/commands/awd_readiness_gate.go`
- `code/backend/internal/module/contest/application/commands/awd_round_admin_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/commands/contest_update_commands.go`
- `code/backend/internal/module/contest/application/jobs/status_awd_readiness.go`
- `code/backend/internal/module/contest/api/http/awd_readiness_audit.go`
- `code/frontend/src/components/platform/contest/AWDOperationsPanel.vue`
- `code/frontend/src/components/platform/contest/AWDReadinessChecklist.vue`
- `code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts`

## 9. 验证标准

- 无题目、镜像未就绪、缺少 checker、最近校验失败或 stale 时，关键动作都会被门禁拦截。
- 提供 `force_override=true` 和非空原因后，创建轮次、当前轮巡检和赛事启动可以继续执行，并保留审计快照。
- 自动开赛在 readiness 不通过时会继续停留在 `registration`，不会无条件推进到 `running`。
