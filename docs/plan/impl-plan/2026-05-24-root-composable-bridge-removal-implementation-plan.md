> 状态：Current
> 事实源：2026-05-24 前端架构 review、当前 `src/composables` / `features/*` 双入口现状
> 替代：无

# Root Composable Bridge Removal Implementation Plan

## 目标

- 删除 `src/composables/` 下这批已经没有 runtime 引用的历史 root-composable 桥接文件。
- 同步收口 `featureBoundaries.test.ts` 中的 migrated root-composable 名单，避免测试继续冻结已经退出的历史入口。

## 非目标

- 本轮不改仍承载真实通用逻辑的 composable。
- 本轮不调整对应 feature owner 的实现、接口契约或用户可见行为。
- 本轮不处理 teacher/admin route owner 重新拆分。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/composables/use*.ts`
- `code/frontend/src/features/**/index.ts`
- `code/frontend/src/features/__tests__/featureBoundaries.test.ts`

## 当前结论

- `featureBoundaries.test.ts` 当前通过一份静态 migrated list 约束 `views` / `components` 不再 import 历史 root composable。
- 这份名单里的路径现在已经全部没有 runtime 引用；仓库搜索只剩 `featureBoundaries.test.ts` 自己在引用它们。
- 继续保留根层桥接文件和这份静态名单，会让 “迁移中的目录约定” 持续占用真实目录结构，而不是把收口结果反映到磁盘事实上。

## 任务切片

### Slice 1：删除无引用 root-composable 桥接并收口 migrated list

- 目标：
  - 删除 `src/composables/` 下仅做 `@/features/*` 转发、且仅剩 `featureBoundaries.test.ts` 引用的历史桥接文件。
  - 移除 `featureBoundaries.test.ts` 中已经失效的 migrated root-composable 名单，只保留仍有真实约束价值的边界断言。
- 预期改动：
  - `.harness/reuse-decisions/root-composable-bridge-removal.md`
  - `docs/plan/impl-plan/2026-05-24-root-composable-bridge-removal-implementation-plan.md`
  - `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
  - `code/frontend/src/composables/useChallengeDetailInteractions.ts`
  - `code/frontend/src/composables/useChallengeDetailPresentation.ts`
  - `code/frontend/src/composables/useChallengeInstance.ts`
  - `code/frontend/src/composables/useChallengeTopologyStudioPage.ts`
  - `code/frontend/src/composables/useContestAWDWorkspace.ts`
  - `code/frontend/src/composables/useChallengeManagePage.ts`
  - `code/frontend/src/composables/useChallengePackageImport.ts`
  - `code/frontend/src/composables/useChallengeWriteupEditorPage.ts`
  - `code/frontend/src/composables/useContestAnnouncementManagement.ts`
  - `code/frontend/src/composables/useContestDetailPage.ts`
  - `code/frontend/src/composables/useAuditLogPage.ts`
  - `code/frontend/src/composables/useAdminNotificationPublisher.ts`
  - `code/frontend/src/composables/useAuth.ts`
  - `code/frontend/src/composables/useAwdCheckResultPresentation.ts`
  - `code/frontend/src/composables/useAwdInspectorCoreState.ts`
  - `code/frontend/src/composables/useAwdInspectorDerivedData.ts`
  - `code/frontend/src/composables/useAwdInspectorExports.ts`
  - `code/frontend/src/composables/useAwdInspectorFilters.ts`
  - `code/frontend/src/composables/useAwdInspectorFormatting.ts`
  - `code/frontend/src/composables/useAwdInspectorSummaryMetrics.ts`
  - `code/frontend/src/composables/useAwdTrafficPanel.ts`
  - `code/frontend/src/composables/useContestAnnouncementRealtime.ts`
  - `code/frontend/src/composables/useContestAwdChallengePicker.ts`
  - `code/frontend/src/composables/useContestAwdPreviewRealtime.ts`
  - `code/frontend/src/composables/useContestChallengePool.ts`
  - `code/frontend/src/composables/useContestEditAwdWorkspace.ts`
  - `code/frontend/src/composables/useContestExportFlow.ts`
  - `code/frontend/src/composables/useContestProjectorData.ts`
  - `code/frontend/src/composables/useContestProjectorDerived.ts`
  - `code/frontend/src/composables/useContestScoreboardRealtime.ts`
  - `code/frontend/src/composables/useContestWorkbench.ts`
  - `code/frontend/src/composables/useImageManagePage.ts`
  - `code/frontend/src/composables/useInstanceListPage.ts`
  - `code/frontend/src/composables/useNotificationRealtime.ts`
  - `code/frontend/src/composables/usePlatformContestAwd.ts`
  - `code/frontend/src/composables/usePlatformAwdChallenges.ts`
  - `code/frontend/src/composables/usePlatformChallenges.ts`
  - `code/frontend/src/composables/usePlatformContests.ts`
  - `code/frontend/src/composables/usePlatformOverviewWorkspace.ts`
  - `code/frontend/src/composables/useScoreboardView.ts`
  - `code/frontend/src/composables/useSkillProfilePage.ts`
  - `code/frontend/src/composables/useStudentDirectoryQuery.ts`
  - `code/frontend/src/composables/useStudentFilters.ts`
  - `code/frontend/src/composables/useStudentListQuery.ts`
  - `code/frontend/src/composables/useTeacherAwdReviewDetail.ts`
  - `code/frontend/src/composables/useTeacherAwdReviewIndex.ts`
  - `code/frontend/src/composables/useTeacherClassReportExport.ts`
  - `code/frontend/src/composables/useTeacherDashboardMetrics.ts`
  - `code/frontend/src/composables/useTeacherInstances.ts`
  - `code/frontend/src/composables/useTeacherStudentAnalysisPage.ts`
  - `code/frontend/src/composables/useTeacherStudentReviewArchive.ts`
  - `code/frontend/src/composables/useTeacherWorkspace.ts`
- 验证：
  - `rg -n "@/composables/use" code/frontend/src`
  - `npm run test:run -- src/features/__tests__/featureBoundaries.test.ts`
  - `npm run test:run -- src/views/platform/__tests__/ImageManage.test.ts src/views/platform/__tests__/ChallengeManage.test.ts src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/profile/__tests__/SkillProfile.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
  - `npm run typecheck`
  - `bash scripts/check-consistency.sh`
  - `git diff --check -- <touched files>`
- Review focus：
  - 删除后是否仍有 runtime/test 代码尝试 import 已删 root composable。
  - `featureBoundaries.test.ts` 是否从“维护历史名单”转成只保留仍然有效的边界检查。

## 风险

- 如果仓库外层脚本、别名 mock 或未覆盖测试仍依赖已删 root composable，删除后会在 typecheck 或 targeted tests 中暴露。
- 一次删除文件较多，若 stage 范围不严，容易混入其他无关脏改。

## 回退方式

- 如发现仍有存量引用，可从 Git 历史恢复单个 root bridge 文件，再按真实调用方重新拆更小切片。
