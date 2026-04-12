# AWD Phase 8 Readiness Gate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 AWD checker 校验状态升级为可执行的开赛就绪门禁，让创建轮次、立即巡检当前轮、以及 AWD 赛事切到 `running` 时默认被 readiness 阻止，并支持一次性、可审计的强制放行。

**Architecture:** 后端沿用现有 contest/AWD 命令入口，不新开独立写接口；新增一层共享 readiness query + gate helper，再通过专用错误码和附加审计把门禁结果传回前端。前端不新增页面，而是在 `AWDOperationsPanel` 内增加 readiness 摘要区，并复用一个阻塞确认弹层处理创建轮次、立即巡检和启动赛事三类动作，视觉保持现有 CTF 管理后台语义。

**Tech Stack:** Go, Gin, GORM, Vue 3, TypeScript, Vitest, Vue Test Utils, existing audit middleware, existing admin AWD UI

---

## Execution Notes

- 后端执行严格按 `@superpowers:test-driven-development` 先补 RED 测试，再做最小实现。
- 前端 AWD 运维页与弹层必须显式使用 `@ctf-ui-theme-system`、`@frontend-skill`，必要时补 `@ctf-dark-surface-alignment`，不要做脱离现有工作区视觉的独立样式。
- 跑测试和类型检查时遵守 `@runtime-ops-safety`，所有命令都在当前 worktree 内执行，不保留后台进程。
- 对外声称“完成”前，必须执行 `@superpowers:verification-before-completion`。

## Planned File Map

### Backend: readiness query / gate / audit

- Modify: `code/backend/internal/dto/awd.go`
  - 增加 readiness DTO、创建轮次和当前轮巡检的 override 请求字段。
- Modify: `code/backend/internal/dto/contest.go`
  - 给 `UpdateContestReq` 增加 `force_override` / `override_reason`。
- Modify: `code/backend/pkg/errcode/errcode.go`
  - 增加专用 `ErrAWDReadinessBlocked`，让前端通过 `Envelope.code` 稳定分流。
- Modify: `code/backend/internal/module/contest/ports/awd.go`
  - 为 readiness 查询增加新的 repo 读取能力，避免把 query 逻辑塞回 handler。
- Create: `code/backend/internal/module/contest/domain/awd_readiness.go`
  - 收敛 blocking action、blocking reason、全局阻塞原因和状态归类 helper。
- Create: `code/backend/internal/module/contest/application/queries/awd_readiness_query.go`
  - 负责把 repo 读取结果组装成 readiness DTO。
- Create: `code/backend/internal/module/contest/application/commands/awd_readiness_gate.go`
  - 负责对三类关键动作执行 gate 判定、trim `override_reason`、返回阻塞快照。
- Modify: `code/backend/internal/module/contest/application/commands/awd_service.go`
  - 给 AWD command service 注入 readiness gate 依赖。
- Modify: `code/backend/internal/module/contest/application/commands/awd_round_admin_commands.go`
  - 在创建轮次前走 readiness gate。
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
  - 只给“当前轮巡检”接 gate，不影响指定轮次重跑。
- Modify: `code/backend/internal/module/contest/application/commands/contest_service.go`
  - 给 contest command service 注入 AWD readiness gate 依赖。
- Modify: `code/backend/internal/module/contest/application/commands/contest_update_commands.go`
  - 只在 AWD 且状态切到 `running` 时执行 gate。
- Modify: `code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go`
  - 提供 readiness 所需的 contest challenge 读取查询。
- Modify: `code/backend/internal/module/contest/api/http/awd_handler.go`
  - 扩展 command/query 接口定义。
- Modify: `code/backend/internal/module/contest/api/http/handler.go`
  - 给 `UpdateContest` 所在 handler 注入 readiness query，用于 forced override 审计预取。
- Create: `code/backend/internal/module/contest/api/http/awd_readiness_handler.go`
  - 新增 `GET /admin/contests/:id/awd/readiness` handler。
- Modify: `code/backend/internal/module/contest/api/http/awd_round_manage_handler.go`
  - 绑定创建轮次 body。
- Modify: `code/backend/internal/module/contest/api/http/awd_round_check_handler.go`
  - 让当前轮巡检支持 JSON body。
- Modify: `code/backend/internal/module/contest/api/http/contest_command_handler.go`
  - 接受 `UpdateContest` 的 override 字段。
