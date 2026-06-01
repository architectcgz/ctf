# managed instance destroy workflow owner 收口计划

## Objective

- 新建 `features/managed-instance-workflow`，统一承接 teacher / platform 共用的 role-aware managed instance destroy workflow。
- teacher / platform 页面只保留列表 reload、页码回退、成功提示和错误展示的本地策略。

## Non-goals

- 不改普通实例 `instance-workflow`。
- 不改 teacher / platform 的目录加载、过滤、分页和路由 owner。
- 不处理 AWD / contest 里的其它 destroy 或 terminate workflow。

## Source Inputs

- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/teacher/instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`

## Plan Review Result

- 这里仍然属于 `features/*`，因为问题是“教师/平台在这里如何执行实例销毁动作”，不是实体展示。
- 共享 owner 只收口 destroy workflow，不吸入 teacher/platform 的目录列表、route 和 filter owner。
- destroy workflow 需要直接防重复点击，不能只靠按钮 disabled。

## Task Slices

### Slice 1: 建立 managed destroy workflow feature

- 目标：新增 `features/managed-instance-workflow`，收口 confirm + role-aware destroy + `destroyingId` guard。
- 风险：
  - 如果只抽一个裸 API helper，teacher/platform 仍会回流出各自的 confirm 和重复点击 guard。

### Slice 2: teacher 切到共享 destroy workflow

- 目标：`useInstances` 保留列表 owner，destroy 逻辑改为调用新 feature；`useInstanceManagementPage` 不再本地持有 confirm。
- 风险：
  - 如果 `useInstanceManagementPage` 继续保留 confirm，owner 仍是双份。

### Slice 3: platform 切到共享 destroy workflow

- 目标：`usePlatformInstanceManagementPage` 只保留 admin 目录状态和错误展示，destroy workflow 改为共享 owner。
- 风险：
  - 如果翻页回退和 reload 被错误抽进共享层，会污染平台列表 owner。

### Slice 4: 锁住行为和边界

- 目标：补共享 feature 直测，并更新 teacher/platform 页面测试，锁住 confirm 与 role-aware destroy 行为。
- 风险：
  - 只测页面文本，不测共享 destroy action，会让重复逻辑再次回流。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision managed-instance-destroy-workflow-owner`
- `npm run test:run -- src/features/managed-instance-workflow/model/useManagedInstanceDestroyAction.test.ts src/pages/teacher/__tests__/InstanceManagement.test.ts src/pages/platform/__tests__/InstanceManage.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- 共享 owner 是否只收口 managed instance destroy workflow，没有吸入 teacher/platform 列表 owner。
- teacher / platform 是否已经不再各自维护确认框 + role-aware destroy 流程。
- 重复点击是否被 handler 层阻止。

## Rollback / Recovery

- 如果共享 feature 的参数形态不合适，可以调整回调接口；但 role-aware managed destroy owner 仍必须停留在共享 feature。
