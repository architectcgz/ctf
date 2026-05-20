# Frontend Destructive Confirm Replacement Implementation Plan

## Plan Summary

- Objective
  - 用仓库内的自定义危险确认弹窗替换 `element-plus` 的 `ElMessageBox`，保留现有 `confirmDestructiveAction()` Promise 调用方式，并移除相关 Element Plus 前端残留。
- Non-goals
  - 不在这次切片里重写所有历史前端页面设计稿。
  - 不改各业务删除/销毁/结束动作的调用边界与 toast 策略。
  - 不顺手重构无关的通用弹窗模板组件。
- Source architecture or design docs
  - `README.md`
  - `code/frontend/README.md`
  - `docs/architecture/frontend/01-architecture-overview.md`
  - `docs/architecture/frontend/06-components.md`
  - `docs/architecture/frontend/08-build-deploy.md`
- Dependency order
  - 先落自定义危险确认弹窗组件与全局宿主，再替换 `useDestructiveConfirm` 的运行实现，最后清理 Element Plus 依赖、样式残留与受影响测试/README。
- Expected specialist skills
  - `frontend-engineer`
  - `test-engineer`

## Task 1

- Goal
  - 新增可直接复用的危险确认弹窗组件与全局宿主，支持 `v-model + loading + confirm/cancel` 交互，并通过主题 token 适配深浅色。
- Touched modules or boundaries
  - `code/frontend/src/components/common/DeleteConfirmModal.vue`
  - `code/frontend/src/components/common/AppDestructiveConfirm.vue`
  - `code/frontend/src/App.vue`
- Dependencies
  - 依赖现有 `ModalTemplateShell.vue` 的遮罩、ESC 和滚动锁定能力。
- Validation
  - `npm run test:run -- src/composables/useDestructiveConfirm.test.ts src/views/__tests__/destructiveConfirmThemeAlignment.test.ts`
- Review focus
  - 组件是否仍然保留可直接使用的 `v-model`/`loading` 契约。
  - 遮罩关闭、ESC、关闭按钮和确认按钮是否各自落到明确的状态出口。
- Risk notes
  - 需要避免 `ModalTemplateShell` 的 `update:open` 与组件自身 `cancel` 逻辑重复触发。

## Task 2

- Goal
  - 保持 `confirmDestructiveAction()` 现有 Promise API 不变，把运行实现切到新的全局确认框宿主。
- Touched modules or boundaries
  - `code/frontend/src/composables/useDestructiveConfirm.ts`
  - 依赖该 composable 的前端 feature/model
- Dependencies
  - 依赖 Task 1 的宿主组件已经挂到应用根节点。
- Validation
  - `npm run test:run -- src/composables/useDestructiveConfirm.test.ts src/features/platform-challenges/model/usePlatformChallenges.test.ts src/features/challenge-detail/model/useChallengeInstance.test.ts`
- Review focus
  - 取消时不能继续发请求。
  - 多次触发确认时不能留下悬空 Promise。
  - 焦点恢复和全局状态清理是否明确。
- Risk notes
  - 现有大量业务测试通过 mock composable 隔离实现，新增实现测试需要单独覆盖 Promise 状态机。

## Task 3

- Goal
  - 移除前端 Element Plus 依赖、Vite resolver 和已确认无运行使用的 Element Plus 覆盖样式，并同步最小必要 README / 架构总览说明。
- Touched modules or boundaries
  - `code/frontend/package.json`
  - `code/frontend/package-lock.json`
  - `code/frontend/vite.config.ts`
  - `code/frontend/src/main.ts`
  - `code/frontend/src/style.css`
  - `code/frontend/src/assets/styles/teacher-surface.css`
  - `code/frontend/src/assets/styles/element-override.css`
  - `README.md`
  - `code/frontend/README.md`
  - `docs/architecture/frontend/01-architecture-overview.md`
  - `docs/architecture/frontend/06-components.md`
  - `docs/architecture/frontend/08-build-deploy.md`
  - 受影响测试
- Dependencies
  - 依赖 Task 2 完成后，运行态不再需要 `ElMessageBox`。
- Validation
  - `npm run typecheck`
  - `npm run check:theme-tail`
  - `npm run test:run -- src/views/__tests__/destructiveConfirmThemeAlignment.test.ts src/views/platform/__tests__/UserManage.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/ClassManagement.test.ts src/views/teacher/__tests__/TeacherStudentManagement.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts`
- Review focus
  - 前端构建配置里不应再保留 `element-plus` 专用按需解析和 chunk 切分。
  - README / 架构总览至少不能继续把 Element Plus 写成当前事实。
- Risk notes
  - `docs/architecture/frontend/pages/` 下仍有大量历史 `El*` 页面稿，这次只同步高层事实，不做全量清扫。

## Integration Checks

- `confirmDestructiveAction()` 仍可被现有业务 composable 直接 `await`。
- 应用根节点同时挂载 toast 与新的危险确认宿主。
- 前端依赖与 Vite 配置不再显式依赖 `element-plus`。
- 旧 `ElMessageBox` 覆盖样式和 `teacher-surface-table` 样式已移除。

## Review Gate Notes

- 本次改动属于前端多文件实现切片，需要做实现后自审与受影响验证。
- 当前会话未获用户授权使用 subagent 做独立 gate review，因此只能完成实现内自审，不能把“独立 review 已完成”作为完成依据。

## Rollback / Recovery Notes

- 若新确认框回归，可回退 `useDestructiveConfirm.ts`、新宿主组件和 App 挂载点，不涉及后端或数据迁移。
- `package-lock.json` 变更仅影响前端依赖清单，可独立回退。

## Residual Risks

- `docs/architecture/frontend/pages/` 目录中仍有大量历史页面稿继续描述 `El*` 组件，这次不做全量更新。
- 若后续需要在确认框内部展示真实删除进度，需要把现有 Promise API 扩展为“确认后由宿主持有 loading”的新契约，这次不处理。
