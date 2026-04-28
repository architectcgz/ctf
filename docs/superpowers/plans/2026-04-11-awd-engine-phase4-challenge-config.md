# AWD Phase 4 Challenge Config Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 AWD 运维页内新增题目配置面板，接通 contest challenge 的 AWD checker 配置增改能力。

**Architecture:** 继续复用 `ContestManage -> AWDOperationsPanel -> useAdminContestAWD` 这条链路，不新增路由。前端补齐 contest challenge 的 AWD 配置契约和新增/更新 API，再在 AWD 运维页内部增加一个局部配置面板和统一编辑对话框。

**Tech Stack:** Vue 3, TypeScript, Vue Test Utils, Vitest, existing admin workspace UI

---

### Task 1: 为 contest challenge AWD 配置补 RED 测试

**Files:**
- Modify: `code/frontend/src/api/__tests__/admin.test.ts`

- [x] **Step 1: 写失败测试，要求 contest challenge 列表归一化 AWD 配置字段**

扩展 `listAdminContestChallenges` 的 mock 数据，要求结果保留：

- `awd_checker_type`
- `awd_checker_config`
- `awd_sla_score`
- `awd_defense_score`

- [x] **Step 2: 写失败测试，要求新增/更新 contest challenge 接口按后端契约发送 AWD 字段**

为 `createAdminContestChallenge` 和 `updateAdminContestChallenge` 各加一个用例，要求请求体包含上述 AWD 配置字段。

- [x] **Step 3: 跑 API 定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts
```

预期：FAIL，失败原因是前端 contract 和 admin API 还没有接 contest challenge 的 AWD 配置能力。

### Task 2: 为 AWD 题目配置面板补 RED 测试

**Files:**
- Modify: `code/frontend/src/views/admin/__tests__/ContestManage.test.ts`
- Modify: `code/frontend/src/components/admin/__tests__/AWDOperationsPanel.test.ts`

- [x] **Step 1: 写失败测试，要求 AWD 运维页可切换到题目配置面板**

要求页面出现：

- `题目配置`
- 已关联题目列表
- `新增题目`
- 行内 `编辑配置`

- [x] **Step 2: 写失败测试，要求可新增和编辑 AWD 题目配置**

通过对话框交互验证：

- 新增时会调用 `createAdminContestChallenge`
- 编辑时会调用 `updateAdminContestChallenge`
- 保存后列表显示新的 checker 配置摘要

- [x] **Step 3: 跑页面定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/views/admin/__tests__/ContestManage.test.ts src/components/admin/__tests__/AWDOperationsPanel.test.ts
```

预期：FAIL，失败原因是 AWD 运维页还没有题目配置面板和配置对话框。

### Task 3: 实现 contest challenge AWD 配置链路

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/admin.ts`
- Modify: `code/frontend/src/composables/useAdminContestAWD.ts`
- Modify: `code/frontend/src/components/admin/contest/AWDOperationsPanel.vue`
- Create: `code/frontend/src/components/admin/contest/AWDChallengeConfigPanel.vue`
- Create: `code/frontend/src/components/admin/contest/AWDChallengeConfigDialog.vue`

- [x] **Step 1: 接入前端 contract 与 admin API**

补：

- `AdminContestChallengeData` 的 AWD 字段
- `createAdminContestChallenge`
- `updateAdminContestChallenge`

- [x] **Step 2: 在 `useAdminContestAWD` 内补题库选择与配置保存逻辑**

补：

- 题库列表读取
- 新增 contest challenge
- 更新 contest challenge
- 保存后刷新 `challengeLinks`

- [x] **Step 3: 在 `AWDOperationsPanel` 内加入局部配置面板切换**

要求：

- 保留现有 `轮次态势`
- 新增 `题目配置`
- 切换符合现有后台 tab 语义

- [x] **Step 4: 实现配置目录与统一编辑对话框**

要求：

- 列出已关联题目
- 显示 checker 类型、SLA 分、防守分和 JSON 摘要
- 支持新增和编辑
- 视觉风格对齐当前后台工作台

- [x] **Step 5: 跑定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/admin/__tests__/AWDOperationsPanel.test.ts src/views/admin/__tests__/ContestManage.test.ts
```

预期：PASS。

### Task 4: 做最小验证并提交

**Files:**
- Modify: `docs/superpowers/plans/2026-04-11-awd-engine-phase4-challenge-config.md`

- [x] **Step 1: 跑前端 typecheck**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run typecheck
```

预期：PASS。

- [x] **Step 2: 更新计划文档执行状态**

把已完成步骤勾掉。

- [ ] **Step 3: 提交**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration
git add -f docs/architecture/features/2026-04-11-awd-challenge-config-design.md docs/superpowers/plans/2026-04-11-awd-engine-phase4-challenge-config.md code/frontend/src/api/contracts.ts code/frontend/src/api/admin.ts code/frontend/src/api/__tests__/admin.test.ts code/frontend/src/composables/useAdminContestAWD.ts code/frontend/src/components/admin/contest/AWDOperationsPanel.vue code/frontend/src/components/admin/contest/AWDChallengeConfigPanel.vue code/frontend/src/components/admin/contest/AWDChallengeConfigDialog.vue code/frontend/src/components/admin/__tests__/AWDOperationsPanel.test.ts code/frontend/src/views/admin/__tests__/ContestManage.test.ts
git commit -m "feat(awd): 增加题目配置面板"
```
