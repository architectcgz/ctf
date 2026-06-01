> 状态：Current
> 事实源：class students workspace page owner、class workspace redirect helper、前端 slice 边界
> 替代：无

# Class Workspace Redirect Feature Merge Plan

## 目标

- 把 `class-workspace-redirect` 这个只剩单一 helper 的 feature slice 并回 `class-students-workspace`。
- 保持 `useClassStudentsPage.ts` 继续作为唯一 page owner，不改变实际行为。
- 删除空掉的旧 feature 目录与 public API。

## 非目标

- 不改班级工作区的 route、panel query、API 请求或 teacher / platform 页面行为。
- 不扩展到其它 feature merge。
- 不进一步拆 `useClassStudentsPage.ts`。

## 输入依据

- `code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `useClassWorkspaceSection.ts` 已经不再是 route-aware helper，只剩 alias route -> canonical target 的局部解析。
- 它只有 `useClassStudentsPage.ts` 这一个 consumer，继续单独占一个 feature slice 没有结构收益。
- 更紧的 owner 是 `features/class-students-workspace/model`，把 helper 并回去更符合“业务 helper 跟随实际 feature owner”。

## 设计边界

### `class-students-workspace` 本轮负责

- 持有 class students page owner
- 持有 class workspace alias redirect 的本地解析 helper
- 对外继续通过 `features/class-students-workspace` public API 暴露 page model / shell

### `class-workspace-redirect` 本轮不再保留

- 不再保留独立 public API
- 不再作为单独 feature slice 存在

## 任务切片

### Slice 1：helper 并入 class students workspace

- 目标：
  - 新增 `features/class-students-workspace/model/useClassWorkspaceSection.ts`
  - `useClassStudentsPage.ts` 改为本地 import
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- Review focus：
  - helper 是否已落到更紧的 feature owner
  - 行为 owner 是否仍停留在 `useClassStudentsPage.ts`

### Slice 2：删除旧 slice 与护栏更新

- 目标：
  - 删除 `features/class-workspace-redirect/*`
  - 更新 raw-source 护栏与 backlog
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- Review focus：
  - 不再存在对旧 feature 的运行时依赖
  - route view 仍然保持薄壳

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts src/views/platform/__tests__/PlatformClassWorkspaceSection.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/class-workspace-redirect-feature-merge.md docs/plan/impl-plan/2026-05-29-class-workspace-redirect-feature-merge-plan.md docs/reviews/frontend/2026-05-29-class-workspace-redirect-feature-merge-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts code/frontend/src/features/class-students-workspace/model/useClassWorkspaceSection.ts code/frontend/src/views/teacher/__tests__/TeacherClassWorkspaceSection.test.ts code/frontend/src/features/class-workspace-redirect/index.ts code/frontend/src/features/class-workspace-redirect/model/index.ts code/frontend/src/features/class-workspace-redirect/model/useClassWorkspaceSection.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮是 feature merge，不是 page owner 再设计；`useClassStudentsPage.ts` 仍偏大，但不在本轮继续扩 scope。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