- Create: `code/backend/internal/middleware/awd_readiness_audit.go`
  - 读取 handler 放进 context 的 override 审计快照，在动作成功/失败后都写 `admin_op` 审计。
- Modify: `code/backend/internal/app/router_routes.go`
  - 注册 readiness GET 路由，并把 readiness 审计 middleware 接到 3 个关键动作。
- Modify: `code/backend/internal/app/router_test.go`
  - 断言新 readiness 路由存在。
- Modify: `code/backend/internal/app/composition/contest_module.go`
  - 把 readiness query/gate 装配进 contest/AWD handler 使用的 service。

### Backend tests

- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
  - 覆盖创建轮次、当前轮巡检的 gate 行为。
- Modify: `code/backend/internal/module/contest/application/commands/contest_service_test.go`
  - 覆盖 AWD 赛事切到 `running` 的 gate 行为。
- Create: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
  - 覆盖 readiness 摘要查询与零题目系统级阻塞。

### Frontend: readiness summary / override dialog / guarded actions

- Modify: `code/frontend/src/api/contracts.ts`
  - 增加 readiness 数据结构和 override payload 字段。
- Modify: `code/frontend/src/api/admin.ts`
  - 新增 readiness GET；扩展创建轮次、当前轮巡检、更新赛事状态的 override 参数；为 readiness-aware 调用支持 `suppressErrorToast`。
- Modify: `code/frontend/src/composables/useAdminContestAWD.ts`
  - 拉取 readiness；在创建轮次、立即巡检前捕获 `ErrAWDReadinessBlocked` 并打开 override 流程。
- Modify: `code/frontend/src/composables/useAdminContests.ts`
  - 在 AWD 赛事切到 `running` 时支持 readiness block -> reload -> override 重提交流程。
- Create: `code/frontend/src/components/admin/contest/AWDReadinessSummary.vue`
  - 渲染 readiness 概况卡与阻塞短名单。
- Create: `code/frontend/src/components/admin/contest/AWDReadinessOverrideDialog.vue`
  - 渲染阻塞确认层、原因输入、系统级阻塞摘要。
- Modify: `code/frontend/src/components/admin/contest/AWDOperationsPanel.vue`
  - 在赛事选择与 tab rail 之间插入 readiness 区块，并接入创建轮次/当前轮巡检 override dialog。
- Modify: `code/frontend/src/views/admin/ContestManage.vue`
  - 承接 AWD 开赛状态切换的 override dialog。

### Frontend tests

- Modify: `code/frontend/src/api/__tests__/admin.test.ts`
  - 覆盖 readiness GET、override payload 和 gate-aware request 选项。
- Modify: `code/frontend/src/components/admin/__tests__/AWDOperationsPanel.test.ts`
  - 覆盖 readiness 区块、零题目提示、被拦动作弹层。
- Modify: `code/frontend/src/views/admin/__tests__/ContestManage.test.ts`
  - 覆盖 AWD 赛事切到 `running` 时的门禁与强制放行。

## Task 1: 为后端 readiness query / gate 补 RED 测试

