# AWD Phase 10 学员实战工作台 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在学生竞赛详情页内补齐 AWD 实战工作台，让学员可在同一页完成服务启动、目标查看、偷旗提交和实时分数反馈。

**Architecture:** 后端在 contest 模块新增一条用户态 `awd workspace` read model，复用现有 rounds / services / attack logs / service instances 组织学生所需事实；前端继续使用 `ContestDetail` 作为路线壳子，在 AWD 模式下把“题目”面板切成“战场”面板，并通过新的 `ContestAWDWorkspacePanel` 与 `useContestAWDWorkspace` 承接服务目录、目标目录、提交流程和排行榜刷新。

**Tech Stack:** Go, Gin, GORM, Vue 3, TypeScript, Vitest, Vue Test Utils, existing contest/practice APIs, existing scoreboard realtime bridge, journal user shell

## Phase 10 Batch 2 Scope

- 在运行中 AWD 赛事里，每 15 秒自动刷新学生战场数据，并给出同步状态提示。
- 在“我的服务”前增加防守告警摘要，突出 `down / compromised / attack_received > 0` 的题目。
- 在目标目录增加队伍关键字筛选和“仅看可用地址”过滤，保持比赛页内操作闭环。

---

## Execution Notes

- 后端与前端都严格按 `@superpowers:test-driven-development` 先写 RED 测试，再做最小实现。
- 学生竞赛页改造必须显式使用 `@ctf-ui-theme-system`、`@frontend-skill`，保持与现有 student contest / scoreboard / instance 页面一致。
- 测试和类型检查全部在当前 worktree `/home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design` 中串行执行，并使用 `timeout`。
- `docs/superpowers/*` 仍被 `.gitignore` 覆盖；提交计划或 spec 时需要 `git add -f`。

## Planned File Map

### Backend

- Create: `code/backend/internal/dto/contest_awd_workspace.go`
  - 定义学生侧 AWD workspace DTO。
- Create: `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
  - 实现 `GetUserWorkspace`。
- Modify: `code/backend/internal/module/contest/api/http/awd_handler.go`
  - 扩展 query service 接口。
- Create: `code/backend/internal/module/contest/api/http/awd_workspace_handler.go`
  - 提供用户态 workspace handler。
- Modify: `code/backend/internal/app/router_routes.go`
  - 注册 `GET /api/v1/contests/:id/awd/workspace`。
- Modify: `code/backend/internal/app/router_test.go`
  - 断言新路由注册。
- Modify: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
  - 新增 workspace query 测试。

### Frontend

- Modify: `code/frontend/src/api/contracts.ts`
  - 增加学生侧 AWD workspace 数据结构。
- Modify: `code/frontend/src/api/contest.ts`
  - 增加 `getContestAWDWorkspace`、`startContestChallengeInstance`、`submitContestAWDAttack`。
- Create: `code/frontend/src/api/__tests__/contest.test.ts`
  - 覆盖新 API 请求与归一化。
- Create: `code/frontend/src/composables/useContestAWDWorkspace.ts`
  - 管理 workspace 加载、刷新、实例启动、攻击提交、排行榜同步。
- Create: `code/frontend/src/composables/__tests__/useContestAWDWorkspace.test.ts`
  - 覆盖运行态自动刷新与停止刷新。
- Create: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - 渲染学生 AWD 战场主面板。
- Modify: `code/frontend/src/views/contests/ContestDetail.vue`
  - 在 AWD 模式下切到 `battle` 面板并接入新组件。
- Modify: `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
  - 覆盖 AWD 模式展示与交互。

## Task 1: 为后端用户态 AWD workspace 补 RED 测试

**Files:**
- Modify: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- Modify: `code/backend/internal/app/router_test.go`

- [ ] **Step 1: 给 query service 增加已入队 workspace 用例**

新增最小断言：

- `TestAWDServiceGetUserWorkspaceBuildsOwnServicesTargetsAndRecentEvents`

至少覆盖：

- 能拿到 running round
- `my_team` 正确
- `services` 只包含当前队伍
- `targets` 不包含当前队伍
- `recent_events` 只包含与当前队伍相关事件

