# AWD Phase 5 Structured Config Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 AWD 题目配置对话框里提供结构化 checker 配置编辑器，替代手写 JSON。

**Architecture:** 不改后端 API，只在前端增加一层 checker 配置解析/构建支持。`AWDChallengeConfigDialog` 根据 checker 类型切换表单区，`legacy_probe` 使用简化字段，`http_standard` 使用动作级结构化表单，并在底部展示只读 JSON 预览。

**Tech Stack:** Vue 3, TypeScript, Vue Test Utils, Vitest, existing admin workspace UI

---

### Task 1: 为结构化 checker 配置编辑补 RED 测试

**Files:**
- Create: `code/frontend/src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts`
- Modify: `code/frontend/src/views/admin/__tests__/ContestManage.test.ts`

- [x] **Step 1: 写失败测试，要求对话框能展示并保存 `http_standard` 结构化字段**

在 `AWDChallengeConfigDialog.test.ts` 中补用例，要求：

- 编辑既有 `http_standard` draft 时能回填 `put_flag / get_flag / havoc`
- 提交时发出嵌套的 `awd_checker_config`
- 底部能看到 JSON 预览

- [x] **Step 2: 写失败测试，要求 `legacy_probe` 使用简化字段保存**

在同一测试文件中补用例，要求：

- 选择 `legacy_probe` 时显示 `health_path`
- 提交时构造 `{ health_path: ... }`

- [x] **Step 3: 更新管理页集成测试为结构化交互**

修改 `ContestManage.test.ts`：

- 创建 `http_standard` 题目时不再直接写 JSON 文本框
- 改为填写结构化字段或使用预置模板
- 断言最终调用 `createAdminContestChallenge` / `updateAdminContestChallenge` 的 payload 不变

- [x] **Step 4: 跑前端定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts src/views/admin/__tests__/ContestManage.test.ts
```

预期：FAIL，失败原因是当前对话框仍只有原始 JSON 文本框。

### Task 2: 实现 checker 配置解析、构建和预置支持

**Files:**
- Create: `code/frontend/src/components/admin/contest/awdCheckerConfigSupport.ts`
- Modify: `code/frontend/src/components/admin/contest/AWDChallengeConfigDialog.vue`

- [x] **Step 1: 新增 checker 配置支持文件**

在 `awdCheckerConfigSupport.ts` 中实现：

- `legacy_probe` 草稿解析与构建
- `http_standard` 草稿解析与构建
- `http_standard` 预置模板
- `headers` JSON 校验与归一化

- [x] **Step 2: 在对话框内接入结构化状态**

让 `AWDChallengeConfigDialog.vue` 在打开时：

- 根据 draft 解析出结构化草稿
- 切换 checker 类型时切换对应表单区
- 底部实时生成 JSON 预览

- [x] **Step 3: 实现字段级校验**

要求：

- `http_standard` 的 `put_flag.path / get_flag.path` 必填
- `expected_status > 0`
- `headers` 仅接受 JSON 对象
- `havoc.path` 为空时不写入结果

- [x] **Step 4: 重新跑定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts src/views/admin/__tests__/ContestManage.test.ts
```

预期：PASS。

### Task 3: 做最小回归并提交

**Files:**
- Modify: `docs/superpowers/plans/2026-04-11-awd-engine-phase5-structured-config-editor.md`

- [x] **Step 1: 跑与 AWD 配置相关的最小充分回归**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts src/components/admin/__tests__/AWDOperationsPanel.test.ts src/views/admin/__tests__/ContestManage.test.ts
```

预期：PASS。

- [x] **Step 2: 跑前端 typecheck**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run typecheck
```

预期：PASS。

- [x] **Step 3: 更新计划文档执行状态**

把已完成步骤勾掉。

- [ ] **Step 4: 提交**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration
git add -f docs/superpowers/specs/2026-04-11-awd-checker-structured-editor-design.md docs/superpowers/plans/2026-04-11-awd-engine-phase5-structured-config-editor.md code/frontend/src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts code/frontend/src/components/admin/contest/awdCheckerConfigSupport.ts code/frontend/src/components/admin/contest/AWDChallengeConfigDialog.vue code/frontend/src/views/admin/__tests__/ContestManage.test.ts
git commit -m "feat(awd): 增加结构化checker配置编辑器"
```
