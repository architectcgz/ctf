# Reuse Decision

## Change type
frontend refactor / feature slice merge

## Existing code searched
- code/frontend/src/features/class-workspace-redirect/index.ts
- code/frontend/src/features/class-workspace-redirect/model/index.ts
- code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts
- code/frontend/src/features/class-students-workspace/index.ts
- code/frontend/src/features/class-students-workspace/model/index.ts
- code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts
- code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts

## Similar implementations found
- `class-insight-window` 这类只服务单一 page owner 的 helper 已经直接挂在对应 feature 下，而不是额外拆成独立 feature。
- `useClassWorkspaceSection.ts` 当前只被 `useClassStudentsPage.ts` 调用，不再存在单独的 route view / page 组合 owner。

## Decision
refactor_existing

## Reason
`class-workspace-redirect` 当前只剩一个 helper，并且唯一 consumer 已经是 `class-students-workspace` 的 page owner。继续保留独立 feature slice 的收益很低，反而会制造“helper 在 feature，但 feature 只服务另一个 feature 内单一 page model”的额外跳转。

最小正确改动是：

- 把 `useClassWorkspaceSection.ts` 并入 `features/class-students-workspace/model`
- 让 `useClassStudentsPage.ts` 直接从同 feature 本地 model 引用它
- 更新 raw-source 护栏与 backlog 说明
- 删除空掉的 `features/class-workspace-redirect` slice

本轮不做：

- 不改 `useClassStudentsPage.ts` 的 page owner 职责
- 不调整 class workspace route、panel query 或 teacher / platform 页面行为
- 不扩展到其它“单薄 feature”清理

## Files to modify
- .harness/reuse-decisions/class-workspace-redirect-feature-merge.md
- docs/plan/impl-plan/2026-05-29-class-workspace-redirect-feature-merge-plan.md
- docs/reviews/frontend/2026-05-29-class-workspace-redirect-feature-merge-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts
- code/frontend/src/features/class-students-workspace/model/useClassWorkspaceSection.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts
- code/frontend/src/features/class-workspace-redirect/index.ts
- code/frontend/src/features/class-workspace-redirect/model/index.ts
- code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts

## After implementation
- `useClassWorkspaceSection.ts` 落到 `class-students-workspace` 本地 model
- `features/class-workspace-redirect` 退出主路径
- 班级工作区 alias redirect 仍由 `useClassStudentsPage.ts` 统一承接