**Files:**
- Create: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_service_test.go`
- Modify: `code/backend/internal/app/router_test.go`

- [x] **Step 1: 在 readiness query 测试里锁定摘要口径**

在 `awd_service_test.go`（queries）新增至少 3 个用例：

- `TestAWDQueryServiceGetReadinessCountsBlockingStates`
- `TestAWDQueryServiceGetReadinessTreatsZeroChallengesAsGlobalBlock`
- `TestAWDQueryServiceGetReadinessTreatsBrokenCheckerConfigAsMissingChecker`

断言至少包含：

```go
if resp.TotalChallenges != 0 || resp.PassedChallenges != 0 {
	t.Fatalf("unexpected readiness counts: %+v", resp)
}
if resp.BlockingCount != 1 {
	t.Fatalf("expected blocking count 1, got %d", resp.BlockingCount)
}
if len(resp.GlobalBlockingReasons) != 1 || resp.GlobalBlockingReasons[0] != "no_challenges" {
	t.Fatalf("unexpected global blocking reasons: %+v", resp.GlobalBlockingReasons)
}
if len(resp.Items) > 0 && resp.Items[0].LastAccessURL == nil {
	t.Fatalf("expected last_access_url for preview-backed item: %+v", resp.Items[0])
}
```

- [x] **Step 2: 在 AWD command 测试里锁定创建轮次与当前轮巡检 gate**

在 `awd_service_test.go` 新增用例：

- `TestAWDServiceCreateRoundBlocksWhenReadinessNotReady`
- `TestAWDServiceCreateRoundAllowsForceOverrideWithReason`
- `TestAWDServiceCreateRoundRejectsBlankOverrideReason`
- `TestAWDServiceRunCurrentRoundChecksBlocksWhenReadinessNotReady`
- `TestAWDServiceRunCurrentRoundChecksRejectsBlankOverrideReason`
- `TestAWDServiceRunRoundChecksSkipsReadinessGate`

错误断言统一按专用错误码写：

```go
var appErr *errcode.AppError
if !errors.As(err, &appErr) || appErr.Code != errcode.ErrAWDReadinessBlocked.Code {
	t.Fatalf("expected ErrAWDReadinessBlocked, got %v", err)
}
```

- [x] **Step 3: 在 contest command 测试里锁定 AWD 开赛 gate**

在 `contest_service_test.go` 新增：

- `TestContestServiceUpdateContestBlocksAWDStartWhenReadinessNotReady`
- `TestContestServiceUpdateContestAllowsAWDStartOverride`
- `TestContestServiceUpdateContestRejectsBlankOverrideReason`
- `TestContestServiceUpdateContestDoesNotGateNonAWDStatusUpdate`

最小断言：

```go
resp, err := service.UpdateContest(ctx, contest.ID, &dto.UpdateContestReq{
	Status:         strPtr(model.ContestStatusRunning),
	ForceOverride:  boolPtr(true),
	OverrideReason: strPtr("teacher drill"),
})
```

- [x] **Step 4: 给 router 测试补 readiness 路由断言**

在 `router_test.go` 增加：

```go
assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id/awd/readiness", "internal/module/contest/api/http")
```

- [x] **Step 5: 跑后端定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate/code/backend
go test ./internal/module/contest/application/queries -run AWDService -count=1
go test ./internal/module/contest/application/commands -run 'AWDService|ContestService' -count=1
go test ./internal/app -run TestNewRouterRegistersStudentChallengeRoutes -count=1
```

预期：FAIL，失败点分别落在 readiness DTO/route 不存在、gate 未接入、专用错误码未定义。

## Task 2: 实现后端 readiness query、gate 和 HTTP 契约

**Files:**
- Modify: `code/backend/internal/dto/awd.go`
- Modify: `code/backend/internal/dto/contest.go`
- Modify: `code/backend/pkg/errcode/errcode.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Create: `code/backend/internal/module/contest/domain/awd_readiness.go`
- Create: `code/backend/internal/module/contest/application/queries/awd_readiness_query.go`
- Create: `code/backend/internal/module/contest/application/commands/awd_readiness_gate.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_round_admin_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_update_commands.go`
- Modify: `code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go`
- Modify: `code/backend/internal/module/contest/application/queries/awd_service.go`
- Modify: `code/backend/internal/module/contest/api/http/awd_handler.go`
- Create: `code/backend/internal/module/contest/api/http/awd_readiness_handler.go`
- Modify: `code/backend/internal/module/contest/api/http/awd_round_manage_handler.go`
- Modify: `code/backend/internal/module/contest/api/http/awd_round_check_handler.go`
- Modify: `code/backend/internal/module/contest/api/http/contest_command_handler.go`
- Modify: `code/backend/internal/app/composition/contest_module.go`
- Modify: `code/backend/internal/app/router_routes.go`

- [x] **Step 1: 先补 DTO 与错误码**

在 `dto/awd.go` 定义 readiness 响应和 override body 字段，目标结构至少是：

```go
type AWDReadinessResp struct {
	ContestID             int64                    `json:"contest_id"`
	Ready                 bool                     `json:"ready"`
	TotalChallenges       int                      `json:"total_challenges"`
	PassedChallenges      int                      `json:"passed_challenges"`
	PendingChallenges     int                      `json:"pending_challenges"`
	FailedChallenges      int                      `json:"failed_challenges"`
	StaleChallenges       int                      `json:"stale_challenges"`
	MissingCheckerChallenges int                   `json:"missing_checker_challenges"`
	BlockingCount         int                      `json:"blocking_count"`
	BlockingActions       []string                 `json:"blocking_actions"`
	GlobalBlockingReasons []string                 `json:"global_blocking_reasons"`
	Items                 []*AWDReadinessItemResp  `json:"items"`
}

