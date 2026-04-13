# AWD Phase 11 能力画像回流 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 AWD 学生成功攻击结果能进入个人能力画像与推荐口径，补齐攻防实战到技能评估的个人闭环。

**Architecture:** 在 `awd_attack_logs` 上新增 `submitted_by_user_id` 记录学生真实提交人；在 contest 模块发布 `contest.awd.attack_accepted` 事件；assessment 继续复用现有事件驱动更新链路，把练习正确提交与 AWD 首次成功攻击按 `challenge_id` 去重并集后作为能力画像与推荐的统一事实源。

**Tech Stack:** Go, GORM, SQL migrations, in-memory event bus, assessment module, contest module, Go test

## Execution Notes

- 先补 RED 测试，再写最小实现。
- 这轮默认只改后端，不改 student UI。
- 所有验证继续在当前 worktree `/home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design` 内串行执行，并加 `timeout`。
- `docs/superpowers/*` 仍被 `.gitignore` 覆盖；提交时要 `git add -f`。

## Planned File Map

### Backend

- Create: `code/backend/migrations/000056_add_submitter_to_awd_attack_logs.up.sql`
- Create: `code/backend/migrations/000056_add_submitter_to_awd_attack_logs.down.sql`
- Modify: `code/backend/internal/model/awd.go`
- Create: `code/backend/internal/module/contest/contracts/events.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_attack_submit_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- Modify: `code/backend/internal/app/composition/contest_module.go`
- Modify: `code/backend/internal/app/composition/assessment_module.go`
- Modify: `code/backend/internal/module/assessment/application/commands/profile_service.go`
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- Modify: `code/backend/internal/module/assessment/infrastructure/repository.go`

### Tests

- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`

## Task 1: 为 AWD 个人归因补红测

**Files:**
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`

- [x] **Step 1: 补日志写入提交人的用例**

新增：

- `TestAWDServiceSubmitAttackPersistsSubmittedByUserID`

至少断言：

- 提交成功后 `awd_attack_logs.submitted_by_user_id == 当前 user_id`
- 管理员手工写攻击日志不写该字段

- [x] **Step 2: 补 AWD 成功事件发布用例**

新增：

- `TestAWDServiceSubmitAttackPublishesAttackAcceptedEvent`

至少断言：

- 只在 `score_gained > 0` 的成功提交时发布
- 事件负载包含 `user_id / challenge_id / dimension`

- [x] **Step 3: 运行定向命令确认 RED**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/contest/application/commands -run 'SubmitAttack.*SubmittedBy|SubmitAttack.*PublishesAttackAcceptedEvent' -count=1
```

## Task 2: 为 assessment AWD 画像口径补红测

**Files:**
- Modify: `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`

- [x] **Step 1: 补 AWD 首次成功攻击进入画像的用例**

新增：

- `TestCalculateSkillProfileWithContextCountsSuccessfulAWDAttacks`

至少断言：

- AWD 首次成功攻击会计入维度得分
- 同一题练习 + AWD 成功只算一次

- [x] **Step 2: 补推荐排除集合合并 AWD 证据的用例**

新增：

- `TestRecommendationServiceRecommendChallengesIncludesAWDSolvedIDs`

至少断言：

- `excludeSolved` 同时包含练习正确题和 AWD 已证明题

- [x] **Step 3: 补 AWD 事件驱动缓存刷新/画像更新用例**

至少覆盖：

- profile service 可消费 `contest.awd.attack_accepted`
- recommendation service 可清理该用户缓存

- [x] **Step 4: 运行定向命令确认 RED**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/assessment/application/commands -run 'SkillProfile.*AWD|AttackAccepted' -count=1
timeout 120s go test ./internal/module/assessment/application/queries -run 'Recommend.*AWD|AttackAccepted' -count=1
```

## Task 3: 实现 AWD 个人归因与 assessment 口径

**Files:**
- Create: `code/backend/migrations/000056_add_submitter_to_awd_attack_logs.up.sql`
- Create: `code/backend/migrations/000056_add_submitter_to_awd_attack_logs.down.sql`
- Modify: `code/backend/internal/model/awd.go`
- Create: `code/backend/internal/module/contest/contracts/events.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_attack_submit_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- Modify: `code/backend/internal/app/composition/contest_module.go`
- Modify: `code/backend/internal/app/composition/assessment_module.go`
- Modify: `code/backend/internal/module/assessment/application/commands/profile_service.go`
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- Modify: `code/backend/internal/module/assessment/infrastructure/repository.go`

- [x] **Step 1: 扩展 `awd_attack_logs` 与模型**

- [x] **Step 2: 在 AWD 提交链路写入 `submitted_by_user_id` 并发布事件**

- [x] **Step 3: 在 assessment 注册 AWD 事件消费者**

- [x] **Step 4: 调整画像与推荐查询口径，合并 AWD 首次成功攻击**

- [x] **Step 5: 重跑定向测试至 GREEN**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/contest/application/commands -run 'SubmitAttack.*SubmittedBy|SubmitAttack.*PublishesAttackAcceptedEvent' -count=1
timeout 120s go test ./internal/module/assessment/application/commands -run 'SkillProfile.*AWD|AttackAccepted' -count=1
timeout 120s go test ./internal/module/assessment/application/queries -run 'Recommend.*AWD|AttackAccepted' -count=1
```

## Task 4: 做最小充分回归

- [x] **Step 1: 跑 contest 与 assessment 相关回归**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/contest/application/commands -count=1
timeout 120s go test ./internal/module/assessment/application/commands -count=1
timeout 120s go test ./internal/module/assessment/application/queries -count=1
timeout 120s go test ./internal/app -run 'TestNewRouter|TestFullRouter' -count=1
```
