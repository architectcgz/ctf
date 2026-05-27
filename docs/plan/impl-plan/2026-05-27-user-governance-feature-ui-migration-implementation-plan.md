> 状态：Current
> 事实源：`UserGovernancePage.vue` 当前 owner、`feature-owned UI` 规则、用户治理 route page 与 panel query owner 边界
> 替代：无

# User Governance Feature UI Migration Implementation Plan

## 目标

- 把 `UserGovernancePage.vue` 从 `components/platform/user/` 迁到 `features/platform-user-management/ui/`。
- 把用户治理页里的 `panel` query 读取与切换 owner 从 page shell 收回 `features/platform-user-management/model`。
- 让 `views/platform/UserManage.vue` 直接通过 `features/platform-user-management` public API 组合 page-sized UI 与业务 model。
- 收掉 `UserGovernancePage.vue` 对应的 `legacyComponentPageAllowlist` 例外，并为新的 route-aware feature model 补上 `vue-router` 允许项。

## 非目标

- 本轮不改 `usePlatformUsers()` 的列表请求、删除、导入、对话框保存或分页 owner。
- 本轮不迁 `PlatformUserFormDialog.vue` 或 `UserGovernanceOverviewPanel.vue` / `UserGovernanceDetailModal.vue` / `UserGovernanceImportPanel.vue` 的目录位置。
- 本轮不改变用户治理 overview / import 两个面板的用户可见交互和文案语义。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/platform/user/UserGovernancePage.vue`
- `code/frontend/src/views/platform/UserManage.vue`
- `code/frontend/src/features/platform-user-management/index.ts`
- `code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform-user-management/model/usePlatformUsers.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `UserGovernancePage.vue` 当前只服务 `UserManage` route，并直接围绕 `platform-user-management` 的列表 / 导入 / 编辑 workflow 组织 page shell，已经是典型的单一 feature page-sized UI。
- 这页和 `ChallengeTopologyStudioPage.vue` 不同，它自己还持有 `useRoute/useRouter` 来同步 `panel` query；如果直接迁目录，会把 route/query owner 继续埋在 feature ui 里。
- 用户治理页的最小正确收口点是补一个小型 `panel` route model，把 `overview/import` 切换 owner 收回 feature model，再迁 page shell。

## 设计边界

### route view 继续负责

- 组合 `usePlatformUserManagePage()`、`useUserGovernancePanelRoute()` 与 `UserGovernancePage`
- 组合 `PlatformUserFormDialog`
- 不直接持有 API、删除确认或导入 workflow owner

### `features/platform-user-management/model` 本轮负责

- 列表请求、分页、删除确认、导入、编辑对话框 owner 继续由 `usePlatformUserManagePage()` / `usePlatformUsers()` 持有
- 新增 `panel` query 解析与切换 owner

### `features/platform-user-management/ui` 本轮负责

- 用户治理 page-sized shell
- 消费上层派生状态、route panel 状态与用户操作 handler
- 不直接持有 `useRoute/useRouter`

### `components/platform/user/*` 继续保留

- 稳定的 overview / detail / import 子分区与对话框
- 不再承担用户治理整页 shell owner

## 任务切片

### Slice 1：收回 panel query owner 并迁移 page shell

- 目标：
  - 新增 `useUserGovernancePanelRoute()` 持有 `overview/import` query owner
  - 新增 `features/platform-user-management/ui/UserGovernancePage.vue`
  - `UserManage.vue` 改从 feature public API 引用
- 预期改动：
  - `code/frontend/src/features/platform-user-management/model/index.ts`
  - `code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts`
  - `code/frontend/src/features/platform-user-management/ui/*`
  - `code/frontend/src/features/platform-user-management/index.ts`
  - `code/frontend/src/views/platform/UserManage.vue`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/UserManage.test.ts`
- Review focus：
  - page shell 是否已经不再直接持有 router
  - route view 是否继续保持组合壳

### Slice 2：清理 guardrail 与 backlog

- 目标：
  - 更新 raw-source 测试路径
  - 清理 `legacyComponentPageAllowlist` 例外
  - 为新的 feature router owner 补 allowlist
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/components.d.ts`
  - `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformRootShellCleanup.test.ts`
  - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
  - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts`
  - `code/frontend/src/views/__tests__/journalNoteStyles.test.ts`
  - `code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-user-governance-feature-ui-migration-review.md`
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/UserManage.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/platform/__tests__/journalPlatformShellStyles.test.ts src/views/platform/__tests__/platformRootShellCleanup.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/adminRootHeroLayout.test.ts src/views/__tests__/journalNoteStyles.test.ts src/views/__tests__/pageTabsStyles.test.ts`
- Review focus：
  - allowlist 是否真实下降
  - raw-source 测试是否已切到新 owner

## 结构收口检查

- `UserGovernancePage.vue` 不再作为 `components/*Page.vue` 遗留页存在。
- `UserManage.vue` 只组合 feature public API 与对话框。
- `panel` query owner 不再直接写在 page shell 里。
- touched surface 上至少移除一条 `legacyComponentPageAllowlist`，并新增明确的 feature router owner allowlist。

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/UserManage.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/platform/__tests__/journalPlatformShellStyles.test.ts src/views/platform/__tests__/platformRootShellCleanup.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/adminRootHeroLayout.test.ts src/views/__tests__/journalNoteStyles.test.ts src/views/__tests__/pageTabsStyles.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/user-governance-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-user-governance-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-user-governance-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/platform-user-management code/frontend/src/views/platform/UserManage.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/platform/__tests__/UserManage.test.ts code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts code/frontend/src/views/platform/__tests__/platformRootShellCleanup.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts code/frontend/src/views/__tests__/journalNoteStyles.test.ts code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review 关注点

- `features/platform-user-management/ui` 是否成为用户治理 page shell 的唯一 owner。
- `useUserGovernancePanelRoute()` 是否明确接管 `panel` query owner，而不是把 router 继续埋在 UI 里。
- 测试与 allowlist 是否同步反映新边界，而不是继续绑定旧路径。

## 回退 / 恢复说明

- 如迁移后出现问题，可把 `UserGovernancePage.vue` 移回 `components/platform/user/` 并恢复 route view import。
- 如 `panel` query owner 迁移有问题，可把 `useUserGovernancePanelRoute()` 回退到 page shell 内部实现。

## 残余风险

- overview / detail / import 子分区仍在 `components/platform/user/`，本轮只做 page shell 归位，不处理更深层目录重排。
- 其它 teacher / student 管理页仍在 backlog 中，本轮不一并迁移。