type AWDReadinessItemResp struct {
	ChallengeID     int64                `json:"challenge_id"`
	Title           string               `json:"title"`
	CheckerType     model.AWDCheckerType `json:"checker_type,omitempty"`
	ValidationState string               `json:"validation_state"`
	LastPreviewAt   *time.Time           `json:"last_preview_at,omitempty"`
	LastAccessURL   *string              `json:"last_access_url,omitempty"`
	BlockingReason  string               `json:"blocking_reason"`
}
```

同时在 `dto/contest.go` 与现有 AWD 请求 DTO 中加入：

```go
ForceOverride  *bool   `json:"force_override"`
OverrideReason *string `json:"override_reason" binding:"omitempty,max=500"`
```

并在 `pkg/errcode/errcode.go` 增加 `ErrAWDReadinessBlocked`。

- [x] **Step 2: 实现 repo 读取与 readiness query**

在 `ports/awd.go` 增加新的读取结构与方法，例如：

```go
type AWDReadinessChallengeRecord struct {
	ChallengeID               int64
	Title                     string
	AWDCheckerType            model.AWDCheckerType
	AWDCheckerConfig          string
	AWDCheckerValidationState model.AWDCheckerValidationState
	AWDCheckerLastPreviewAt   *time.Time
	AWDCheckerLastPreviewResult string
}
```

`awd_contest_relation_repository.go` 负责一次性查出题目标题、checker 配置、validation state 与 preview 快照；`awd_readiness_query.go` 负责把它们转换成：

- `missing_checker`
- `invalid_checker_config`
- `pending_validation`
- `last_preview_failed`
- `validation_stale`
- `global_blocking_reasons = ["no_challenges"]`
- `last_access_url = preview_context.access_url`

- [x] **Step 3: 实现共享 gate helper**

在 `awd_readiness_gate.go` 提供单一入口，例如：

```go
decision, err := gate.Evaluate(ctx, contestID, action, forceOverride, overrideReason)
```

行为要求：

- 无阻塞：放行
- 有阻塞且未强制：返回 `ErrAWDReadinessBlocked`
- 有阻塞且 `force_override = true`：
  - trim 原因
  - trim 后为空时报 `ErrInvalidParams`
  - 长度合法后返回 allow + blocking snapshot

- [x] **Step 4: 把 gate 接到 3 个命令路径**

最小接入点固定为：

- `awd_round_admin_commands.go` 的 `CreateRound`
- `awd_service_run_commands.go` 的 `RunCurrentRoundChecks`
- `contest_update_commands.go` 的 `UpdateContest`

其中 `RunRoundChecks(contestID, roundID)` 保持不变，不调用 gate。

- [x] **Step 5: 补 HTTP handler 与路由**

实现：

- `GET /api/v1/admin/contests/:id/awd/readiness`
- 当前轮巡检支持空 body 或 JSON body
- 创建轮次/更新赛事状态都能透传 override 字段

并在 `router_routes.go` 注册：

```go
adminOnly.GET("/contests/:id/awd/readiness", middleware.ParseInt64Param("id"), deps.contest.AWDHandler.GetReadiness)
```

- [x] **Step 6: 跑后端定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate/code/backend
go test ./internal/module/contest/application/queries -run AWDService -count=1
go test ./internal/module/contest/application/commands -run 'AWDService|ContestService' -count=1
go test ./internal/app -run TestNewRouterRegistersStudentChallengeRoutes -count=1
```

预期：PASS。

