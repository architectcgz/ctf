> 状态：Current
> 事实源：2026-05-24 前端架构 review、当前 `platform-users` 结构、已落地平台 feature 拆分
> 替代：无

# Platform User Management Feature Split Implementation Plan

## 目标

- 把平台用户治理的真实 workflow owner 从 `platform-users` 这个过宽 bucket 中收口到更窄的 `platform-user-management` feature。
- 保留 `platform-users` 为兼容桥，避免一次性删除旧入口。

## 非目标

- 本轮不删除 `platform-users` 目录中的任何历史桥文件。
- 本轮不改用户治理页面的用户可见结构、接口契约或交互流程。
- 本轮不处理 `request.ts` 的全局错误策略。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/features/platform-users/**`
- `code/frontend/src/views/platform/UserManage.vue`
- `code/frontend/src/components/platform/user/PlatformUserFormDialog.vue`

## 当前结论

- `platform-users` 不再真实承接班级、学生、实例页 owner，这三块已经通过 bridge 转到独立 feature。
- 但当前真正的用户治理 workflow 仍放在 `platform-users/model/usePlatformUsers.ts` 与 `usePlatformUserManagePage.ts`。
- 这会让 `platform-users` 继续表现为“用户治理主 owner”，无法兑现 review 中“桶只保留用户治理或退为兼容层”的边界目标。

## 任务切片

### Slice 1：平台用户治理 owner 收口到独立 feature

- 目标：
  - 新建 `platform-user-management` 作为真实用户治理 owner。
  - `UserManage.vue` 与 `PlatformUserFormDialog.vue` 切到新 feature 公共出口。
  - `platform-users` 仅保留兼容桥。
- 预期改动：
  - `docs/plan/impl-plan/2026-05-24-platform-user-management-feature-split-implementation-plan.md`
  - `.harness/reuse-decisions/frontend-architecture-review-prompt-and-platform-feature-split.md`
  - `code/frontend/src/features/platform-user-management/**`
  - `code/frontend/src/features/platform-users/index.ts`
  - `code/frontend/src/features/platform-users/model/index.ts`
  - `code/frontend/src/features/platform-users/model/usePlatformUserManagePage.ts`
  - `code/frontend/src/features/platform-users/model/usePlatformUsers.ts`
  - `code/frontend/src/composables/usePlatformUsers.ts`
  - `code/frontend/src/views/platform/UserManage.vue`
  - `code/frontend/src/components/platform/user/PlatformUserFormDialog.vue`
  - `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- 依赖：
  - 继续复用现有 `usePagination`、`useToast`、`confirmDestructiveAction` 与 `@/api/admin/users`。
  - 不复制新的用户 CRUD / 导入逻辑，只迁移 owner 和公共出口。
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/platform/__tests__/UserManage.test.ts src/features/__tests__/featureBoundaries.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - `npm run typecheck`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 新 feature 是否成为实际用户治理 owner，而不是空壳再包装。
  - `platform-users` 是否只剩兼容桥，不再承载真实用户治理实现。
  - 旧 `usePlatformUsers` / `usePlatformUserManagePage` 入口是否仍保持兼容。

## 风险

- `PlatformUserFormDialog.vue` 直接引用了 `PlatformUserFormDraft`，如果类型出口迁移不完整，会导致组件和 page owner 各自依赖不同入口。
- `usePlatformUsers.ts` 同时承接筛选、分页、CRUD、导入结果和错误提示，若桥接路径处理不一致，容易留下“表面换了 feature，实际还是旧 owner”的假拆分。

## 回退方式

- 如新 feature 引入回归，可回退 `platform-user-management` 新目录和引用更新，并恢复 `UserManage.vue`、`PlatformUserFormDialog.vue`、`usePlatformUsers.ts` 指向 `platform-users` 的原入口。
- `platform-users` 兼容桥会保留，不要求本轮继续删旧文件。
