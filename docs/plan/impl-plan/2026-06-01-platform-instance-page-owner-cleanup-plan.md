# Platform Instance Page Owner Cleanup 计划

## Objective

- 把平台实例页的 page shell owner 从 `pages/platform/InstanceManageRoutePage.vue` 收回 `features/platform/instance-management/ui`。
- 保持平台实例目录的加载、筛选、分页、销毁与 route target owner 不变。

## Non-goals

- 不改 `usePlatformInstanceManagementPage.ts` 的 query、destroy、pagination 或 route target 逻辑。
- 不改 `useManagedInstanceDirectory`、`useManagedInstanceDestroyAction` 的共享 workflow。
- 不改教师实例页、学生实例页或 challenge detail 实例 workflow。

## Source Inputs

- `code/frontend/src/pages/platform/InstanceManageRoutePage.vue`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/platform/instance-management/ui/InstanceManageHeroPanel.vue`
- `code/frontend/src/features/platform/instance-management/ui/InstanceManageWorkspacePanel.vue`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceManagementPage.vue`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这条线优先收口 page owner，不继续拆 page model 或 managed-instance 共享 workflow。
- route page 退回薄壳后，feature page 要成为唯一组合层；不能只是把现有模板搬到新文件里，再让 route page 继续直接绑定 page model。

## Task Slices

### Slice 1: 新增 feature-owned platform instance page shell

- 目标：在 `features/platform/instance-management/ui` 下新增 `PlatformInstanceManagementPage.vue`，由它直接调用 `usePlatformInstanceManagementPage()` 并组合 hero / workspace panel。
- 风险：需要保持现有 admin workspace shell class 与 refresh / destroy / filter 契约不变。

### Slice 2: route page 退回薄壳

- 目标：`pages/platform/InstanceManageRoutePage.vue` 只渲染 `PlatformInstanceManagementPage`，不再直接 import内部 panel 或 page model。
- 风险：raw-source 护栏需要同步，否则后续容易回流到 route page 直连内部 panel。

### Slice 3: 补平台实例页护栏与 backlog 进展

- 目标：更新 `InstanceManage.test.ts` 的 raw-source 断言，并把本轮进展记录到 frontend backlog。
- 风险：只补 owner 护栏，不重写目录页已有运行态用例。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-instance-page-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/platform/__tests__/InstanceManage.test.ts`
- `git diff --check -- .harness/reuse-decisions/platform-instance-page-owner-cleanup.md docs/plan/impl-plan/2026-06-01-platform-instance-page-owner-cleanup-plan.md code/frontend/src/features/platform/instance-management/ui/PlatformInstanceManagementPage.vue code/frontend/src/features/platform/instance-management/ui/index.ts code/frontend/src/pages/platform/InstanceManageRoutePage.vue code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- route page 是否真正退回薄壳，而不是继续直连 page model / feature 内部 panel。
- feature page 是否成为平台实例目录的唯一组合层。
- 平台实例目录原有 refresh、筛选、分页、销毁运行态是否保持不变。

## Rollback / Recovery

- 如果 route page 和 feature page 的导出名或 raw-source 护栏命名需要微调，可以继续改 public API，但不能回退成 route page 直接组合 hero / workspace panel。
