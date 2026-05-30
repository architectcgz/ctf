> 状态：Superseded
> 事实源：`widgets/review-archive-workspace`、frontend backlog
> 替代：`widgets/review-archive-workspace` 最终 owner 收口实现（见 Git 历史与 `.harness/reuse-decisions/review-archive-component-public-owner-neutralization.md`）

# Review Archive Component Public Owner Neutralization Implementation Plan

## 说明

这份计划记录的是 2026-05-27 的中间过渡方案：先补 `components/review-archive` barrel，把 shared workspace 对 `components/teacher/review-archive/*` 的直连收成一条中立入口。

后续实现已经进一步完成最终收口：`ReviewArchiveWorkspace`、`ReviewArchiveState`、`ReviewArchiveSummarySection`、presentation model 与四个 detail 子组件都直接并入 `widgets/review-archive-workspace`，`components/review-archive`、`components/teacher/review-archive` 与 `widgets/teacher-review-archive` 不再保留。

当前以代码现状和 `.harness/reuse-decisions/review-archive-component-public-owner-neutralization.md` 为准，不再按本中间方案继续实施。

## 非目标

- 本轮不移动 `ReviewArchiveHero.vue` 等实际组件文件位置。
- 本轮不改 `ReviewArchiveWorkspace` 的结构、导出行为、状态处理或视觉样式。
- 本轮不处理 AWD review shared widget 的 teacher 组件入口。

## 输入依据

- `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue`
- `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.test.ts`
- `code/frontend/src/widgets/review-archive-workspace/index.ts`
- `code/frontend/src/components/teacher/review-archive/*.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/teacher/__tests__/teacherStudentReviewArchiveWorkspaceExtraction.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- review archive 的 route view 和 widget owner 已经中立化，但 widget 内部还挂着 teacher 组件入口，这条 public owner 债适合继续按最小切片收口。
- 由于 widgets -> components 仍属于当前边界测试的 legacy 例外，本轮至少要把 allowlist 从四条 teacher 组件路径压成一条中立 barrel。

## 任务切片

### Slice 1：提供中立 review archive component owner

- 目标：
  - 新增 `components/review-archive/index.ts`，集中导出 hero / observation / evidence / reflection 四个共享面板。
- 预期改动：
  - `code/frontend/src/components/review-archive/index.ts`
- review focus：
  - 不新增组件副本，不改变面板来源

### Slice 2：迁移 workspace import 并收缩 allowlist

- 目标：
  - 让 `ReviewArchiveWorkspace` 改从中立 barrel import，并同步 architecture allowlist。
- 预期改动：
  - `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- review focus：
  - workspace 模板和交互保持不变
  - allowlist 只保留最小必要例外

### Slice 3：同步测试与 backlog 记录

- 目标：
  - 更新 extraction / workspace 测试和 backlog 对残余耦合面的描述。
- 预期改动：
  - `code/frontend/src/widgets/teacher-review-archive/ReviewArchiveWorkspace.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherStudentReviewArchiveWorkspaceExtraction.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-review-archive-component-public-owner-neutralization-review.md`

## 验证

- `npm run test:run -- src/widgets/teacher-review-archive/ReviewArchiveWorkspace.test.ts src/views/teacher/__tests__/teacherStudentReviewArchiveWorkspaceExtraction.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts src/views/platform/__tests__/PlatformStudentReviewArchive.test.ts src/__tests__/architectureBoundaries.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退只需恢复 `ReviewArchiveWorkspace` 到四个 teacher 组件直连 import，并移除 `components/review-archive/index.ts` 与对应 allowlist 调整。

## 残余风险

- 即便压缩到一条中立 barrel，`widgets -> components` 的 legacy 结构例外仍然存在；更彻底的 owner 收口需要后续决定这些 review archive 面板应该继续留在 components 还是迁到 widget 内部局部目录。