- [ ] **Step 2: 给 query service 增加未入队不泄露目标目录的用例**

新增：

- `TestAWDServiceGetUserWorkspaceWithoutTeamHidesTargets`

至少断言：

- 不报错
- `my_team == nil`
- `targets` 为空
- `recent_events` 为空

- [ ] **Step 3: 给 router test 增加用户态 workspace 路由断言**

新增：

```go
assertRouteHandlerContains(t, router, "GET", "/api/v1/contests/:id/awd/workspace", "internal/module/contest/api/http")
```

- [ ] **Step 4: 运行定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/contest/application/queries -run GetUserWorkspace -count=1
timeout 120s go test ./internal/app -run TestNewRouterRegistersStudentChallengeRoutes -count=1
```

预期：

- query 包因 DTO / query 方法 / handler 尚不存在而 FAIL
- router test 因路由未注册而 FAIL

## Task 2: 实现后端学生 AWD workspace 查询与路由

**Files:**
- Create: `code/backend/internal/dto/contest_awd_workspace.go`
- Create: `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- Create: `code/backend/internal/module/contest/api/http/awd_workspace_handler.go`
- Modify: `code/backend/internal/module/contest/api/http/awd_handler.go`
- Modify: `code/backend/internal/app/router_routes.go`

- [ ] **Step 1: 先定义 workspace DTO**

包括：

- `ContestAWDWorkspaceResp`
- `ContestAWDWorkspaceTeamResp`
- `ContestAWDWorkspaceServiceResp`
- `ContestAWDWorkspaceTargetTeamResp`
- `ContestAWDWorkspaceTargetServiceResp`
- `ContestAWDWorkspaceRecentEventResp`

- [ ] **Step 2: 在 query service 里实现 `GetUserWorkspace`**

顺序：

1. 校验 AWD contest
2. 解析当前用户队伍
3. 读取当前 running round
4. 组装我的服务目录
5. 在已入队时组装目标目录
6. 在当前轮过滤本队相关攻击事件

- [ ] **Step 3: 暴露 handler 和路由**

新增：

- `GET /api/v1/contests/:id/awd/workspace`

- [ ] **Step 4: 重跑后端定向测试至 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/contest/application/queries -run GetUserWorkspace -count=1
timeout 120s go test ./internal/app -run TestNewRouterRegistersStudentChallengeRoutes -count=1
```

## Task 3: 为前端 contest AWD workspace API 与页面补 RED 测试

**Files:**
- Create: `code/frontend/src/api/__tests__/contest.test.ts`
- Modify: `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`

- [ ] **Step 1: 给 contest API 补 workspace / 实例启动 / AWD 提交红测**

覆盖：

- `getContestAWDWorkspace`
- `startContestChallengeInstance`
- `submitContestAWDAttack`

- [ ] **Step 2: 给 ContestDetail 补 AWD 模式红测**

至少覆盖：

- AWD 赛事显示“战场”页签
- 未入队时显示受限提示
- 已入队时能看到目标目录
- 提交 stolen flag 后显示成功反馈

- [ ] **Step 3: 运行前端定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
timeout 120s npx vitest run src/api/__tests__/contest.test.ts src/views/contests/__tests__/ContestDetail.test.ts
```

## Task 4: 实现前端学生 AWD 战场面板

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/contest.ts`
- Create: `code/frontend/src/composables/useContestAWDWorkspace.ts`
- Create: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- Modify: `code/frontend/src/views/contests/ContestDetail.vue`

- [ ] **Step 1: 先接通 API 与类型**

- [ ] **Step 2: 实现 `useContestAWDWorkspace`**

至少包括：

- workspace 加载
- scoreboard 刷新
- 启动服务
- 提交 stolen flag
- 动作后刷新

- [ ] **Step 3: 实现 `ContestAWDWorkspacePanel`**

布局包含：

- 我的服务目录
- 目标目录与提交区
- 当前轮次摘要
- 实时排行榜
- 最近攻防反馈

- [ ] **Step 4: 在 `ContestDetail.vue` 做 AWD 模式分流**

规则：

- AWD 模式下页签文案显示“战场”
- 只在 AWD 模式渲染新面板
- Jeopardy 模式保持原行为

- [ ] **Step 5: 重跑前端定向测试至 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
timeout 120s npx vitest run src/api/__tests__/contest.test.ts src/views/contests/__tests__/ContestDetail.test.ts
```