- [x] **Step 7: 提交后端门禁主链路**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate
git add code/backend/internal/dto/awd.go code/backend/internal/dto/contest.go code/backend/pkg/errcode/errcode.go code/backend/internal/module/contest/ports/awd.go code/backend/internal/module/contest/domain/awd_readiness.go code/backend/internal/module/contest/application/queries/awd_readiness_query.go code/backend/internal/module/contest/application/commands/awd_readiness_gate.go code/backend/internal/module/contest/application/commands/awd_service.go code/backend/internal/module/contest/application/commands/awd_round_admin_commands.go code/backend/internal/module/contest/application/commands/awd_service_run_commands.go code/backend/internal/module/contest/application/commands/contest_service.go code/backend/internal/module/contest/application/commands/contest_update_commands.go code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go code/backend/internal/module/contest/application/queries/awd_service.go code/backend/internal/module/contest/api/http/awd_handler.go code/backend/internal/module/contest/api/http/awd_readiness_handler.go code/backend/internal/module/contest/api/http/awd_round_manage_handler.go code/backend/internal/module/contest/api/http/awd_round_check_handler.go code/backend/internal/module/contest/api/http/contest_command_handler.go code/backend/internal/app/composition/contest_module.go code/backend/internal/app/router_routes.go code/backend/internal/module/contest/application/queries/awd_service_test.go code/backend/internal/module/contest/application/commands/awd_service_test.go code/backend/internal/module/contest/application/commands/contest_service_test.go code/backend/internal/app/router_test.go
git commit -m "feat(awd): 增加开赛就绪门禁后端链路"
```

## Task 3: 为强制放行补独立审计链路并做后端回归

**Files:**
- Create: `code/backend/internal/middleware/awd_readiness_audit.go`
- Create: `code/backend/internal/middleware/awd_readiness_audit_test.go`
- Modify: `code/backend/internal/module/contest/api/http/handler.go`
- Modify: `code/backend/internal/module/contest/api/http/awd_round_manage_handler.go`
- Modify: `code/backend/internal/module/contest/api/http/awd_round_check_handler.go`
- Modify: `code/backend/internal/module/contest/api/http/contest_command_handler.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_service_test.go`

- [x] **Step 1: 先写 forced override 审计失败测试**

优先在 `awd_readiness_audit_test.go` 里直接覆盖 middleware 行为，至少锁定两类结果：

- `execution_outcome = "succeeded"`
- `execution_outcome = "failed"`

至少锁定 detail 结构：

```go
if entry.Action != model.AuditActionAdminOp || entry.ResourceType != "contest" {
	t.Fatalf("unexpected audit envelope: %+v", entry)
}
if entry.ResourceID == nil || *entry.ResourceID != contestID {
	t.Fatalf("unexpected resource id: %+v", entry.ResourceID)
}
detail := entry.Detail
if detail["module"] != "awd_readiness_gate" {
	t.Fatalf("unexpected module: %+v", detail)
}
```

- [x] **Step 2: 用 middleware 记录附加审计，不碰现有通用审计中间件语义**

在 `awd_readiness_audit.go` 里实现一个专用 middleware：

- `c.Next()` 后读取 handler 放入 context 的 gate audit payload
- 只在 `force_override = true` 且 gate 已放行时写审计
- 无论最终响应是否 2xx，都记录：
  - `gate_action`
  - `override_reason`
  - `blocking_count`
  - `global_blocking_reasons`
  - `blocking_items`
  - `execution_outcome`
  - `execution_error`

- [x] **Step 3: 在 3 个 handler 里预取 readiness snapshot 并写入 context 审计快照**

实现约定改为：

- handler 在满足“目标动作可被 gate 且 `force_override = true`”时，先用与 command 相同口径的 readiness query 立刻取一份 snapshot
- command 仍独立执行自己的 gate，handler 不复制 gate 判定逻辑
- command 返回后，若不是 `ErrAWDReadinessBlocked` 且 snapshot 的 `blocking_count > 0`，handler 再把统一审计 payload 写进 context

这样保持：

- command 不需要修改返回签名
- handler 负责知道请求意图和最终 HTTP 结果
- middleware 负责真正写 audit log

其中 `PUT /admin/contests/:id` 需要在 `handler.go` 里给主 `Handler` 增加 readiness query 依赖，不能只改 `contest_command_handler.go`。

- [x] **Step 4: 把专用审计 middleware 接到 3 条路由**

在 `router_routes.go` 的以下路由上加 middleware：

- `POST /admin/contests/:id/awd/rounds`
- `POST /admin/contests/:id/awd/current-round/check`
- `PUT /admin/contests/:id`

顺序要求：放在业务 handler 同链路内，但不要替换现有通用审计 middleware。

- [x] **Step 5: 跑后端最小充分回归**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate/code/backend
go test ./internal/module/contest/application/commands -run 'AWDService|ContestService' -count=1
go test ./internal/middleware -run AWDReadinessAudit -count=1
go test ./internal/module/contest/... -count=1
```

预期：PASS。

