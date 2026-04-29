# AWD Engine Phase 3 Inspector Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让后台 AWD 面板与前端 API 适配新的 checker 结果结构，正确展示 `sla / attack / defense` 三段分，并把 `http_standard` 的 `put_flag / get_flag / havoc` 结果展示出来。

**Architecture:** 这轮优先做前端适配，不额外扩散后端改造范围。复用 phase1/phase2 已经落下的接口字段，在 `frontend/src/api` 完成新字段归一化，在 `useAwdCheckResultPresentation` 与 `AWDRoundInspector` 内把旧的“探活视图”升级为“checker 结果视图”，并同步更新导出内容。

**Tech Stack:** Vue 3, TypeScript, Vitest, existing admin AWD inspector UI

---

### Task 1: 为前端 AWD API 归一化补 RED 测试

**Files:**
- Modify: `code/frontend/src/api/__tests__/admin.test.ts`

- [x] **Step 1: 写失败测试，要求服务记录归一化 `checker_type` 与 `sla_score`**

扩展 AWD 轮次巡检接口与服务列表接口的 mock 数据，要求归一化结果包含：
- `checker_type`
- `sla_score`

- [x] **Step 2: 写失败测试，要求轮次 summary item 归一化 `sla_score`**

扩展 AWD 轮次汇总接口 mock 数据，要求归一化后的 `items` 保留 `sla_score`。

- [x] **Step 3: 跑前端 API 定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts
```

预期：FAIL，失败原因是 contracts 与 normalize 逻辑尚未接入新字段。

### Task 2: 为 checker 结果展示补 RED 测试

**Files:**
- Create: `code/frontend/src/composables/__tests__/useAwdCheckResultPresentation.test.ts`

- [x] **Step 1: 写失败测试，要求能识别 `http_standard` checker 标签与动作摘要**

对 `useAwdCheckResultPresentation` 增加用例，要求：
- `checker_type = http_standard` 时能返回可读标签
- `summarizeCheckResult` 能带出 checker 类型与状态原因
- 能读取 `put_flag / get_flag / havoc` 动作结果

- [x] **Step 2: 跑 composable 测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/composables/__tests__/useAwdCheckResultPresentation.test.ts
```

预期：FAIL，失败原因是当前展示层只支持 legacy probe 结构。

### Task 3: 实现 AWD inspector 的 checker 结果视图

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/admin.ts`
- Modify: `code/frontend/src/composables/useAwdCheckResultPresentation.ts`
- Modify: `code/frontend/src/composables/useAwdInspectorExports.ts`
- Modify: `code/frontend/src/components/platform/contest/AWDRoundInspector.vue`

- [x] **Step 1: 接入前端 contracts 与 normalize 新字段**

让前端保留并输出：
- `AWDTeamServiceData.checker_type`
- `AWDTeamServiceData.sla_score`
- `AWDRoundSummaryItemData.sla_score`

- [x] **Step 2: 扩展 checker 结果展示 helper**

补充：
- checker 类型标签
- `flag_mismatch / invalid_checker_config / flag_unavailable` 等状态文案
- `put_flag / get_flag / havoc` 动作解析

- [x] **Step 3: 升级 AWD inspector 汇总与服务表**

要求：
- 本轮汇总表展示 `SLA / 攻击 / 防守`
- 服务表得分列展示 `SLA / 防守 / 攻击`
- 检查结果区显示 checker 类型、动作摘要、target 级 action 明细
- legacy probe 旧结果结构继续可读

- [x] **Step 4: 同步服务导出内容**

导出 CSV 时补：
- `Checker类型`
- `SLA得分`
- `攻击得分`

- [x] **Step 5: 跑前端定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/composables/__tests__/useAwdCheckResultPresentation.test.ts src/views/platform/__tests__/ContestManage.test.ts
```

预期：PASS，说明 API 归一化、展示 helper 和后台 AWD 面板已兼容 checker 结果结构。

### Task 4: 做最小验证并提交

**Files:**
- Modify: `docs/superpowers/plans/2026-04-11-awd-engine-phase3-inspector.md`

- [x] **Step 1: 跑前端 typecheck**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run typecheck
```

预期：PASS。

- [x] **Step 2: 更新计划文档执行状态**

把已完成步骤勾掉。

- [x] **Step 3: 提交**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration
git add -f docs/superpowers/plans/2026-04-11-awd-engine-phase3-inspector.md code/frontend/src/api/contracts.ts code/frontend/src/api/admin.ts code/frontend/src/api/__tests__/admin.test.ts code/frontend/src/composables/useAwdCheckResultPresentation.ts code/frontend/src/composables/__tests__/useAwdCheckResultPresentation.test.ts code/frontend/src/composables/useAwdInspectorExports.ts code/frontend/src/components/platform/contest/AWDRoundInspector.vue code/frontend/src/views/platform/__tests__/ContestManage.test.ts
git commit -m "feat(awd): 升级后台巡检结果视图"
```
