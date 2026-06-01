# 实例 workflow owner 收口计划

## Objective

- 新建 `features/instance-workflow`，统一承接普通实例的 `open / extend / destroy` workflow。
- 让 `challenge-detail` 和 `instance-list` 只保留各自页面 owner：创建、轮询、列表回填、warning dialog 和 manual-action policy。

## Non-goals

- 不改实例创建 API、列表加载 API 或 polling 策略。
- 不处理 teacher / platform 的 managed instance workflow。
- 不改 AWD workspace 的 contest-aware access workflow。
- 不把动作流程下沉到 `entities/instance`。

## Source Inputs

- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `code/frontend/src/features/instance-list/model/useInstanceOperations.ts`
- `code/frontend/src/features/instance-list/model/useInstanceListPage.ts`
- `code/frontend/src/pages/instances/__tests__/InstanceList.test.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`

## Plan Review Result

- 共享 owner 应落在 `features/*`，因为这里回答的是“用户怎样对实例执行动作”，不是“实例对象如何稳定展示”。
- 这层只抽公共 workflow 主体，不接管页面本地的目标列表、轮询、warning、刷新策略。
- 共享动作应允许页面注入本地策略：目标解析、共享实例 / AWD 限制文案、成功后的回填或刷新。

## Task Slices

### Slice 1: 建立共享 instance workflow feature

- 目标：新增 `features/instance-workflow`，提供 `open / extend / destroy` 动作工厂和直接单测。
- 风险：
  - 如果 helper 只写成无状态 util，重复 guard 和 toast owner 会再次回流到调用方。

### Slice 2: challenge-detail 切换到共享 workflow

- 目标：`useChallengeInstance` 只保留实例创建、刷新和 polling；共享动作改由新 feature 承接。
- 风险：
  - 如果本地 `instance` 更新和 polling 收尾没有保留在页面 owner，会让 feature 吸入页面状态。

### Slice 3: instance-list 切换到共享 workflow

- 目标：`useInstanceOperations` 改为包装共享 workflow，只保留列表 mutation 和 warning 清理。
- 风险：
  - 如果 `warnedInstances`、`warningInstance`、`showWarning` 逻辑被抽走，会模糊列表页 owner。

### Slice 4: 验证边界与行为

- 目标：补共享动作单测，确认现有页面测试与架构边界测试继续通过。
- 风险：
  - 只验证页面快照而不验证共享动作本身，后续又容易在其他页面复制逻辑。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision instance-workflow-owner`
- `npm run test:run -- src/features/instance-workflow/model/useInstanceWorkflowActions.test.ts src/features/challenge-detail/model/useChallengeInstance.test.ts src/pages/instances/__tests__/InstanceList.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- `instance-workflow` 是否只承接共享实例动作，而没有吸入页面轮询、warning 或列表 owner。
- `challenge-detail` 是否只剩 `create / refresh / polling` 责任。
- `instance-list` 是否只剩列表 mutation 和 warning 责任。
- 动作重复点击是否被 handler 自身阻止，而不是只靠按钮 disabled。

## Rollback / Recovery

- 如果共享 helper 的参数形态不合适，可以调整回调接口，但 `open / extend / destroy` owner 仍必须停留在 `features/instance-workflow`。