## Task 5: 做最小充分回归与阶段收尾

**Files:**
- Modify: `docs/superpowers/plans/2026-04-12-awd-phase10-student-battle-workspace-implementation.md`

- [ ] **Step 1: 运行后端回归**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/contest/application/queries -count=1
timeout 120s go test ./internal/app -run 'TestNewRouter|TestFullRouter' -count=1
```

- [ ] **Step 2: 运行前端回归与类型检查**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
timeout 180s npx vitest run --maxWorkers=1 --testTimeout=20000 src/api/__tests__/contest.test.ts src/views/contests/__tests__/ContestDetail.test.ts src/views/__tests__/pageTabsStyles.test.ts src/views/__tests__/journalUserShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts
timeout 180s npx vue-tsc --noEmit
```

- [ ] **Step 3: 勾掉已完成步骤并整理提交**

建议提交：

```bash
git add -f docs/superpowers/specs/2026-04-12-awd-phase10-student-battle-workspace-design.md \
  docs/superpowers/plans/2026-04-12-awd-phase10-student-battle-workspace-implementation.md
git add code/backend/internal/dto/contest_awd_workspace.go \
  code/backend/internal/module/contest/application/queries/awd_workspace_query.go \
  code/backend/internal/module/contest/api/http/awd_workspace_handler.go \
  code/backend/internal/module/contest/api/http/awd_handler.go \
  code/backend/internal/module/contest/application/queries/awd_service_test.go \
  code/backend/internal/app/router_routes.go \
  code/backend/internal/app/router_test.go \
  code/frontend/src/api/contracts.ts \
  code/frontend/src/api/contest.ts \
  code/frontend/src/api/__tests__/contest.test.ts \
  code/frontend/src/composables/useContestAWDWorkspace.ts \
  code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue \
  code/frontend/src/views/contests/ContestDetail.vue \
  code/frontend/src/views/contests/__tests__/ContestDetail.test.ts
git commit -m "feat(awd): 新增学生实战工作台"
```

## Task 6: 补学生战场第二批交互增强

**Files:**
- Create: `code/frontend/src/composables/__tests__/useContestAWDWorkspace.test.ts`
- Modify: `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
- Modify: `code/frontend/src/composables/useContestAWDWorkspace.ts`
- Modify: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`

- [x] **Step 1: 先补自动刷新与防守告警红测**

覆盖：

- 运行中的 AWD 竞赛会每 15 秒自动刷新一次 workspace 与 scoreboard
- 停止运行后不再继续轮询
- 战场面板会显示防守告警摘要与自动刷新提示

- [x] **Step 2: 补目标目录筛选红测**

至少覆盖：

- 可按队伍关键字过滤目标
- 可切换“仅看可用地址”
- 没有匹配项时给出明确空态

- [x] **Step 3: 在 `useContestAWDWorkspace` 落地自动刷新**

要求：

- 只在 `running / frozen` 状态开启
- 间隔 15 秒
- 卸载后主动清理 timer
- 暴露自动刷新状态与最近同步时间给面板使用

- [x] **Step 4: 在 `ContestAWDWorkspacePanel` 落地防守告警与筛选工具条**

要求：

- 使用现有 `contest-section` / `metric-panel` / `contest-inline-note` 风格
- 不引入 teacher/admin 样式
- 目标筛选输入有显式标签

- [x] **Step 5: 重跑新增前端定向测试至 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
timeout 180s npx vitest run --maxWorkers=1 --testTimeout=20000 src/composables/__tests__/useContestAWDWorkspace.test.ts src/views/contests/__tests__/ContestDetail.test.ts
```
