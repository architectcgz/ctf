<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# teaching-fact-semantics-unification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 recommendation 与 class review 共用同一套 live teaching fact 语义，把普通训练、普通竞赛提交和学生可归因的 AWD 提交统一回流到 snapshot，再由 `internal/teaching/advice` 统一做弱项/补样本/进阶判断。

**Architecture:** `assessment` 与 `teaching_query` 的 repository 继续做 live snapshot owner，但这次要补齐显式 submission 统计与 dimension fact 的竞赛/AWD 语义。`RecommendationService` 不再直接按 `skill_profiles` 自己判断 weak dimension，而是改为消费 teaching snapshot，调用 `teaching/advice` 做 evaluation 与 reason 生成，保证 recommendation 与 class review 看的是同一份事实与同一套规则 owner。

**Tech Stack:** Go, GORM, SQLite package tests, `internal/teaching/advice`

---

## Task Metadata

- Task Slug: `2026-06-11-teaching-fact-semantics-unification`
- Started At: `2026-06-11T10:08:08Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-teaching-fact-semantics-unification`
- Branch: `task/2026-06-11-teaching-fact-semantics-unification`

## Objective And Non-Goals

- Objective:
  - 收口 recommendation / class review 的 live teaching fact 口径，让训练、普通竞赛提交和学生 scoped AWD 提交在 query-time snapshot 中统一反映到 submission stats、dimension facts 和推荐 reason。
- Non-Goals:
  - 不处理 `docs/todos/2026-05-17-project-tech-debt-from-migrations.md` 里的 `challenges.image_id = 0` 清理。
  - 不引入 recommendation 专用持久化表、离线回填作业或新的 profile 计算公式；本次以 query-time snapshot 与 recommendation owner 收口为主，普通 contest 只补最小必要的 cache/profile 刷新事件链。
  - 不改 `internal/teaching/advice` 的规则阈值，不重写 challenge recommendation repository 的筛题排序。

## Inputs

- Source docs:
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/architecture/features/教学复盘建议生成架构.md`
  - `docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`
  - `docs/文档规范.md`
  - `docs/plan/README.md`
- Related architecture/contracts:
  - `code/backend/internal/teaching/advice/advice.go`
  - `code/backend/internal/teaching/classreview/review.go`
  - `code/backend/internal/module/assessment/ports/ports.go`
  - `code/backend/internal/module/challenge/contracts/contracts.go`
- Related prior work:
  - `code/backend/internal/module/assessment/application/commands/report_review_archive_builder.go`
  - `code/backend/internal/module/assessment/infrastructure/repository.go`
  - `code/backend/internal/module/teaching_query/infrastructure/repository.go`
  - `code/backend/internal/module/assessment/application/queries/recommendation_service.go`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达 `assessment` / `teaching_query` / `teaching/advice` 调用边界与推荐读路径，属于会影响模块 owner、查询语义、测试口径和 backlog 收口状态的结构性后端改动。

## Files

- Create:
  - 无
- Modify:
  - `code/backend/internal/module/assessment/infrastructure/repository.go`
  - `code/backend/internal/module/assessment/infrastructure/repository_test.go`
  - `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
  - `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
  - `code/backend/internal/module/assessment/application/commands/profile_service.go`
  - `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
  - `code/backend/internal/module/teaching_query/infrastructure/repository.go`
  - `code/backend/internal/module/teaching_query/infrastructure/repository_test.go`
  - `code/backend/internal/module/contest/contracts/events.go`
  - `code/backend/internal/module/contest/application/commands/submission_service.go`
  - `code/backend/internal/module/contest/application/commands/submission_scoring.go`
  - `code/backend/internal/module/contest/application/commands/submission_service_test.go`
  - `docs/architecture/features/教学复盘建议生成架构.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Review:
  - `code/backend/internal/module/assessment/application/commands/report_review_archive_builder.go`
  - `code/backend/internal/teaching/advice/advice.go`
- Test:
  - `code/backend/internal/module/assessment/infrastructure/repository_test.go`
  - `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
  - `code/backend/internal/module/teaching_query/infrastructure/repository_test.go`
  - `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
  - `code/backend/internal/module/contest/application/commands/submission_service_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `assessment/application/commands/report_review_archive_builder.go` 已经定义了 review archive snapshot 的显式 submission / AWD 事件语义。
  - `teaching/advice` 已经提供 `EvaluateStudent(...)` 与 `BuildRecommendationPlan(...)`，理论上就是 recommendation/class review 的共享规则 owner。
- Reuse / extend / split / create-new decision:
  - 复用 `teaching/advice` 的 evaluation / recommendation plan 逻辑。
  - 扩展 live snapshot owner 的 SQL 聚合与显式计数字段，不新增新的 snapshot type 或并行 service。
  - 不新建 shared repository helper；当前 `assessment` / `teaching_query` 已经各自持有 snapshot owner，实现上保持本地收口。
- Owner boundary:
  - `assessment.Repository.GetStudentTeachingFactSnapshot(...)`：个人 recommendation live snapshot owner。
  - `teaching_query.Repository.ListClassTeachingFactSnapshots(...)`：班级 class review live snapshot owner。
  - `RecommendationService`：只负责取 snapshot、查候选题、映射 contract；不再自己做 weak dimension 推断。
  - `internal/teaching/advice`：继续作为唯一 recommendation / class review 规则 owner。
