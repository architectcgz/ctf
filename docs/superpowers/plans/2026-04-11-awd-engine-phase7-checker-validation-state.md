# AWD Phase 7 Checker Validation State Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 checker 试跑结果接入题目配置保存链路，形成可持久化、可失效、可回看的 AWD 配置校验状态。

**Architecture:** 继续复用 phase6 的 preview 执行链，不直接把 preview 写入正式业务表。后端增加 Redis 短时 `preview_token` 缓存和 `contest_challenges` 校验状态字段；前端在 `AWDChallengeConfigDialog` 里维护“当前草稿是否仍对应最近一次试跑”的 token 状态，并在 `AWDChallengeConfigPanel` 中展示最近校验状态。

**Tech Stack:** Go, Gin, GORM, Redis, Vue 3, TypeScript, Vue Test Utils, Vitest, existing admin AWD UI

---

### Task 1: 为后端校验状态持久化补 RED 测试

**Files:**
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/contest/application/queries/challenge_service_test.go`

- [ ] **Step 1: 写 preview token 失败测试**

在 `awd_service_test.go` 中扩展 preview 用例，要求：

- `PreviewChecker` 返回 `preview_token`
- token 对应的 preview 结果仍不写 `awd_team_services`

- [ ] **Step 2: 写新增题目消费 token 的失败测试**

在 `challenge_service_test.go` 中补用例，要求：

- 新增 AWD 题目时如果带有效 `awd_checker_preview_token`
- 保存后 `contest_challenges` 会写入：
  - `awd_checker_validation_state`
  - `awd_checker_last_preview_at`
  - `awd_checker_last_preview_result`

- [ ] **Step 3: 写更新题目失效逻辑的失败测试**

在同一测试文件中补用例，要求：

- 已有 `passed` 记录的题目，修改 `awd_checker_type / awd_checker_config` 但不带 token
- 保存后状态变成 `stale`
- 仅修改 `points / order / is_visible / awd_sla_score / awd_defense_score` 时保留原状态

- [ ] **Step 4: 写管理端列表查询失败测试**

在 `challenge_service_test.go`（queries）中补用例，要求管理端列表返回：

- `awd_checker_validation_state`
- `awd_checker_last_preview_at`
- `awd_checker_last_preview_result`

- [ ] **Step 5: 跑后端定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/application/commands -run 'AWDServicePreviewChecker|ChallengeService' -count=1
go test ./internal/module/contest/application/queries -run ChallengeService -count=1
```

预期：FAIL，失败点分别对应 preview token、校验状态字段和状态流转逻辑尚未实现。

### Task 2: 实现后端 preview token 与校验状态存储

**Files:**
- Create: `code/backend/migrations/000055_add_awd_checker_validation_fields_to_contest_challenges.up.sql`
- Create: `code/backend/migrations/000055_add_awd_checker_validation_fields_to_contest_challenges.down.sql`
- Modify: `code/backend/internal/model/contest_challenge.go`
- Modify: `code/backend/internal/dto/awd.go`
- Modify: `code/backend/internal/dto/contest_challenge.go`
- Modify: `code/backend/internal/module/contest/domain/mappers.go`
- Create: `code/backend/internal/module/contest/domain/awd_checker_validation_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- Create: `code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/challenge_add_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/challenge_manage_commands.go`
- Modify: `code/backend/internal/module/contest/infrastructure/challenge_repository.go`
- Modify: `code/backend/internal/module/contest/ports/challenge.go`
- Modify: `code/backend/internal/pkg/redis/keys.go`

- [ ] **Step 1: 增加数据库字段与 DTO**

实现 `contest_challenges` 新字段：

- `awd_checker_validation_state`
- `awd_checker_last_preview_at`
- `awd_checker_last_preview_result`

同步补齐 model 与管理端响应 DTO。

- [ ] **Step 2: 为 preview 结果增加 token 缓存**

在 `PreviewChecker` 里：

- 生成 `preview_token`
- 将 `contest_id / challenge_id / checker_type / checker_config / preview 响应` 写入 Redis
- 继续保持 preview 本身不写正式轮次数据

- [ ] **Step 3: 在新增/更新题目配置时消费 token**

实现：

- 新增题目时消费有效 token，写入 `passed / failed`
- 更新题目时按 checker 配置是否变更决定：
  - 写入新的 `passed / failed`
  - 或转为 `stale / pending`
  - 或保留现有状态

- [ ] **Step 4: 让管理端查询返回校验状态字段**

确保 `ListAdminChallenges` 返回最新的：

- 状态枚举
- 时间
- 最近结果快照

- [ ] **Step 5: 跑后端定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/application/commands -run 'AWDServicePreviewChecker|ChallengeService' -count=1
go test ./internal/module/contest/application/queries -run ChallengeService -count=1
```

