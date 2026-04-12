# AWD Phase 6 Checker Preview Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 AWD 题目配置对话框里增加真实 checker 试跑能力，帮助管理员在保存前验证配置是否可用。

**Architecture:** 后端在现有 `AWDRoundUpdater` 上增加一个不写库的 preview 执行入口，复用 `legacy_probe / http_standard` checker 运行逻辑，只返回临时结果。前端继续使用现有 `AWDChallengeConfigDialog`，在不新增路由的前提下补试跑输入区和结果区，并保持 CTF 管理端 UI 语言一致。

**Tech Stack:** Go, Gin, GORM, Vue 3, TypeScript, Vue Test Utils, Vitest, existing admin workspace UI

---

### Task 1: 为 checker preview 补 RED 测试

**Files:**
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/frontend/src/api/__tests__/admin.test.ts`
- Modify: `code/frontend/src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts`

- [x] **Step 1: 写后端命令失败测试**

在 `awd_service_test.go` 中补用例，要求：

- `PreviewChecker` 能对 `http_standard` 执行真实请求并返回 `service_status / check_result`
- preview 结束后数据库里没有新增 `awd_team_services` 记录

- [x] **Step 2: 写路由注册失败测试**

在 `router_test.go` 中补断言：

- `POST /api/v1/admin/contests/:id/awd/checker-preview` 已注册到 `internal/module/contest/api/http`

- [x] **Step 3: 写前端 API 契约失败测试**

在 `admin.test.ts` 中补用例，要求：

- 新增 `runContestAWDCheckerPreview`
- 请求路径、请求体字段和返回值归一化都符合新契约

- [x] **Step 4: 写配置对话框失败测试**

在 `AWDChallengeConfigDialog.test.ts` 中补用例，要求：

- 填写 `access_url` 后点击“试跑 Checker”
- 组件发出 preview 请求或展示 preview 结果
- 返回结果后出现状态摘要和动作明细

- [x] **Step 5: 跑定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/application/commands ./internal/app -count=1
```

以及：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts
```

预期：FAIL，失败原因是 preview 接口和对话框交互尚未实现。

### Task 2: 实现后端 checker preview 执行链

**Files:**
- Modify: `code/backend/internal/dto/awd.go`
- Modify: `code/backend/internal/module/contest/api/http/awd_handler.go`
- Modify: `code/backend/internal/module/contest/api/http/awd_round_check_handler.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- Create: `code/backend/internal/module/contest/application/commands/awd_checker_preview_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- Modify: `code/backend/internal/module/contest/domain/awd_source_support.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/app/router_routes.go`

- [x] **Step 1: 加 DTO 和 handler 契约**

新增 preview request/response DTO，并在 `AWDHandler` 上暴露 `PreviewChecker`。

- [x] **Step 2: 扩展 command service**

在 `AWDService` 中新增 preview 命令：

- 校验 AWD 赛事
- 校验题目存在
- 复用 `validateAndNormalizeContestAWDFields`
- 组装 preview 上下文

- [x] **Step 3: 在 round manager 上实现不写库 preview**

新增 `PreviewServiceCheck` 接口，实现：

- `legacy_probe` 走现有探活逻辑
- `http_standard` 走现有 checker 动作链
- 结果 `check_source = checker_preview`
- 不调用任何写库、缓存刷新和分数重算逻辑

- [x] **Step 4: 补路由**

在 `router_routes.go` 注册：

```bash
POST /api/v1/admin/contests/:id/awd/checker-preview
```

- [x] **Step 5: 跑后端定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/application/commands ./internal/app -count=1
```

预期：PASS。

### Task 3: 实现前端试跑入口与结果展示

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/admin.ts`
- Modify: `code/frontend/src/components/admin/contest/AWDChallengeConfigDialog.vue`
- Modify: `code/frontend/src/composables/useAwdCheckResultPresentation.ts`
- Modify: `code/frontend/src/components/admin/contest/AWDOperationsPanel.vue`

- [x] **Step 1: 接前端 API 封装**

在 `admin.ts` 中新增 `runContestAWDCheckerPreview`，并补对应 contract 类型。

- [x] **Step 2: 在对话框中增加 preview 状态**

新增本地状态：

- `preview_form.access_url`
- `preview_form.preview_flag`
- `previewing`
- `preview_result`
- `preview_error`

- [x] **Step 3: 实现试跑交互与结果区**

要求：

- 点击按钮前先复用现有 checker 配置校验
- 请求成功后展示状态摘要、动作明细、目标摘要和 JSON 预览
- 请求失败后显示错误提示，但不清空表单

- [x] **Step 4: 扩展结果标签语义**

让 `useAwdCheckResultPresentation.ts` 支持 `checker_preview` 来源标签，避免 dialog 自己重复维护一套映射。

- [x] **Step 5: 跑前端定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts
```

预期：PASS。

### Task 4: 做最小回归并更新计划状态

**Files:**
- Modify: `docs/superpowers/plans/2026-04-11-awd-engine-phase6-checker-preview.md`

- [x] **Step 1: 跑后端 AWD 相关回归**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/... -count=1
```

预期：PASS。

- [x] **Step 2: 跑前端 AWD 管理端相关回归**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts src/components/admin/__tests__/AWDOperationsPanel.test.ts src/views/admin/__tests__/ContestManage.test.ts
```

预期：PASS。

- [x] **Step 3: 跑前端 typecheck**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run typecheck
```

预期：PASS。

- [x] **Step 4: 更新计划文档执行状态**

把已完成步骤勾掉。

- [ ] **Step 5: 提交**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration
git add -f docs/superpowers/specs/2026-04-11-awd-checker-preview-design.md docs/superpowers/plans/2026-04-11-awd-engine-phase6-checker-preview.md code/backend/internal/dto/awd.go code/backend/internal/module/contest/api/http/awd_handler.go code/backend/internal/module/contest/api/http/awd_round_check_handler.go code/backend/internal/module/contest/application/commands/awd_checker_preview_support.go code/backend/internal/module/contest/application/commands/awd_service_run_commands.go code/backend/internal/module/contest/application/commands/awd_service_test.go code/backend/internal/module/contest/application/jobs/awd_checker_preview.go code/backend/internal/module/contest/application/jobs/awd_round_updater.go code/backend/internal/module/contest/domain/awd_source_support.go code/backend/internal/module/contest/ports/awd.go code/backend/internal/app/router_routes.go code/backend/internal/app/router_test.go code/frontend/src/api/contracts.ts code/frontend/src/api/admin.ts code/frontend/src/api/__tests__/admin.test.ts code/frontend/src/components/admin/contest/AWDChallengeConfigDialog.vue code/frontend/src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts code/frontend/src/composables/useAwdCheckResultPresentation.ts docs/superpowers/plans/2026-04-11-awd-engine-phase6-checker-preview.md
git commit -m "feat(awd): 增加checker试跑能力"
```
