# Frontend Contest Domain Convergence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 收口 `contest` domain 的 feature public API、API client owner 和 guardrail，让平台竞赛相关页面不再依赖活动兼容 barrel 与超级 API client。

**Architecture:** 这轮不做大范围 FSD 重写，只在当前最重的 `contest` 主链路上收口真实 owner。先清空 `features/platform/contests` 兼容入口的运行时 consumer，再把 `api/admin/contests.ts` 按 use case 拆回更窄的 transport owner，最后补防回流 guardrail 和文档事实同步。

**Tech Stack:** Vue 3, Vue Router 4, TypeScript, Vitest, Vite

---

## Objective

- 清退 `features/platform/contests/index.ts` 在运行时主路径中的兼容 owner。
- 把 `api/admin/contests.ts` 的职责按 use case 拆回更窄的 admin contest API client。
- 为 `contest` domain 补一组“不会再回流”的源码级 guardrail。
- 同步更新前端架构台账、architecture facts 和 review/backlog 基线。

## Non-goals

- 不改 contest 相关页面的用户可见行为、文案和权限判定。
- 不顺手重写 `contest-awd-admin`、`contest-workbench`、`contest-awd-config` 的内部 workflow。
- 不在本轮扩展到 `entity` 全量补强；`instance` / `user` entity 收口另开 plan。
- 不把整个 `src/api` 目录一次性重排到新的顶层结构。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `docs/reviews/architecture/2026-06-01-ctf-frontend-architecture-review.md`
- `docs/reviews/architecture/2026-06-01-frontend-sliced-plan-review.md`
- `code/frontend/src/features/platform/contests/index.ts`
- `code/frontend/src/pages/platform/contests/*.vue`
- `code/frontend/src/features/platform/contest-manage/**`
- `code/frontend/src/features/platform/contest-announcements/**`
- `code/frontend/src/features/platform/contest-operations/**`
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/api/admin/index.ts`
- `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/src/features/__tests__/featureBoundaries.test.ts`

## Architecture Fit Check

- 当前代码的主问题不是 route page 太厚，而是 `contest` 这条线拆了目录但还没拆 owner。
- 这份计划只收口 touched surface 上已经确认的 debt：
  - feature compatibility barrel
  - super API client
  - 缺少对应防回流 guardrail
- 计划完成后，`contest` domain 仍会保留多个 feature，但每个 feature 的运行时 consumer、transport owner 和测试基线会更一致，不需要立刻再开第二轮同类重构。

## File Ownership Map

- `code/frontend/src/features/platform/contest-manage/**`
  - 平台竞赛目录、编辑、表单、contest manage route builder 与 save flow owner。
- `code/frontend/src/features/platform/contest-announcements/**`
  - 平台竞赛公告页面与公告数据 owner。
- `code/frontend/src/features/platform/contest-operations/**`
  - 平台竞赛运维页、运维 hub 与 inspector 入口 owner。
- `code/frontend/src/features/platform/contests/index.ts`
  - 当前仅为兼容重导出层；本轮目标是清空 consumer 后删除。
- `code/frontend/src/api/admin/contests.ts`
  - 当前超级 API client；本轮把 use case 按 owner 切回更窄模块。
- `code/frontend/src/api/admin/index.ts`
  - admin API 公共出口；本轮按新拆分模块更新稳定出口。
- `code/frontend/src/pages/platform/contests/*.vue`
  - route entry；应只依赖真实 owning feature public API。
- `code/frontend/src/__tests__/*`
  - 机械 guardrail；本轮补“禁止 compatibility barrel 回流”和“禁止 contest 超级 API client 新增 consumer”的约束。

## Task Breakdown

### Task 1: 先收口 platform contest route entry 到真实 feature public API

**Files:**
- Modify: `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue`
- Modify: `code/frontend/src/pages/platform/contests/ContestEditRoutePage.vue`
- Modify: `code/frontend/src/pages/platform/contests/ContestAnnouncementsRoutePage.vue`
- Modify: `code/frontend/src/pages/platform/contests/ContestOperationsRoutePage.vue`
- Modify: `code/frontend/src/pages/platform/contests/ContestOperationsHubRoutePage.vue`
- Modify: `code/frontend/src/features/platform/contest-manage/model/useContestEditPage.ts`
- Test: `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- Test: `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- Test: `code/frontend/src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts`
- Test: `code/frontend/src/pages/platform/contests/__tests__/ContestOperations.test.ts`
- Test: `code/frontend/src/pages/platform/contests/__tests__/ContestOperationsHub.test.ts`

- [ ] Step 1: 盘点所有 `@/features/platform/contests` 运行时 consumer，并确认每个 consumer 的真实 owning feature。
- [ ] Step 2: 把上述 route page 与 `useContestEditPage.ts` 改为直接依赖 `contest-manage`、`contest-announcements`、`contest-operations` 的根入口。
- [ ] Step 3: 更新 raw-source 测试断言，确保它们不再接受 `@/features/platform/contests` 作为运行时 import。
- [ ] Step 4: 运行聚焦测试，确认 route entry 行为不变。

**Validation:**
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestManage.test.ts src/pages/platform/contests/__tests__/ContestEdit.test.ts src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts src/pages/platform/contests/__tests__/ContestOperations.test.ts src/pages/platform/contests/__tests__/ContestOperationsHub.test.ts`
- `cd code/frontend && rg "@/features/platform/contests" src/pages src/features`

### Task 2: 删除 platform contest compatibility barrel，并补 feature-level 防回流护栏

**Files:**
- Delete or empty-and-retire: `code/frontend/src/features/platform/contests/index.ts`
- Modify: `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
- Modify: `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- Test: `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- Test: `code/frontend/src/features/__tests__/featureBoundaries.test.ts`

- [ ] Step 1: 在 `Task 1` consumer 清零后再次搜索确认运行时代码中已无 `@/features/platform/contests` 依赖。
- [ ] Step 2: 删除兼容 barrel，或先把它改成只保留明确的 `throw new Error` 型临时失败占位用于暴露残留引用。
- [ ] Step 3: 在 feature / architecture guardrail 中补一条约束，禁止新增 `features/<slice>/index.ts` 形式的跨-slice compatibility barrel 回流到主路径。
- [ ] Step 4: 运行架构测试，确认 guardrail 只拦截兼容桶，不误伤真实 feature public API。

**Validation:**
- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/__tests__/featureBoundaries.test.ts`
- `cd code/frontend && rg "@/features/platform/contests" src`

### Task 3: 按 use case 拆分 `api/admin/contests.ts`

**Files:**
- Create: `code/frontend/src/api/admin/contest-manage.ts`
- Create: `code/frontend/src/api/admin/contest-announcements.ts`
- Create: `code/frontend/src/api/admin/contest-operations.ts`
- Create: `code/frontend/src/api/admin/contest-reviews.ts`
- Modify: `code/frontend/src/api/admin/index.ts`
- Modify: `code/frontend/src/features/platform/contest-manage/model/*.ts`
- Modify: `code/frontend/src/features/platform/contest-announcements/model/*.ts`
- Modify: `code/frontend/src/features/platform/contest-operations/model/*.ts`
- Modify: `code/frontend/src/features/contest-awd-admin/model/*.ts`
- Modify: `code/frontend/src/features/contest-awd-config/model/*.ts`
- Modify: `code/frontend/src/features/contest-projector/model/*.ts`
- Modify: `code/frontend/src/features/contest-workbench/model/*.ts`
- Modify or retire: `code/frontend/src/api/admin/contests.ts`
- Test: `code/frontend/src/api/__tests__/admin.test.ts`

- [ ] Step 1: 先按现有函数 consumer 分组，列出每个函数应该落到哪个 API module，避免“按文件体积切”。
- [ ] Step 2: 新建更窄的 admin contest API client，并把 normalize helper 一并迁到对应 owner，而不是复制到多个模块。
- [ ] Step 3: 逐组迁移 feature consumer：
  - `contest-manage`
  - `contest-announcements`
  - `contest-operations`
  - `contest-awd-admin / awd-config / projector / workbench`
- [ ] Step 4: 当运行时 consumer 清零后，删除旧 `api/admin/contests.ts`，或只保留短期兼容出口并立刻补退场 TODO 与护栏。
- [ ] Step 5: 更新 `api/admin/index.ts` 的稳定出口和 API 测试基线。

**Validation:**
- `cd code/frontend && npm run test:run -- src/api/__tests__/admin.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd code/frontend && rg "from '@/api/admin/contests'" src/features`

### Task 4: 给 contest domain 增加 growth / owner guardrail

**Files:**
- Modify: `code/frontend/scripts/frontend-growth-baseline.json`
- Modify: `code/frontend/scripts/check-frontend-growth-guard.mjs`（仅当当前格式无法表达目录级或新增热点文件预算时）
- Modify: `code/frontend/src/__tests__/architectureBoundaries.test.ts`（如需补 source-search 型护栏）
- Modify: `TODO/frontend-sliced-architecture.md`

- [ ] Step 1: 根据当前最大 owner 面，选择最少但有效的 guardrail 落点，不把所有大文件一次性拉进 baseline。
- [ ] Step 2: 至少把 `contest-manage`、`contest-operations`、`contest-awd-admin` 里当前最关键的 page model / page shell 纳入 budget。
- [ ] Step 3: 如果单文件预算不足以覆盖回流风险，再补一条针对 contest domain compatibility import 的源码级断言。
- [ ] Step 4: 运行 growth guard 与架构测试，确认新增预算和断言不会制造大量噪音。

**Validation:**
- `cd code/frontend && npm run check:frontend-growth`
- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/__tests__/featureBoundaries.test.ts`

### Task 5: 同步事实文档、backlog 与 review 基线

**Files:**
- Modify: `TODO/frontend-sliced-architecture.md`
- Modify: `docs/architecture/frontend/01-architecture-overview.md`
- Modify: `docs/architecture/frontend/07-pages-dataflow.md`
- Modify: `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Create or update review archive: `docs/reviews/frontend/` or `docs/reviews/architecture/`

- [ ] Step 1: 把 `contest` domain 收口后的真实 owner 写回 TODO 和 frontend overview，移除已失效的兼容描述。
- [ ] Step 2: 在 pages dataflow 文档里补上 contest route entry 与 owning feature 的最终依赖关系。
- [ ] Step 3: 更新 backlog 优先级，让它反映“contest domain 已收口到哪一层，还剩什么 debt”，而不是继续沿用旧体量判断。
- [ ] Step 4: 归档独立 review，记录这轮收口是否真正关闭了 touched surface 上的已知 debt。

**Validation:**
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## Recommended Execution Order

1. `Task 1`: 先切 route / feature consumer，缩小运行时 blast radius。
2. `Task 2`: 在 consumer 清零后删 compatibility barrel，并立即补 guardrail。
3. `Task 3`: 再拆超级 API client，避免 route 和 transport 两端同时悬空。
4. `Task 4`: 把新的热点 owner 纳入机械预算。
5. `Task 5`: 最后同步事实文档、backlog 和 review 基线。

## Review Focus

- 是否真的清空了 `@/features/platform/contests` 和 `@/api/admin/contests` 的运行时主路径依赖，而不是只换了一个中间别名。
- `api/admin/contests.ts` 的拆分是否按 use case owner 切，而不是把同一套 normalize helper 复制到多个文件。
- guardrail 是否在最小噪音下锁住了 compatibility barrel 和 super API client 回流。
- 文档更新是否写成当前事实，而不是新的候选目录猜想。

## Rollback / Recovery

- 如果 `Task 2` 的 compatibility barrel 删除后暴露出隐藏 consumer，优先回到搜索和 consumer 清理，不恢复长期兼容层。
- 如果 `Task 3` 的 API client 拆分范围过大，优先按 use case 拆成两刀以上，不把全部 contest API 一次改完。
- 如果 growth guard 噪音过高，先收窄到最关键的 2-4 个文件，再逐轮补强，不要为了通过脚本直接放宽全局预算。
