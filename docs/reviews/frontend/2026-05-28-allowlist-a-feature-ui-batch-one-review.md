# Allowlist A Feature UI Batch One 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-allowlist-a-feature-ui-batch-one-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/allowlist-a-feature-ui-batch-one.md`
  - `docs/plan/impl-plan/2026-05-28-allowlist-a-feature-ui-batch-one-plan.md`
  - `docs/reviews/frontend/2026-05-28-allowlist-a-feature-ui-batch-one-review.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/scoreboard/ui/ScoreboardRealtimeBridge.vue`
  - `code/frontend/src/features/admin-notification-publisher/ui/AdminNotificationPublishDrawer.vue`
  - `code/frontend/src/features/platform-user-management/ui/PlatformUserFormDialog.vue`
  - `code/frontend/src/features/platform-challenges/ui/ChallengeManageDirectoryPanel.vue`
  - `code/frontend/src/features/teacher-student-analysis/ui/InterventionPanel.vue`
  - `code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportDialog.vue`
  - `code/frontend/src/features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/features/scoreboard/**`
  - `code/frontend/src/features/admin-notification-publisher/**`
  - `code/frontend/src/features/platform-user-management/**`
  - `code/frontend/src/features/platform-challenges/**`
  - `code/frontend/src/features/teacher-student-analysis/**`
  - `code/frontend/src/features/teacher-class-report-export/**`
  - 受影响 `views/**` 与测试文件
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：属于前端 `P2` 级结构收口，主范围是 `A` 档第一批 6 条 feature-owned UI allowlist；实现中因 `ScoreboardRealtimeBridge` 的唯一 legacy component consumer 仍在 `ContestAWDWorkspacePanel.vue`，顺带把该 panel 迁入 `features/contest-awd-workspace/ui`，未扩展到其余 `contest-awd-workspace` 残片。
- Gate verdict：Implemented, pending final workflow gate

## Review focus

- 这 6 个组件是否真正迁入 feature `ui`，而不是只加一层 wrapper 保留旧 owner
- `componentFeatureImportAllowlist` 是否实际减少
- view / page / test import 是否全部切到新 owner

## Findings

- `ScoreboardRealtimeBridge` 迁入 `features/scoreboard/ui` 后，如果保留 `ContestAWDWorkspacePanel.vue` 在 `components/contests`，会新增一条 `components/** -> @/features/scoreboard` 违规；本轮已把该 panel 一并迁入 `features/contest-awd-workspace/ui`，避免留下新的 allowlist 负债。
- `InterventionPanel.vue` 迁入 feature `ui` 后，原有内联 `useRouter()` 会触发“feature router access 应留在 reviewed route-aware composable / page owner”护栏；本轮已把打开学生详情的跳转 owner 提回 `ClassStudentsPage.vue`，`InterventionPanel` 只通过 `open-student` 事件上抛。
- `components/reports/index.ts` 与 `components/teacher/reports/index.ts` 原本计划保留兼容 barrel，但它们会重新引入 `components/** -> @/features/teacher-class-report-export` 例外；本轮已删除这两个 barrel，并把所有 route view / 测试消费面直接切到 `@/features/teacher-class-report-export`。

## Material findings

- 无新增未收口的 material finding。上述 3 个实现中暴露的边界问题已在本轮内同步收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/notifications/__tests__/*.test.ts src/views/platform/__tests__/UserManage.test.ts src/views/platform/__tests__/ChallengeManage.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentManagement.test.ts src/components/notifications/__tests__/AdminNotificationPublishDrawer.test.ts src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- same-context review only；当前 session 未执行独立子代理复核。
- `challenge-detail`、`contest-awd-workspace` 剩余 6 条、`topology`、`layout` 仍在后续批次。
