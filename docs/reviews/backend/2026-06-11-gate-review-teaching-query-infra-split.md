# Teaching Query Infrastructure Split Gate Review

日期：2026-06-11

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-teaching-query-infra-split`
- Branch: `task/2026-06-11-teaching-query-infra-split`
- Task: `2026-06-11-teaching-query-infra-split`
- Plan: `docs/plan/impl-plan/2026-06-11-teaching-query-infra-split-implementation-plan.md`
- Diff source: 当前 worktree 未提交 diff
- Files reviewed:
  - `code/backend/internal/module/teaching_query/runtime/module.go`
  - `code/backend/internal/module/teaching_query/architecture_test.go`
  - `code/backend/internal/module/teaching_query/infrastructure/{shared_rows.go,class_query_repository.go,student_directory_repository.go,student_profile_repository.go,student_activity_repository.go,class_insight_repository.go,overview_repository.go,repository_test.go}`
  - `code/backend/internal/app/composition/assessment_module.go`
  - `code/backend/internal/testutil/systemapp/practice_flow.go`
  - `code/backend/cmd/seed-teaching-review-data/main.go`
  - `docs/architecture/backend/02-database-design.md`
  - `docs/architecture/features/{教师教学概览聚合架构,攻击证据链与教学复盘架构,教学复盘优化设计,教学复盘建议生成架构}.md`

## Classification Check

同意当前任务按 `非琐碎任务` 处理。该 diff 同时触达 `teaching_query` runtime wiring、infrastructure 结构、跨模块 composition/testutil/seed tooling 装配、架构 guard 和架构事实文档，属于典型结构性后端重构。

## Gate Verdict

Pass with review-process limitation.

当前 self-check 未发现需要返修的 material finding，`completion-full` 已通过。但这份 review 仍来自同一实现上下文，不是独立 `code-reviewer` subagent，因此不能算真正满足 `code-workflow` 的独立 reviewer gate。

## Findings

无剩余 material finding。

本轮 review / completion 过程中实际修正了两个真实漏点：

- `code/backend/internal/app/composition/assessment_module.go`
  - `assessment` composition 仍引用已删除的 `queryinfra.NewRepository(root.DB())`。
  - 已改为 `queryinfra.NewClassInsightRepository(root.DB())`，并通过 `completion-full` backend architecture gate 复验。
- `code/backend/internal/testutil/systemapp/practice_flow.go` 与 `code/backend/cmd/seed-teaching-review-data/main.go`
  - 测试支撑和 seed tooling 仍保留旧宽 repo 装配，说明重构还没有覆盖全部 wiring surface。
  - 已改为按具体 query port 组合小 adapter，seed 侧同时补了最小 `AssessmentClassInsightRepository` adapter。

## Material Findings

无。

## Senior Implementation Assessment

当前实现方向是这块 surface 上最小且正确的收口：

- `application/queries` 和 `ports` 既有边界保持不变，没有把 query port 再改成 provider-owned 宽接口。
- `runtime/module.go` 从单个 `repo interface{...}` + `NewRepository(db)` 变成按 `TeachingClassQueryRepository`、`TeachingStudentDirectoryRepository`、`TeachingStudentProfileRepository`、`TeachingStudentActivityRepository`、`TeachingClassInsightRepository`、`TeachingOverviewRepository` 分别装配 concrete，再在 consumer 侧按需要组合。
- 宽 `infrastructure/repository.go` 被真正拆掉，而不是只在 runtime 外面包一层 facade；新的大头 surface 也收敛到了 `class_insight_repository.go` 和 `student_activity_repository.go` 两个明确 owner。

这比“保留一个 God repo，只是多建几个 wrapper constructor”风险更低，因为后者会继续让后续需求在同一个 owner-mixed 文件里膨胀。

## Required Re-validation

本轮实际执行并通过：

```bash
cd code/backend && go test ./internal/module/teaching_query/... -count=1
cd code/backend && go test ./internal/module/teaching_query -run 'TestRuntimeUsesTypedDeps|TestInfrastructureDoesNotExposeWideRepository' -count=1
cd code/backend && go test ./internal/module/teaching_query/infrastructure -run TestClassInsightRepositoryListClassTeachingFactSnapshotsBackfillsAWDSuccessDimensionFacts -count=1
git diff --check
python3 scripts/check-docs-consistency.py
bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full
```

`completion-full` 中途两次失败并已修复：

- 第一次暴露 `assessment_module.go` 仍引用旧 `NewRepository`。
- 第二次暴露 `practice_flow.go` 仍引用旧 `NewRepository`。

修复后再次运行 `completion-full`，最终通过。

## Residual Risk

- 当前工具集中没有独立 reviewer subagent 可用，因此独立 review gate 仍未满足；本记录只能算同上下文 self-check 证据。
- `class_insight_repository.go` 现在约 `783` 行，职责已经集中在 class insight / teaching fact snapshot enrichment，不再是跨全部教师查询的宽仓储；但如果这块后续继续膨胀，下一步应优先把 snapshot enrichment 再拆成更细的 loader/helper，而不是重新引入宽 repo。

## Touched Known-debt Status

- 本轮触达的是已确认的 oversized / owner-mixed `teaching_query` infrastructure surface。
- 当前 diff 已在 touched surface 内完成收口：宽 `Repository` / `NewRepository` 已移除，runtime 不再持有宽 `repo interface`，相关 composition/testutil/seed wiring 也已同步更新。
- 未发现同一 touched surface 上仍残留 blocker 级结构债。
