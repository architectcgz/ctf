> 状态：Superseded
> 事实源：`widgets/awd-review-workspace`、frontend backlog
> Superseded by：`widgets/awd-review-workspace` 最终 owner 收口实现（见 Git 历史与 `.harness/reuse-decisions/awd-review-component-public-owner-neutralization.md`）
> 替代：无

# AWD Review Component Public Owner Neutralization Implementation Plan

## 说明

这份计划记录的是 2026-05-27 的中间过渡方案：先补 `components/awd-review` barrel，把 shared workspace 对 `components/teacher/awd-review/*` 的直连收成一条中立入口。

后续实现已经进一步完成最终收口：四个 detail 子组件直接并入 `widgets/awd-review-workspace`，`components/awd-review` 与 `components/teacher/awd-review` 不再保留。

当前以代码现状和 `.harness/reuse-decisions/awd-review-component-public-owner-neutralization.md` 为准，不再按本中间方案继续实施。

## 非目标

- 本轮不移动 `AwdReviewRoundSelector.vue` 等实际组件文件位置。
- 本轮不改 AWD review workspace 的交互、筛选、导出、drawer 或详情加载行为。
- 本轮不处理 teacher / platform route view、API owner 或 DTO 命名。

## 输入依据

- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts`
- `code/frontend/src/widgets/awd-review-workspace/index.ts`
- `code/frontend/src/components/teacher/awd-review/*.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- AWD review detail route view 和 workspace owner 已经共享化，但 workspace 内部还挂着 teacher 组件入口，这条 public owner 债适合继续按最小切片收口。
- 参考 review archive 那刀，这里有机会把 allowlist 从四条 teacher 组件路径压到更小，甚至完全删掉。

## 任务切片

### Slice 1：提供中立 AWD review component owner

- 目标：
  - 新增 `components/awd-review/index.ts`，集中导出 round selector / analysis / evidence / team drawer 四个共享面板。
- 预期改动：
  - `code/frontend/src/components/awd-review/index.ts`
- review focus：
  - 不新增组件副本，不改变面板来源

### Slice 2：迁移 workspace import 并收缩 allowlist

- 目标：
  - 让 `AwdReviewWorkspace` 改从中立 barrel import，并同步 architecture allowlist。
- 预期改动：
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- review focus：
  - workspace 模板和交互保持不变
  - allowlist 只保留最小必要例外

### Slice 3：同步测试与 backlog 记录

- 目标：
  - 更新 extraction 测试和 backlog 对残余耦合面的描述。
- 预期改动：
  - `code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-awd-review-component-public-owner-neutralization-review.md`

## 验证

- `npm run test:run -- src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退只需恢复 `AwdReviewWorkspace` 到四个 teacher 组件直连 import，并移除 `components/awd-review/index.ts` 与对应 allowlist 调整。

## 残余风险

- 即便压缩掉 teacher 组件路径，`widgets -> components` 这类 legacy 结构例外是否还要继续保留，仍要以架构测试结果为准。
