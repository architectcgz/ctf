# AWD 防守内容页 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让学生在 AWD 战场页点击某个服务的“防守”按钮后，进入独立的只读防守内容页，左侧浏览防守文件，右侧查看文件内容。

**Architecture:** 路由新增 `ContestAWDDefenseWorkbench` 独立页，页面 owner 使用 `useContestAwdDefenseWorkbenchPage` 持有 route params、workspace 校验、目录/文件请求和 stale-response 保护；后端仅恢复 `GET defense/directories` 与 `GET defense/files`，继续禁用写文件和命令执行，并把敏感目录项过滤收敛在 runtime adapter。

**Tech Stack:** Vue 3 + Vite + Vitest + Vue Router；Go + Gin；现有 `contest-awd-workspace` feature、runtime module、AWD 只读工作台组件。

---

## Source Docs

- `docs/architecture/features/awd-defense-content-page-design.md`
- `docs/architecture/features/awd-web-defense-workbench-design.md`

## Task 1: 恢复后端只读防守文件接口

**Files:**
- Modify: `code/backend/internal/module/runtime/api/http/handler.go`
- Modify: `code/backend/internal/module/runtime/api/http/handler_test.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters_test.go`

- [ ] **Step 1: 先写 handler 行为测试，锁定 read/list 放开、save/command 继续 forbidden**

测试目标：

- `GET /defense/directories` 调用 service 并返回目录内容
- `GET /defense/files` 调用 service 并返回文件内容
- `GET /defense/files?path=.env`、`GET /defense/files?path=.ssh/id_rsa`、`GET /defense/files?path=../app.py` 必须失败
- `PUT /defense/files` 继续返回 `403`
- `POST /defense/commands` 继续返回 `403`

Run:

```bash
cd code/backend && go test ./internal/module/runtime/api/http -run AWDDefenseWorkbench -count=1
```

Expected:

- 当前失败，因为 handler 还在对 read/list 直接返回 forbidden

- [ ] **Step 2: 对齐 runtime HTTP adapter 的只读开关、根目录约束和敏感路径过滤**

实现要求：

- 为 `newRuntimeHTTPServiceAdapter` 增加 `defenseWorkbenchReadOnlyEnabled`、`defenseWorkbenchRoot`
- `ReadAWDDefenseFile` / `ListAWDDefenseDirectory` 只在只读开关开启时可用
- 目录和文件路径继续做 normalize
- 目录项过滤 `.env`、`.env.*`、`.ssh`、`id_rsa`、`id_ed25519`、`authorized_keys`、`known_hosts`
- 直接猜路径读取敏感文件必须被后端拒绝，不能只靠目录列表隐藏
- rooted path 继续落在 `container.defense_workbench_root`

- [ ] **Step 3: 放开 handler 的 read/list 路径，保持 save/command 继续 forbidden**

实现要求：

- `ReadAWDDefenseFile`、`ListAWDDefenseDirectory` 改为调用 service
- `SaveAWDDefenseFile`、`RunAWDDefenseCommand` 保持当前 forbidden 语义不变

- [ ] **Step 4: 跑后端最小验证**

Run:

```bash
cd code/backend && go test ./internal/module/runtime/api/http ./internal/module/runtime/runtime -count=1
```

Expected:

- handler 测试通过
- adapter 测试通过

## Task 2: 补齐前端防守内容页的数据入口

**Files:**
- Modify: `code/frontend/src/api/contest.ts`
- Modify: `code/frontend/src/api/__tests__/contest.test.ts`
- Modify: `code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.ts`
- Create: `code/frontend/src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.test.ts`
- Modify: `code/frontend/src/features/contest-awd-workspace/model/index.ts`
- Modify: `code/frontend/src/features/contest-awd-workspace/index.ts`

- [ ] **Step 1: 先补 API client 测试，锁定目录/文件读取请求形状**

测试目标：

- `requestContestAWDDefenseDirectory(contestId, serviceId, path)` 使用 `GET /defense/directories?path=...`
- `requestContestAWDDefenseFile(contestId, serviceId, path)` 使用 `GET /defense/files?path=...`

Run:

```bash
cd code/frontend && npm run test:run -- src/api/__tests__/contest.test.ts
```

Expected:

- 当前失败，因为 API client 还不存在

- [ ] **Step 2: 扩展 `useContestAwdDefenseWorkbenchPage` 为页面 owner**

实现要求：

- 读取 `contestId`、`serviceId`
- 生成 `backLink`
- 加载 workspace 并校验 `serviceId` 属于当前用户自己的 AWD 服务
- 首次进入页面后自动请求根目录 `.`
- 持有 `directory`、`file`、`loading`、`error`
- 在切目录/切文件时做 sequence guard，晚到响应不能覆盖当前状态
- 切目录时先清空旧文件内容
- `backLink` 必须返回 `/contests/:id?panel=challenges`

- [ ] **Step 3: 跑前端 API 与页面模型最小验证**

Run:

