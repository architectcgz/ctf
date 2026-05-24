> 状态：Current
> 事实源：2026-05-24 前端架构 review、AWD 复盘当前共享调用链
> 替代：无

# AWD Review Owner Convergence Implementation Plan

## 目标

- 把 AWD 复盘共享 feature / widget 的 owner 从 `teacher-awd-review` 迁到中性目录。
- 让 teacher / platform AWD route view 与详情 workflow 都只依赖中性 owner。
- 删除旧 `teacher-awd-review` 目录，不保留兼容出口。

## 非目标

- 本轮不改 AWD 复盘页面的可见行为、接口契约或路由名。
- 本轮不做 `Teacher*` widget、常量名、测试描述的大规模命名统一。
- 本轮不重排 AWD 复盘内部模板结构或视觉层。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/views/platform/AWDReviewIndex.vue`
- `code/frontend/src/views/platform/PlatformAwdReviewDetail.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
- `code/frontend/src/features/teacher-awd-review/**`
- `code/frontend/src/features/awd-review-detail-workspace/**`
- `code/frontend/src/widgets/teacher-awd-review/**`

## 当前结论

- AWD 复盘共享逻辑已经被 teacher / platform 两侧共同使用。
- 平台目录页仍直接依赖 `@/features/teacher-awd-review`。
- 平台详情页和 teacher 详情页共享 widget，但 widget owner 仍位于 `@/widgets/teacher-awd-review`。
- 详情 workflow 里的导出流也仍从 `teacher-awd-review` feature 读取。
- 迁移到中性 owner 后，目录页共享 feature 对外仍暴露 `useTeacherAwdReviewIndex`，teacher-specific 命名还在泄漏共享边界。
- 详情页共享 feature 对外仍暴露 `useTeacherAwdReviewExportFlow`，teacher-specific 命名同样还在泄漏共享边界。

## 任务切片

### Slice 1：迁移 AWD 共享 feature / widget owner

- 目标：
  - 把 `features/teacher-awd-review` 迁到中性 `features/awd-review-workspace`。
  - 把 `widgets/teacher-awd-review` 迁到中性 `widgets/awd-review-workspace`。
  - 更新 teacher / platform route views、详情 workflow、架构白名单和源码断言测试。
  - 删除旧 `teacher-awd-review` 目录，不保留 re-export 兼容层。
- 预期改动：
  - `.harness/reuse-decisions/awd-review-owner-convergence.md`
  - `docs/plan/impl-plan/2026-05-24-awd-review-owner-convergence-implementation-plan.md`
  - `code/frontend/src/features/teacher-awd-review/**`
  - `code/frontend/src/features/awd-review-workspace/**`
  - `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
  - `code/frontend/src/widgets/teacher-awd-review/**`
  - `code/frontend/src/widgets/awd-review-workspace/**`
  - `code/frontend/src/views/platform/AWDReviewIndex.vue`
  - `code/frontend/src/views/platform/PlatformAwdReviewDetail.vue`
  - `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  - `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
  - `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- 验证：
  - `rg -n "@/features/teacher-awd-review|@/widgets/teacher-awd-review" code/frontend/src`
  - `npm run test:run -- src/views/platform/__tests__/AWDReviewIndex.test.ts src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts src/views/teacher/__tests__/teacherAwdReviewIndexWorkspaceExtraction.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - `npm run typecheck`
  - `bash scripts/check-consistency.sh`
  - `git diff --check -- <touched files>`
- Review focus：
  - 平台 / 教师 route views 是否都只依赖中性 owner。
  - 详情 workflow 是否不再回跳到 role-specific feature。
  - 旧 owner 目录是否已完全退场，而不是仅仅换成兼容别名。

### Slice 2：收口 AWD 目录页共享 hook 命名

- 目标：
  - 把 `features/awd-review-workspace/model/useTeacherAwdReviewIndex.ts` 收口为中性 `useAwdReviewIndex.ts`。
  - 更新共享 feature 导出、teacher / platform route view、源码断言测试和架构白名单。
  - 删除旧文件，不保留兼容 re-export。
- 预期改动：
  - `.harness/reuse-decisions/awd-review-owner-convergence.md`
  - `docs/plan/impl-plan/2026-05-24-awd-review-owner-convergence-implementation-plan.md`
  - `code/frontend/src/features/awd-review-workspace/index.ts`
  - `code/frontend/src/features/awd-review-workspace/model/index.ts`
  - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
  - `code/frontend/src/features/awd-review-workspace/model/useTeacherAwdReviewIndex.ts`
  - `code/frontend/src/views/platform/AWDReviewIndex.vue`
  - `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  - `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- 验证：
  - `rg -n "useTeacherAwdReviewIndex|useAwdReviewIndex" code/frontend/src .harness/reuse-decisions docs/plan/impl-plan`
  - `npm run test:run -- src/views/platform/__tests__/AWDReviewIndex.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - `npm run typecheck`
  - `bash scripts/check-consistency.sh`
  - `git diff --check -- <touched files>`
- Review focus：
  - 共享 feature public API 是否已经移除 teacher-specific hook 名称。
  - teacher / platform 目录页是否全部切到新 owner 命名。
  - 旧 hook 文件是否已删除而不是保留兼容出口。

### Slice 3：收口 AWD 详情页共享导出流命名

- 目标：
  - 把 `features/awd-review-workspace/model/useTeacherAwdReviewExportFlow.ts` 收口为中性 `useAwdReviewExportFlow.ts`。
  - 更新共享 feature 导出、详情页 workflow owner、源码断言测试和活动计划引用。
  - 删除旧文件，不保留兼容 re-export。
- 预期改动：
  - `.harness/reuse-decisions/awd-review-owner-convergence.md`
  - `docs/plan/impl-plan/2026-05-24-awd-review-owner-convergence-implementation-plan.md`
  - `docs/plan/impl-plan/2026-05-24-platform-route-owner-decoupling-implementation-plan.md`
  - `code/frontend/src/features/awd-review-workspace/index.ts`
  - `code/frontend/src/features/awd-review-workspace/model/index.ts`
  - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
  - `code/frontend/src/features/awd-review-workspace/model/useTeacherAwdReviewExportFlow.ts`
  - `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- 验证：
  - `rg -n "useTeacherAwdReviewExportFlow|useAwdReviewExportFlow" code/frontend/src .harness/reuse-decisions docs/plan/impl-plan`
  - `npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - `npm run typecheck`
  - `bash scripts/check-consistency.sh`
  - `git diff --check -- <touched files>`
- Review focus：
  - 共享 feature public API 是否已经移除 teacher-specific 导出流名称。
  - AWD 详情 workflow 是否全部切到新导出流 owner。
  - 旧导出流文件是否已删除而不是保留兼容出口。

## 风险

- 大量源码断言测试直接写死 import 路径，需要同步更新。
- 路径迁移涉及 widget 原始源码引用，若遗漏某个 `?raw` 或 `process.cwd()` 路径，测试会直接失败。
- 仓库当前很脏，提交时必须只 stage 本轮 AWD owner 收口相关文件。

## 回退方式

- 若发现仍有存量调用方依赖旧路径，可从 Git 历史恢复单个文件或目录名，再按真实调用方拆成更小切片。