预期：PASS。

### Task 3: 为前端校验状态展示补 RED 测试

**Files:**
- Modify: `code/frontend/src/api/__tests__/admin.test.ts`
- Modify: `code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts`
- Modify: `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- Modify: `code/frontend/src/views/platform/__tests__/ContestManage.test.ts`

- [ ] **Step 1: 写 API 契约失败测试**

要求：

- preview 归一化返回 `preview_token`
- 题目列表归一化返回：
  - `awd_checker_validation_state`
  - `awd_checker_last_preview_at`
  - `awd_checker_last_preview_result`
- 新增/更新题目请求可携带 `awd_checker_preview_token`

- [ ] **Step 2: 写配置对话框失败测试**

要求：

- 试跑成功后保存会携带 `awd_checker_preview_token`
- 修改 checker 相关字段后，本地 token 会失效
- 编辑已有题目时能显示最近一次已保存校验结果

- [ ] **Step 3: 写配置面板失败测试**

要求：

- 题目列表展示 `未验证 / 最近通过 / 最近失败 / 待重新验证`
- 行内能看到最近校验时间或最近目标摘要

- [ ] **Step 4: 跑前端定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts src/views/platform/__tests__/ContestManage.test.ts
```

预期：FAIL，失败原因是前端还没有 token 持久化和校验状态展示。

### Task 4: 实现前端校验状态闭环

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/admin.ts`
- Modify: `code/frontend/src/composables/useAdminContestAWD.ts`
- Modify: `code/frontend/src/composables/useAwdCheckResultPresentation.ts`
- Modify: `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
- Modify: `code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue`

- [ ] **Step 1: 接前端 contract 与 admin API**

补：

- `preview_token`
- `AdminContestChallengeData` 的校验状态字段
- 新增/更新题目 payload 里的 `awd_checker_preview_token`

- [ ] **Step 2: 在配置对话框内维护 token 生命周期**

实现：

- 试跑成功后记录 `preview_token`
- 根据当前 `checker_type + checker_config` 生成本地签名
- 只要相关字段变更，就自动清空 token
- 编辑已有题目时显示最近一次已保存校验结果

- [ ] **Step 3: 在题目配置面板展示状态**

要求：

- 每行展示状态标签
- 展示最近校验时间 / 最近目标摘要
- 继续保持现有 flat row 和管理端视觉语义

- [ ] **Step 4: 跑前端定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts src/views/platform/__tests__/ContestManage.test.ts
```

预期：PASS。

### Task 5: 做最小回归并整理提交

**Files:**
- Modify: `docs/architecture/features/awd-checker-validation-state-design.md`
- Modify: `docs/superpowers/plans/2026-04-11-awd-engine-phase7-checker-validation-state.md`

- [ ] **Step 1: 跑后端 AWD 相关最小回归**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/... -count=1
```

预期：PASS。

- [ ] **Step 2: 跑前端 AWD 管理端最小回归**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/frontend
npm run test:run -- src/api/__tests__/admin.test.ts src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts src/views/platform/__tests__/ContestManage.test.ts
npm run typecheck
```

预期：PASS。

- [ ] **Step 3: 更新文档执行状态并提交**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration
git add -f docs/architecture/features/awd-checker-validation-state-design.md docs/superpowers/plans/2026-04-11-awd-engine-phase7-checker-validation-state.md code/backend/migrations/000055_add_awd_checker_validation_fields_to_contest_challenges.up.sql code/backend/migrations/000055_add_awd_checker_validation_fields_to_contest_challenges.down.sql code/backend/internal/model/contest_challenge.go code/backend/internal/dto/awd.go code/backend/internal/dto/contest_challenge.go code/backend/internal/module/contest/domain/mappers.go code/backend/internal/module/contest/domain/awd_checker_validation_support.go code/backend/internal/module/contest/application/commands/awd_service_run_commands.go code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go code/backend/internal/module/contest/application/commands/challenge_add_commands.go code/backend/internal/module/contest/application/commands/challenge_manage_commands.go code/backend/internal/module/contest/infrastructure/challenge_repository.go code/backend/internal/module/contest/ports/challenge.go code/backend/internal/pkg/redis/keys.go code/backend/internal/module/contest/application/commands/awd_service_test.go code/backend/internal/module/contest/application/commands/challenge_service_test.go code/backend/internal/module/contest/application/queries/challenge_service_test.go code/frontend/src/api/contracts.ts code/frontend/src/api/admin.ts code/frontend/src/composables/useAdminContestAWD.ts code/frontend/src/composables/useAwdCheckResultPresentation.ts code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue code/frontend/src/api/__tests__/admin.test.ts code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts code/frontend/src/views/platform/__tests__/ContestManage.test.ts
git commit -m "feat(awd): 增加checker校验状态闭环"
```
