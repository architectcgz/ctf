# Teacher Instance Page Model Naming Cleanup 计划

## Objective

- 把教师实例 feature 内部过泛的 `useInstances()` / `useInstanceManagementPage()` 命名收紧成直接体现 teacher page owner 的命名。
- 保持教师实例目录的加载、筛选、销毁、分页和 route target 行为不变。

## Non-goals

- 不移动文件路径。
- 不改 `useManagedInstanceDirectory`、`useManagedInstanceDestroyAction` 的共享 workflow。
- 不改教师实例页模板结构、平台实例页、学生实例页或 challenge detail 实例流程。

## Source Inputs

- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/teacher/instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/pages/teacher/InstanceManagementRoutePage.vue`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这条线只收口命名和 raw-source owner 表达，不继续拆 workflow owner。
- 不做文件 rename，避免把“命名可读性”扩大成高风险路径迁移；本轮只改导出函数名和对应消费面。

## Task Slices

### Slice 1: 收紧 teacher instance page-state 命名

- 目标：把 `useInstances()` 改成 teacher-specific 的 page-state owner 命名，并同步 `useInstanceManagementPage.ts` 的内部消费。
- 风险：`challenge-detail` 的相关测试也会 import 这条 composable，需要一起改断言或类型引用。

### Slice 2: 收紧 teacher route page model 命名

- 目标：把 `useInstanceManagementPage()` 改成 teacher-specific page-model 命名，并同步 route page 消费。
- 风险：`InstanceManagement.test.ts` 的 raw-source 护栏要一起更新，否则命名调整会被误判成回归。

### Slice 3: 补 backlog 进展记录

- 目标：把这次命名 owner 收口记录到 frontend backlog，说明 teacher instance feature 已从过泛 page-model 命名收口到 teacher-specific owner 命名。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision teacher-instance-page-model-naming-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/InstanceManagement.test.ts src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `git diff --check -- .harness/reuse-decisions/teacher-instance-page-model-naming-cleanup.md docs/plan/impl-plan/2026-06-01-teacher-instance-page-model-naming-cleanup-plan.md code/frontend/src/features/teacher/instances/model/useInstances.ts code/frontend/src/features/teacher/instances/model/useInstanceManagementPage.ts code/frontend/src/pages/teacher/InstanceManagementRoutePage.vue code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- teacher instance feature 的 page-state / page-model owner 命名是否已经能直接体现边界。
- 命名收口后是否没有把任何运行态行为一起改坏。

## Rollback / Recovery

- 如果下游断言或 import 还漏了旧名字，可以继续补消费面；不能回退成继续保留 `useInstances()` 这种无 owner 的泛命名。
