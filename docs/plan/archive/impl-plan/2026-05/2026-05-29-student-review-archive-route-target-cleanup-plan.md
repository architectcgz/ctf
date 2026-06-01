> 状态：Current
> 事实源：student review archive page model、review archive hero、前端架构 allowlist
> 替代：无

# Student Review Archive Route Target Cleanup Plan

## 目标

- 把 `useStudentReviewArchivePage.ts` 从“返回学生列表 / 返回学员分析”的 `router.push()` 收口成纯 route target contract。
- 保持复盘归档的数据加载、导出轮询、下载和错误提示 owner 不变，同时再清掉 1 条 `featureRouterImportAllowlist`。

## 非目标

- 不处理导出轮询、下载和 toast owner。
- 不迁移 `ReviewArchiveHero.vue` / `ReviewArchiveWorkspace.vue` 的目录归属。
- 不改复盘归档数据、证据展示或 teacher observation 展示逻辑。

## 输入依据

- `code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue`
- `code/frontend/src/views/platform/PlatformStudentReviewArchive.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue`
- `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components/navigation/AppRouteLink.vue`

## 当前结论

- `useStudentReviewArchivePage.ts` 的 router 依赖只剩 hero 上的两条声明式导航。
- 复盘归档数据、导出轮询、下载和错误提示 owner 都在合适位置，不需要再移动。
- 这一条适合直接收口成 route target contract。

## 设计边界

### `studentReviewArchiveRoutes.ts` 本轮负责

- 生成角色感知的“返回学生列表” route target
- 生成角色感知的“返回学员分析” route target

### `useStudentReviewArchivePage()` 本轮负责

- 复盘归档数据加载、导出轮询、下载和错误提示 owner
- 暴露 `backRoute` 和 `analysisRoute`，不再直接导航

### `ReviewArchiveHero.vue` 本轮负责

- 通过共享 `AppRouteLink` 消费 route target
- 保持“导出复盘归档”仍由 button + emit 驱动

## 任务切片

### Slice 1：page model 去 router 化

- 目标：
  - 新增 review archive route helper
  - `useStudentReviewArchivePage.ts` 去掉 `vue-router`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - allowlist 是否可净减少 1 条

### Slice 2：hero CTA 切到 `AppRouteLink`

- 目标：
  - `ReviewArchiveHero.vue` 的“返回学生列表 / 返回学员分析”改成共享 `AppRouteLink`
  - route view 继续只做组合
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- Review focus：
  - route target contract 是否清楚，没有把导出 workflow 一起迁走

### Slice 3：allowlist / review / backlog 收尾

- 目标：
  - 更新 allowlist、review 和 backlog
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `featureRouterImportAllowlist` 是否净减少 1 条

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-review-archive-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-student-review-archive-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-review-archive-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/student-review-archive-workspace/model/index.ts code/frontend/src/features/student-review-archive-workspace/model/studentReviewArchiveRoutes.ts code/frontend/src/features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts code/frontend/src/views/teacher/TeacherStudentReviewArchive.vue code/frontend/src/views/platform/PlatformStudentReviewArchive.vue code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `useStudentReviewArchivePage.ts` 仍保留 `useRoute()` 读取 params，这轮不一并处理。
