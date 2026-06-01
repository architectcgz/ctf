# 实例实体展示 owner 收口计划

## Objective

- 新建 `entities/instance`，统一承接实例状态文案、状态样式、等待提示和轻量展示 helper。
- 让 student instance list、teacher instance management 和 platform instance management 直接消费 `entities/instance`，停止在各自文件里内联状态映射和展示兜底。

## Non-goals

- 不修改实例创建、销毁、延时、复制地址、打开靶机和轮询等 workflow owner。
- 不修改实例 API、后端 contract 或倒计时 hook。
- 不在本轮处理 AWD service / projector / ops 的专门 runtime 展示。

## Source Inputs

- `code/frontend/src/pages/instances/InstanceListRoutePage.vue`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/features/platform/instance-management/ui/InstanceManageWorkspacePanel.vue`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/challenge-detail/ui/ChallengeInstanceCard.vue`
- `code/frontend/src/pages/instances/__tests__/InstanceList.test.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`
- `TODO/frontend-sliced-architecture.md`

## Brainstorming Conclusion

- 推荐方向：先用 `entities/instance/model/presentation.ts` 收 `status label / status class / waiting hint / lightweight meta text`，不先抽实例专属 UI 壳。
- 原因：当前散点主要是实例对象的稳定展示规则，先把 owner 下沉到实体层，后续 student / teacher / platform 三个入口才能共用。
- 先不把 `ChallengeInstanceCard.vue` 一起拉进第一刀，避免把 challenge-detail 的更复杂交互和共享实例提示混进首个切片。

## Plan Review Result

- `instance` 实体层只负责“实例对象如何稳定展示”，不承接 create / destroy / extend / open / polling 这类动作。
- 第一刀先收 student / teacher / platform 三个主目录面；等这批共用 helper 站稳后，再决定是否把 challenge-detail 卡片并入第二刀。

## Task Slices

### Slice 1: 建立 instance entity presentation owner

- 目标：新增 `entities/instance` 公共入口，提供状态文案、状态 class、等待提示和基础 meta helper，并补单测。
- 风险：
  - 如果把 workflow、倒计时或浏览器行为混进实体层，会污染 owner。

### Slice 2: 迁移 student instance list

- 目标：让 student instance list 通过实体层消费状态展示 helper，而不是继续从 feature 本地导出展示规则。
- 风险：
  - 如果 `features/instance-list` 仍继续暴露同名展示 helper，owner 会保持双份。

### Slice 3: 迁移 teacher / platform instance directory

- 目标：让教师和平台实例目录停掉各自维护的状态映射、状态胶囊 class 和基础用户 / 题目展示拼接。
- 风险：
  - 如果 teacher / platform 各自保留本地 `statusMeta` 或 `status_label` 生成，会留下平行 owner。

### Slice 4: 锁住边界并更新台账

- 目标：补源码级测试和迁移记录，明确 `instance` 已经进入实体展示层。
- 风险：
  - 如果只测表格渲染，不测 owner 回流，后续很容易再次散开。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision instance-entity-presentation-owner`
- first slice 完成后聚焦验证：
  - `npm run test:run -- src/entities/instance/model/presentation.test.ts src/pages/instances/__tests__/InstanceList.test.ts src/pages/teacher/__tests__/InstanceManagement.test.ts src/pages/platform/__tests__/InstanceManage.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- `entities/instance` 是否只承接实例展示规则，没有吸入 create / destroy / extend / polling owner。
- student / teacher / platform 三个主消费面是否已停止本地持有相同的状态文案和状态 class。
- 现有实例页用户行为是否保持不变。

## Rollback / Recovery

- 如果 helper 的命名或返回结构不合适，可以调整函数形态，但不能回退 owner 边界；实例展示规则仍必须停留在 `entities/instance` 公共入口。
