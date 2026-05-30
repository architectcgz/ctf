# Reuse Decision

## Change type
frontend refactor / route page naming cleanup

## Existing code searched
- code/frontend/src/pages/auth/LoginRoutePage.vue
- code/frontend/src/pages/auth/RegisterRoutePage.vue
- code/frontend/src/pages/errors/NotFoundRoutePage.vue
- code/frontend/src/pages/errors/ForbiddenRoutePage.vue
- code/frontend/src/views/auth/__tests__/LoginView.test.ts
- code/frontend/src/views/auth/__tests__/RegisterView.test.ts
- code/frontend/src/views/errors/__tests__/NotFoundView.test.ts
- code/frontend/src/views/errors/__tests__/ForbiddenView.test.ts
- docs/ui-style-guide.md
- docs/ui-theme/components-error-empty.md

## Similar implementations found
- 学生侧 route page 迁移后，运行时入口已经统一落到 `code/frontend/src/pages/**`，路由也已经直接引用 `*RoutePage.vue`。
- 现有测试已经直接 import `@/pages/auth/*RoutePage.vue` 和 `@/pages/errors/*RoutePage.vue`，说明这轮不需要新增实现，只需要把残留命名和事实源同步到现状。

## Decision
refactor_existing

## Reason
这轮不是新增页面，也不是新增测试模式，而是把 route page 迁移后的残留命名收口到当前事实：

- 测试文件名、`describe(...)`、局部变量名从历史 `*View` 改成 `*RoutePage`
- 当前设计文档里的错误页组件名和路径改成 `src/pages/errors/*RoutePage.vue`

不做：

- 不改历史 review / plan 文档里的当时事实
- 不处理测试目录仍位于 `src/views/**/__tests__` 的目录债
- 不扩展到 `useLoginViewPage.ts` 这类独立命名债

## Files to modify
- .harness/reuse-decisions/route-page-naming-cleanup.md
- code/frontend/src/views/auth/__tests__/LoginView.test.ts
- code/frontend/src/views/auth/__tests__/RegisterView.test.ts
- code/frontend/src/views/errors/__tests__/NotFoundView.test.ts
- code/frontend/src/views/errors/__tests__/ForbiddenView.test.ts
- docs/ui-style-guide.md
- docs/ui-theme/components-error-empty.md

## After implementation
- 运行时 route page、测试命名和当前事实文档将对齐到同一套 `*RoutePage` 语义
- 当前事实源中不再把错误页写成 `src/views/errors/*View.vue`
