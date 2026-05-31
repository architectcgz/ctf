# Reuse Decision

## Change type
frontend refactor / entity presentation owner strengthening

## Existing code searched
- `code/frontend/src/pages/instances/InstanceListRoutePage.vue`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/features/platform/instance-management/ui/InstanceManageWorkspacePanel.vue`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/challenge-detail/ui/ChallengeInstanceCard.vue`
- `code/frontend/src/pages/instances/__tests__/InstanceList.test.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`
- 前端 feature-sliced architecture 迁移台账文档

## Similar implementations found
- `entities/user` 已经承接显示名、用户名 handle、实名 fallback 和 option label 这类稳定展示规则。
- `entities/contest` 已经承接 contest 的状态 / CTA / accent / badge class 等稳定展示语义。
- 当前 `instance` 还没有实体层入口，状态文案、状态 class、等待提示、剩余时间、所属用户 / 题目展示仍散在 `features/instance-list`、`teacher/instances`、`platform/instance-management` 和 `challenge-detail` 中。

## Decision
refactor_existing

## Reason
当前最小正确切片不是直接重做实例页面，而是先把 `instance` 对象如何稳定展示收口到实体层：

- 学生实例页通过 `features/instance-list` 暴露 `getInstanceStatusClass` / `getInstanceStatusLabel` / `getInstanceWaitingHint`
- 教师实例目录本地维护 `statusMeta`、剩余时间格式化、学生 / 题目标题和状态胶囊样式
- 平台实例目录本地维护 `status_label`、所属用户展示和状态筛选文案
- 题目详情里的 `ChallengeInstanceCard.vue` 继续本地维护 status label / class / waiting callout

本轮最小正确改动应该是：

- 新建 `entities/instance`，先承接实例状态展示、等待态提示和轻量展示 helper
- 先迁学生实例页、教师实例目录、平台实例目录这三块主消费面
- 用源码级测试锁住状态文案和 owner，不让 `instance` 展示规则继续回流到 feature / page

本轮不做：

- 不改实例创建 / 销毁 / 延时 workflow owner
- 不改实例 API、轮询策略、倒计时 hook 或浏览器打开逻辑
- 不处理 AWD service / projector / ops 里更专门的 runtime 展示

## Files to modify
- `.harness/reuse-decisions/instance-entity-presentation-owner.md`
- `docs/plan/impl-plan/2026-05-31-instance-entity-presentation-owner-plan.md`
- 预期下一步 first slice：
  - `code/frontend/src/entities/instance/index.ts`
  - `code/frontend/src/entities/instance/model/index.ts`
  - `code/frontend/src/entities/instance/model/presentation.ts`
  - `code/frontend/src/entities/instance/model/presentation.test.ts`
  - `code/frontend/src/pages/instances/InstanceListRoutePage.vue`
  - `code/frontend/src/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue`
  - `code/frontend/src/features/platform/instance-management/ui/InstanceManageWorkspacePanel.vue`
  - `code/frontend/src/pages/instances/__tests__/InstanceList.test.ts`
  - `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
  - `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`

## After implementation
- `entities/instance` 会成为实例状态、等待态提示、状态 class 和轻量 meta 展示的公共 owner。
- 学生实例页、教师实例目录和平台实例目录不再各自维护同一套实例状态文案。
- `ChallengeInstanceCard.vue` 是否纳入同一刀，等 first slice 落地后再决定。