- [x] **Step 6: 提交 readiness 审计补丁**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate
git add code/backend/internal/middleware/awd_readiness_audit.go code/backend/internal/middleware/awd_readiness_audit_test.go code/backend/internal/module/contest/api/http/handler.go code/backend/internal/module/contest/api/http/awd_round_manage_handler.go code/backend/internal/module/contest/api/http/awd_round_check_handler.go code/backend/internal/module/contest/api/http/contest_command_handler.go code/backend/internal/app/router_routes.go code/backend/internal/module/contest/application/commands/awd_service_test.go code/backend/internal/module/contest/application/commands/contest_service_test.go
git commit -m "feat(awd): 记录开赛门禁强制放行审计"
```

## Task 4: 为前端 readiness 摘要、门禁弹层和启动赛事拦截补 RED 测试

**Files:**
- Modify: `code/frontend/src/api/__tests__/admin.test.ts`
- Modify: `code/frontend/src/components/admin/__tests__/AWDOperationsPanel.test.ts`
- Modify: `code/frontend/src/views/admin/__tests__/ContestManage.test.ts`

- [x] **Step 1: 先写 admin API 契约失败测试**

补 4 类断言：

- `getContestAWDReadiness(contestId)` 命中 `GET /admin/contests/:id/awd/readiness`
- `createContestAWDRound` 可发送 `force_override` / `override_reason`
- `runContestAWDCurrentRoundCheck` 支持可选 body
- `updateContest` 可透传 override 字段，并允许调用方传 `suppressErrorToast`

最小断言示例：

```ts
expect(requestMock).toHaveBeenCalledWith({
  method: 'POST',
  url: '/admin/contests/awd-1/awd/current-round/check',
  data: { force_override: true, override_reason: 'teacher drill' },
  suppressErrorToast: true,
})
```

- [x] **Step 2: 给 AWDOperationsPanel 写 readiness 摘要与被拦动作测试**

至少补下面场景：

- readiness 概况卡正常渲染
- `global_blocking_reasons = ['no_challenges']` 时显示系统级阻塞提示
- 点击创建轮次时若 action 被 gate 拦截，会出现原因输入层而不是直接 toast
- 点击立即巡检当前轮时同样打开确认层
- 普通 `409` 或非 gate 错误码不会误开 readiness 弹层

- [x] **Step 3: 给 ContestManage 写 AWD 启动赛事门禁测试**

在 `ContestManage.test.ts` 增加：

- 编辑 AWD 赛事把状态切到 `running`，若后端抛出 `ApiError { code: ErrAWDReadinessBlocked }`
- 页面会重拉 readiness
- 打开强制放行弹层
- 填写原因后再次调用 `updateContest(..., { force_override: true, override_reason })`
- 普通 `409` 或其他错误码不会误开“启动赛事” gate 弹层

- [x] **Step 4: 跑前端定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/admin/__tests__/AWDOperationsPanel.test.ts src/views/admin/__tests__/ContestManage.test.ts
```

预期：FAIL，失败点分别对应 readiness API、readiness 区块、override dialog 和 AWD 开赛门禁尚未实现。

## Task 5: 实现前端 readiness 区块、override dialog 与 gated action 流程

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/admin.ts`
- Modify: `code/frontend/src/composables/useAdminContestAWD.ts`
- Modify: `code/frontend/src/composables/useAdminContests.ts`
- Create: `code/frontend/src/components/admin/contest/AWDReadinessSummary.vue`
- Create: `code/frontend/src/components/admin/contest/AWDReadinessOverrideDialog.vue`
- Modify: `code/frontend/src/components/admin/contest/AWDOperationsPanel.vue`
- Modify: `code/frontend/src/views/admin/ContestManage.vue`

- [x] **Step 1: 先接好 contract 和 admin API**

增加前端类型：

```ts
export interface AWDReadinessData {
  contest_id: ID
  ready: boolean
  total_challenges: number
  passed_challenges: number
  pending_challenges: number
  failed_challenges: number
  stale_challenges: number
  missing_checker_challenges: number
  blocking_count: number
  global_blocking_reasons: string[]
  blocking_actions: Array<'create_round' | 'run_current_round_check' | 'start_contest'>
  items: AWDReadinessItemData[]
}

