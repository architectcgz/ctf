# managed instance directory owner 收口计划

## Objective

- 新建 `features/managed-instance-directory`，统一承接 teacher / platform 共用的 managed instance directory 状态 owner。
- 让 teacher / platform 只保留各自的 filter shape、summary 回填、班级初始化和页面表达。

## Non-goals

- 不改 destroy workflow owner。
- 不改普通 instance `open / extend / destroy`。
- 不改 entities 展示 helper。
- 不把 teacher 班级目录加载或 platform 行映射抽进共享层。

## Source Inputs

- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/shared/model/common/usePagination.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`

## Plan Review Result

- 这里仍然属于 `features/*`，因为问题是“teacher / platform 在这里如何管理实例目录状态”，不是通用 shared 工具。
- 共享层只收口目录加载时序、分页和 debounce，不接管 teacher/platform 的页面表达与过滤字段含义。
- stale request guard 和 cleanup 必须留在共享 owner，不能散回页面。

## Task Slices

### Slice 1: 建立 managed instance directory feature

- 目标：新增 `features/managed-instance-directory`，收口 role-aware 列表加载、分页、debounce 搜索和 abort cleanup。
- 风险：
  - 如果抽成过宽泛的 shared pagination util，会丢掉实例目录的明确业务边界。

### Slice 2: teacher 切到共享 directory owner

- 目标：`useInstances` 只保留 class 初始化、teacher filters 和 summary 回填。
- 风险：
  - 如果 class 默认值或 teacher summary 被抽走，会模糊 teacher 页面 owner。

### Slice 3: platform 切到共享 directory owner

- 目标：`usePlatformInstanceManagementPage` 只保留 status filter、行映射和 route owner。
- 风险：
  - 如果平台行映射或 `formatDateTime` 被拉进共享层，会让 feature 变成页面壳。

### Slice 4: 补共享直测并锁住边界

- 目标：为新目录 owner 补直测，并更新 teacher/platform 源码级断言。
- 风险：
  - 不测 stale request / debounce，后面又容易回流出平行实现。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision managed-instance-directory-owner`
- `npm run test:run -- src/features/managed-instance-directory/model/useManagedInstanceDirectory.test.ts src/pages/teacher/__tests__/InstanceManagement.test.ts src/pages/platform/__tests__/InstanceManage.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- 共享 directory owner 是否只收口目录状态，而没有吸入 teacher/platform 页面表达。
- teacher / platform 是否不再各自维护 stale request guard、abort cleanup 和 debounce search。
- 页码切换、过滤搜索和错误回填行为是否保持稳定。

## Rollback / Recovery

- 如果共享 owner 的参数接口还不顺手，可以调整输入回调，但 managed instance directory 的时序 owner 仍必须保留在共享 feature。
