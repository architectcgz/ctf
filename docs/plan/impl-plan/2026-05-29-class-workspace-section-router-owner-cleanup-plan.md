> 状态：Current
> 事实源：class workspace redirect helper、teacher/platform route view、前端架构 allowlist、raw-source 测试
> 替代：无

# Class Workspace Section Router Owner Cleanup Plan

## 目标

- 把 `useClassWorkspaceSection.ts` 从 route-aware helper 收口回纯 alias route 解析 helper。
- 让 `PlatformClassWorkspaceSection.vue` 与 `TeacherClassWorkspaceSection.vue` 承接唯一的 `router.replace()` owner。
- 删除对应 `featureRouterImportAllowlist` 条目。

## 非目标

- 不改 class workspace 的 panel 语义或 canonical route 命名。
- 不重构 `PlatformClassStudents.vue`、`TeacherClassStudents.vue` 或更下游班级工作区 feature。
- 不处理其它 `featureRouterImportAllowlist` 条目。

## 输入依据

- `code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts`
- `code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue`
- `code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue`
- `code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useClassWorkspaceSection.ts` 当前只是把 alias route name 映射到 canonical workspace `panel`，不应继续直接依赖 `vue-router`。
- `PlatformClassWorkspaceSection.vue` 和 `TeacherClassWorkspaceSection.vue` 本身就是 route view，天然适合作为 `useRoute()` / `useRouter()` owner。
- 这条 allowlist 如果继续保留，会继续给 feature helper 留出直接导航的口子。

## 设计边界

### `useClassWorkspaceSection.ts` 本轮负责

- 根据 route-like `name` 解析 `panel`
- 生成 canonical workspace route target
- 不直接执行路由跳转

### `useClassStudentsPage.ts` 本轮负责

- 获取 `useRoute()` / `useRouter()`
- 在 alias route 下优先执行 canonical redirect
- 保持 class students workspace 现有 page owner 身份

## 任务切片

### Slice 1：helper 去掉 `vue-router`

- 目标：
  - 从 `useClassWorkspaceSection.ts` 移除 `useRoute()` / `useRouter()`
  - 改为返回 canonical target computed
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- Review focus：
  - helper 是否已经只保留 route-like contract
  - feature source 是否不再出现 `router.replace`

### Slice 2：class students page owner 承接 redirect owner

- 目标：
  - 在 `useClassStudentsPage.ts` 内承接 canonical redirect 的 `router.replace()`
  - 保持 legacy alias route 到 canonical workspace 的行为不变
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- Review focus：
  - redirect owner 是否明确回到既有 page owner，而不是 view
  - teacher / platform 两侧是否仍通过同一 page owner 保持对称

### Slice 3：allowlist / backlog / 护栏收尾

- 目标：
  - 删除该条 allowlist
  - 更新 raw-source 测试与 backlog 记录
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - 非 page owner 的 router import 是否真实减少
  - backlog 记录是否与当前代码事实一致

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/class-workspace-section-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-class-workspace-section-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-class-workspace-section-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts code/frontend/src/views/platform/PlatformClassWorkspaceSection.vue code/frontend/src/views/teacher/TeacherClassWorkspaceSection.vue code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收口 `class-workspace-redirect` 这一条，不代表其它 `featureRouterImportAllowlist` 都合理。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