export interface AWDReadinessItemData {
  challenge_id: ID
  title: string
  checker_type?: AdminContestChallengeData['awd_checker_type']
  validation_state: 'pending' | 'failed' | 'stale' | 'passed'
  last_preview_at?: ISODateTime
  last_access_url?: string
  blocking_reason: 'missing_checker' | 'invalid_checker_config' | 'pending_validation' | 'last_preview_failed' | 'validation_stale'
}
```

并在 `admin.ts` 增加：

- `getContestAWDReadiness`
- override 版 `createContestAWDRound`
- override 版 `runContestAWDCurrentRoundCheck`
- override 版 `updateContest`

readiness-aware 请求默认走：

```ts
suppressErrorToast: true
```

让 UI 自己决定何时弹 gate dialog。

- [x] **Step 2: 在 useAdminContestAWD 里接入 readiness 数据和 2 个 gated action**

新增状态至少包含：

- `readiness`
- `loadingReadiness`
- `overrideDialogState`

流程要求：

- 进入 AWD 运维页时主动拉 readiness
- 创建轮次被 `ApiError.code === ErrAWDReadinessBlocked` 拦截后：
  - 重拉 readiness
  - 打开 override dialog
  - 用户确认后带 `force_override` / `override_reason` 重试
- 当前轮巡检复用同一套流程

- [x] **Step 3: 用独立组件承接摘要区和强制放行弹层**

`AWDReadinessSummary.vue` 负责：

- 概况卡
- 阻塞短名单
- 零题目系统级阻塞提示
- `编辑配置` 入口透传

`AWDReadinessOverrideDialog.vue` 负责：

- 当前动作标题
- 系统级阻塞原因
- 题目阻塞列表
- 原因输入框
- `取消` / `强制继续`

视觉要求：

- 继续使用现有 `metric-panel`、`workspace-directory-section`、flat row
- 不引入新主题色，不脱离现有暗色工作区表面

- [x] **Step 4: 在 AWDOperationsPanel 接上新组件**

布局固定为：

- 赛事选择器
- readiness 摘要区
- tab rail
- inspector / challenges panel

不要把 readiness 放进新页面或 tab。

- [x] **Step 5: 在 useAdminContests + ContestManage 接入 AWD 启动赛事门禁**

要求：

- 只有 `mode === 'awd'` 且目标状态为 `running` 时才接 gate 流程
- 其他 contest 更新行为不受影响
- `AdminContestFormDialog.vue` 保持只负责 `emit('save')`，不把 gate 逻辑回灌进表单组件
- 若被 gate 拦截：
  - 读取当前编辑赛事 id
  - 拉取 readiness
  - 打开同一款 override dialog
  - 用户确认后带 override 参数重提交流程

- [x] **Step 6: 跑前端定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/admin/__tests__/AWDOperationsPanel.test.ts src/views/admin/__tests__/ContestManage.test.ts
```

预期：PASS。

- [x] **Step 7: 提交前端 readiness UI 与门禁交互**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate
git add code/frontend/src/api/contracts.ts code/frontend/src/api/admin.ts code/frontend/src/composables/useAdminContestAWD.ts code/frontend/src/composables/useAdminContests.ts code/frontend/src/components/admin/contest/AWDReadinessSummary.vue code/frontend/src/components/admin/contest/AWDReadinessOverrideDialog.vue code/frontend/src/components/admin/contest/AWDOperationsPanel.vue code/frontend/src/views/admin/ContestManage.vue code/frontend/src/api/__tests__/admin.test.ts code/frontend/src/components/admin/__tests__/AWDOperationsPanel.test.ts code/frontend/src/views/admin/__tests__/ContestManage.test.ts
git commit -m "feat(awd): 增加开赛就绪门禁前端交互"
```

## Task 6: 做最小充分验证并整理交付证据

**Files:**
- Modify: `docs/superpowers/plans/2026-04-12-awd-engine-phase8-readiness-gate.md`

- [x] **Step 1: 跑后端 AWD 相关最小回归**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate/code/backend
go test ./internal/module/contest/... -count=1
go test ./internal/app -run TestNewRouterRegistersStudentChallengeRoutes -count=1
```

预期：PASS。

- [x] **Step 2: 跑前端 AWD 管理端最小回归**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts src/components/admin/__tests__/AWDOperationsPanel.test.ts src/components/admin/__tests__/AWDReadinessSummary.test.ts src/composables/__tests__/useAdminContestAWD.test.ts src/views/admin/__tests__/ContestManage.test.ts
npm run typecheck
```

预期：PASS。

- [x] **Step 3: 更新计划执行勾选并提交收尾**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-phase8-readiness-gate
git add -f docs/superpowers/plans/2026-04-12-awd-engine-phase8-readiness-gate.md
git commit -m "docs(awd): 更新 phase8 实施计划状态"
```