```bash
cd code/frontend && npm run test:run -- src/api/__tests__/contest.test.ts src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.test.ts src/views/contests/__tests__/ContestAWDDefenseWorkbench.test.ts
```

Expected:

- API 请求测试通过
- 页面模型测试覆盖自动加载根目录、返回战场链接和 stale response

## Task 3: 接入独立路由与页面 UI

**Files:**
- Modify: `code/frontend/src/router/routes/studentRoutes.ts`
- Modify: `code/frontend/src/views/contests/ContestAWDDefenseWorkbench.vue`
- Modify: `code/frontend/src/views/contests/__tests__/ContestAWDDefenseWorkbench.test.ts`
- Modify: `code/frontend/src/components/contests/awd/AWDDefenseFileWorkbench.vue`
- Modify: `code/frontend/src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts`
- Modify: `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- Modify: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- Modify: `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- Modify: `code/frontend/src/components/layout/__tests__/AppLayout.test.ts`

- [ ] **Step 1: 先写页面与战场入口测试，锁定独立页路由和“防守”按钮**

测试目标：

- 学生路由包含 `ContestAWDDefenseWorkbench`
- `ContestAWDWorkspacePanel` 服务操作区出现“防守”按钮或对应 emit
- 点击“防守”后 `router.push` 到 `ContestAWDDefenseWorkbench` 且携带正确 `contestId/serviceId`
- 独立页不再显示“防守入口已迁移”的占位文案
- 独立页进入后会触发根目录 `.` 加载

Run:

```bash
cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestAWDDefenseWorkbench.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/components/layout/__tests__/AppLayout.test.ts
```

Expected:

- 当前失败，因为路由和入口都还没接上

- [ ] **Step 2: 把占位页替换成真实防守内容页**

实现要求：

- 顶部显示返回战场、服务标题、当前目录、刷新
- 页面主区域为双栏
- 左侧目录点击、右侧文件内容查看
- 403 / serviceId 无效 / 空目录 / 未选中文件要有独立状态
- 保持暗色战场风格，不引入说明性 UI 文案
- `AWDDefenseFileWorkbench.vue` 继续只做 props/emits，不直接 import `@/api/contest` 或 `vue-router`

- [ ] **Step 3: 在战场服务卡中新增“防守”入口并路由跳转**

实现要求：

- `AWDDefenseServiceList` 增加 `open-defense` emit
- `ContestAWDWorkspacePanel` 负责 `router.push`
- 仍保持当前 source test 对“战场页不直接读取防守文件 API”的约束

- [ ] **Step 4: 跑前端页面最小验证**

Run:

```bash
cd code/frontend && npm run test:run -- src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts src/views/contests/__tests__/ContestAWDDefenseWorkbench.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/components/layout/__tests__/AppLayout.test.ts
```

Expected:

- 页面行为与 source test 通过

## Task 4: 集成验证与完成门禁

**Files:**
- Modify: `docs/reviews/general/2026-05-05-awd-defense-content-page-review.md`

- [ ] **Step 1: 跑本次改动的最小充分验证**

Run:

```bash
cd code/backend && go test ./internal/module/runtime/api/http ./internal/module/runtime/runtime -count=1
cd code/frontend && npm run test:run -- src/api/__tests__/contest.test.ts src/features/contest-awd-workspace/model/useContestAwdDefenseWorkbenchPage.test.ts src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts src/views/contests/__tests__/ContestAWDDefenseWorkbench.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/components/layout/__tests__/AppLayout.test.ts
cd code/frontend && npm run typecheck
```

Expected:

- 后端 targeted tests 通过
- 前端 targeted tests 通过
- typecheck 通过

- [ ] **Step 2: 做计划自审**

检查点：

- 页面 owner 是否仍在独立页而不是展示组件里
- 战场页是否只保留入口而不重新耦合文件请求
- 后端是否只恢复 read/list，没有顺手放开 save/command
- 目录过滤是否后端兜底

- [ ] **Step 3: 归档独立 review 结论**

输出：

- `docs/reviews/general/2026-05-05-awd-defense-content-page-review.md`

Review focus：

- 路由 owner 是否清晰
- stale response 是否被真正处理
- 学生是否会看到敏感路径或旧占位文案
- “防守”入口是否真的落到独立内容页
- 后端是否拒绝手工构造的敏感路径读取

## Plan Self-Review

- 架构边界明确：战场页只做入口，独立页才持有目录/文件请求和展示状态。
- 复用点明确：继续复用现有 `AWDDefenseFileWorkbench.vue`、`ContestAWDDefenseWorkbench.vue`、`useContestAwdDefenseWorkbenchPage.ts`，避免额外引入第二套防守 UI。
- 结构收敛明确：本计划不是只把占位文案改成页面样子，而是把“占位页、现有组件、只读后端接口”三块能力真正接起来。
- 风险明确：runtime module 现有 adapter 与 composition 里同名 adapter 已经出现行为漂移，实现时要以 production 实际接线的 `internal/module/runtime/runtime` 路径为主，避免只改测试用 compat 层。
