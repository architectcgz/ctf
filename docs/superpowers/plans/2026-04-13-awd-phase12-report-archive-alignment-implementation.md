# AWD Phase 12 报告与复盘归档口径对齐 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让个人报告、班级统计和学生复盘归档与 phase11 的 AWD 个人攻击事实源对齐，既能正确计分，也能显示真实攻防过程。

**Architecture:** 在 `report_repository.go` 内统一“个人已证明题目”与“个人攻击活动”两类事实源，分别驱动 solved/score/rank 和 attempts/timeline/evidence 查询；同时补 `report_service.go` 的复盘摘要与教学观察识别 AWD 事件，保证统计、证据链、观察三者一致。

**Tech Stack:** Go, GORM, raw SQL, SQLite integration tests, assessment module, Go test

---

## Execution Notes

- 先写 RED 测试，再写最小实现。
- phase12 只改后端，不改教师复盘页前端。
- 所有验证继续在当前 worktree `/home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design` 内串行执行，并加 `timeout`。
- `docs/superpowers/*` 仍被 `.gitignore` 覆盖；提交时要 `git add -f`。

## Planned File Map

### Docs

- Create: `docs/architecture/features/2026-04-13-awd-phase12-report-archive-alignment-design.md`
- Create: `docs/superpowers/plans/2026-04-13-awd-phase12-report-archive-alignment-implementation.md`

### Backend

- Modify: `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- Create: `code/backend/internal/module/assessment/infrastructure/report_repository_test.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service_test.go`

## Task 1: 为报告仓库补 AWD 对齐红测

**Files:**
- Create: `code/backend/internal/module/assessment/infrastructure/report_repository_test.go`

- [ ] **Step 1: 建 SQLite 测试基座**

准备最小 schema：

- `users`
- `challenges`
- `submissions`
- `awd_rounds`
- `awd_attack_logs`
- `teams`
- `instances`
- `audit_logs`

并提供种子辅助函数，便于后续分别验证统计、时间线、证据链。

- [ ] **Step 2: 先写个人与班级统计红测**

新增至少这些用例：

- `TestReportRepositoryGetPersonalStatsIncludesAWDSolvedAndAttempts`
- `TestReportRepositoryListPersonalDimensionStatsDedupesPracticeAndAWD`
- `TestReportRepositoryClassStatsIncludeAWDSolvedEvidence`

至少断言：

- AWD 成功攻击进入 `total_score / total_solved / rank`
- AWD 攻击日志进入 `total_attempts`
- 同题练习 + AWD 成功只算一次
- 班级平均分与 Top 榜反映 AWD 成功攻击

- [ ] **Step 3: 先写时间线与证据链红测**

新增至少这些用例：

- `TestReportRepositoryGetStudentTimelineIncludesAWDAttackEvents`
- `TestReportRepositoryGetStudentEvidenceIncludesAWDAttackLogs`

至少断言：

- 返回 `awd_attack_submit` 时间线事件
- 事件 detail 带目标队伍与成功/失败结果
- 证据链包含 `awd_attack_submission`
- meta 带 `is_success / score_gained / round_id / victim_team_name`

- [ ] **Step 4: 运行仓库层定向测试确认 RED**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/assessment/infrastructure -run 'ReportRepository.*AWD' -count=1
```

## Task 2: 为复盘摘要与观察补 AWD 红测

**Files:**
- Modify: `code/backend/internal/module/assessment/application/commands/report_service_test.go`

- [ ] **Step 1: 补 AWD 事件进入复盘摘要的用例**

新增：

- `TestBuildReviewArchiveSummaryCountsAWDAttackEvents`

至少断言：

- `CorrectSubmissionCount` 会统计成功 `awd_attack_submit`
- `LastActivityAt` 仍按全量事件计算

- [ ] **Step 2: 补 AWD 事件驱动教学观察的用例**

新增：

- `TestBuildReviewArchiveObservationsTreatsAWDAttacksAsHandsOnEvidence`

至少断言：

- `awd_attack_submission` 会被视为实操证据
- 连续 AWD 失败攻击会触发错误提交观察分支

- [ ] **Step 3: 运行定向命令确认 RED**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/assessment/application/commands -run 'ReviewArchive.*AWD|BuildReviewArchive.*AWD' -count=1
```

## Task 3: 实现报告与复盘归档口径统一

**Files:**
- Modify: `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service.go`

- [ ] **Step 1: 在 `report_repository.go` 抽出统一 solved/activity 查询片段**

至少覆盖：

- 个人 solved challenge 集合
- 全体用户 solved challenge 集合
- 个人 AWD 活动日志查询条件

- [ ] **Step 2: 更新个人与班级统计查询**

实现：

- `GetPersonalStats`
- `ListPersonalDimensionStats`
- `GetClassAverageScore`
- `ListClassTopStudents`

- [ ] **Step 3: 更新时间线与证据链查询**

实现：

- `GetStudentTimeline`
- `GetStudentEvidence`

- [ ] **Step 4: 更新复盘摘要与观察逻辑**

实现：

- `countCorrectSubmissions`
- `hasRepeatedWrongSubmissions`
- `hasHandsOnExploit`

- [ ] **Step 5: 重跑定向测试至 GREEN**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/assessment/infrastructure -run 'ReportRepository.*AWD' -count=1
timeout 120s go test ./internal/module/assessment/application/commands -run 'ReviewArchive.*AWD|BuildReviewArchive.*AWD' -count=1
```

## Task 4: 做最小充分回归

- [ ] **Step 1: 跑 assessment 相关回归**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
timeout 120s go test ./internal/module/assessment/infrastructure -count=1
timeout 120s go test ./internal/module/assessment/application/commands -count=1
timeout 120s go test ./internal/app -run 'TestNewRouter|TestFullRouter' -count=1
```

- [ ] **Step 2: 提交 phase12**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design
git add code/backend/internal/module/assessment/infrastructure/report_repository.go \
  code/backend/internal/module/assessment/infrastructure/report_repository_test.go \
  code/backend/internal/module/assessment/application/commands/report_service.go \
  code/backend/internal/module/assessment/application/commands/report_service_test.go
git add -f docs/architecture/features/2026-04-13-awd-phase12-report-archive-alignment-design.md \
  docs/superpowers/plans/2026-04-13-awd-phase12-report-archive-alignment-implementation.md
git commit -m "feat(assessment): 对齐AWD报告与复盘归档口径"
```