- Why this is the narrowest safe surface:
  - 只改 live snapshot 与 recommendation 读路径，就能让 recommendation 和 class review 共用同一份事实，不需要同时重写 report export、advice 阈值或 contest 事件流。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这不是单点 bug，而是 recommendation、class review、snapshot owner 三处之间的语义收口，需要先确认 owner 边界和现有实现冲突。
- grill-with-docs findings:
  - 架构文档要求 `RecommendationService` 只消费 teaching snapshot + `teaching/advice`，但当前实现仍直接按 `skill_profiles` 判弱项，和 `assessment/teaching_query` snapshot owner 脱节。
  - live snapshot 当前只把 `contest_id IS NULL` 的 challenge submission 当作 attempt / review evidence，AWD 只补了成功覆盖，导致 class review / recommendation 与 report archive 的显式 submission 语义不一致。
  - `assessment/ports` 已经预留 `RecommendationTeachingFactRepository`，说明推荐链路 owner 原本就该落在 teaching snapshot，而不是继续挂在 profile-only 逻辑上。
- Plan adjustments after challenge:
  - 本次实现不仅改 repository SQL，还要把 `RecommendationService` 改成消费 snapshot + `BuildRecommendationPlan(...)`，否则 repository 修正不会真正进入 recommendation 主链路。
  - 继续保持 `teaching/advice` 阈值不变，本次只收口事实输入与 owner。
  - 为避免普通 contest 事实进入 snapshot 后 recommendation cache / `skill_profiles` 仍停留在旧状态，需要补最小事件链：contest correct submission 发布 `contest.flag_accepted`，由 recommendation cache 与 profile 增量更新共同消费。

## Validation

- Commands:
  - `go test ./internal/module/assessment/... -count=1`
  - `go test ./internal/module/teaching_query/... -count=1`
  - `go test ./internal/module/contest/application/commands/... -count=1`
  - `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
  - `bash scripts/check-backend-architecture.sh --full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `python3 scripts/check-docs-consistency.py`
- Results:
  - `2026-06-11`：`go test ./internal/module/assessment/... -count=1` 通过。
  - `2026-06-11`：`go test ./internal/module/teaching_query/... -count=1` 通过。
  - `2026-06-11`：`go test ./internal/module/contest/application/commands/... -count=1` 通过。
  - `2026-06-11`：`go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1` 通过。
  - `2026-06-11`：`bash scripts/check-backend-architecture.sh --full` 通过。
  - `2026-06-11`：`bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full` 通过。
  - `2026-06-11`：`python3 scripts/check-docs-consistency.py` 通过。
- Manual checks:
  - 核对 recommendation 在 contest/AWD 事实存在时不再只跟随陈旧 `skill_profiles` 弱项。
  - 核对 class review 的 snapshot 维度事实与 recommendation 使用的 snapshot 语义一致。
- Review focus:
  - live snapshot 的 success / failure / evidence 是否与 `report_review_archive_builder.go` 的显式 submission 语义对齐。
  - recommendation 是否彻底移除了独立 weak-dimension owner，而不是在 `advice` 之外保留第二套判断。

## Execution Tasks

### Task 1: 收口 live teaching snapshot 语义

**Files:**
- Modify: `code/backend/internal/module/assessment/infrastructure/repository.go`
- Modify: `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- Test: `code/backend/internal/module/assessment/infrastructure/repository_test.go`
- Test: `code/backend/internal/module/teaching_query/infrastructure/repository_test.go`

- [x] 补齐 student/class snapshot 的显式 submission 计数：challenge submission 走全量 `submissions`，AWD 走学生 scoped `awd_attack_logs(source=submission)`，并产出 `ChallengeSuccessCount / SubmissionSuccessCount / SubmissionFailureCount / AWDAttemptCount / AWDSuccessCount / MaxWrongStreak`。
- [x] 把 dimension fact 的 attempt / success / evidence 收口到统一语义：challenge submission 不再限于 `contest_id IS NULL`，AWD 追加 attempt / success 事件，approved review evidence 不再只看练习侧提交。
- [x] 增加 repository tests，覆盖 contest submission、contest approved review、AWD success/failure 对 snapshot 与 dimension fact 的影响。

### Task 2: 收口 recommendation owner 到 teaching snapshot + advice

**Files:**
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`

- [x] 让 recommendation repository interface 显式依赖 `RecommendationTeachingFactRepository`，读取 live snapshot 而不是直接按 `skill_profiles` 判弱项。
- [x] 用 `teachingadvice.EvaluateStudent(...)` 和 `teachingadvice.BuildRecommendationPlan(...)` 生成 weak dimensions / target dimensions / recommendation reasons，再映射回 API contract。
- [x] 更新 recommendation tests，覆盖 contest/AWD 事实回流后的 recommendation 目标变化，以及 dimension-matched reason 仍保持正确。
- [x] 补最小事件链：普通 contest 正确提交发布 `contest.flag_accepted`，由 recommendation cache 与 profile 增量更新消费，避免 contest 事实进入 snapshot 后继续命中旧缓存。

### Task 3: 同步事实文档与 backlog

**Files:**
- Modify: `docs/architecture/features/教学复盘建议生成架构.md`
- Modify: `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`

- [x] 同步架构文档里 live snapshot / recommendation owner 的当前事实描述，避免继续和代码脱节。
- [x] 如果验证表明本任务已收口 recommendation/class review 语义统一，就更新 migration debt todo；未完全收口则明确剩余债边界，不混到 `image_id = 0`。
- [x] 记录实际执行的验证命令与结果，为 completion / review gate 准备证据。
