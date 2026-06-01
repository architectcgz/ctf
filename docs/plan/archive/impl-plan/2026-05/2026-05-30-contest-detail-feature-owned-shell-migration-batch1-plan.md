> 状态：In Progress
> 事实源：`code/frontend/src/pages/contests/ContestDetailRoutePage.vue`、`code/frontend/src/components/contests/*.vue`、`code/frontend/src/features/contest-detail/**`
> 替代：无

# Contest Detail Feature-Owned Shell 第一批迁移计划

## 目标

- 把 `ContestDetailRoutePage.vue` 直接依赖的历史 `components/contests/*` 壳组件收口到 `features/contest-detail/ui/*`。

## 非目标

- 不改 `contest-detail` 的 route/query owner。
- 不迁移 `components/contests/awd/*`。
- 不重做学生竞赛详情页面的视觉结构和交互语义。

## 输入依据

- `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
- `code/frontend/src/components/contests/*.vue`
- `code/frontend/src/features/contest-detail/index.ts`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/pages/contests/__tests__/contestDetailUiStrategy.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestDetailRoutePage.vue` 是这组壳组件的唯一主 consumer。
- 这组组件不应继续挂在全局历史 `components/contests` 目录。
- 最小可审阅切片是：补齐 `features/contest-detail/ui` owner，切 route page / tests / 类型声明，再删旧文件。

## 设计边界

### `pages/contests/ContestDetailRoutePage.vue` 本轮负责

- 继续作为 route page 组合面。
- 只从 `@/features/contest-detail` 读取 page model 与 feature-owned UI。

### `features/contest-detail/model` 本轮负责

- 继续持有 route/query/data workflow。
- 本轮不新增业务逻辑，只维持现有 public API。

### `features/contest-detail/ui` 本轮负责

- 承接 overview、announcements、challenge workspace、team workspace 与 dialogs 壳。
- 对外通过 feature public API 暴露，不再依赖旧 `components/contests/*` 路径。

## 任务切片

- [ ] Slice 1：补齐 feature-owned UI owner
  - 目标：
    - 新建 `features/contest-detail/ui/*`
    - 新建 `features/contest-detail/ui/index.ts`
    - 更新 `features/contest-detail/index.ts`
  - 验证：
    - `rg -n "@/components/contests" code/frontend/src/pages/contests/ContestDetailRoutePage.vue code/frontend/src/features/contest-detail`

- [ ] Slice 2：切 route page、测试与类型声明
  - 目标：
    - `ContestDetailRoutePage.vue` 改从 `@/features/contest-detail` 读取 UI
    - raw-source 测试与 `components.d.ts` 改到新 owner
  - 验证：
    - `pnpm vitest run src/pages/contests/__tests__/ContestDetail.test.ts src/pages/contests/__tests__/contestDetailUiStrategy.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts`

- [ ] Slice 3：删除旧文件并同步 backlog
  - 目标：
    - 删除旧 `components/contests` 这 7 个文件
    - 更新前端技术债 backlog 当前事实
  - 验证：
    - `bash scripts/check-frontend-architecture.sh --quick`

## 验证计划

- `python3 harness/checks/check-reuse-decision.py`
- `bash scripts/check-task-intake.sh --reuse-decision contest-detail-feature-owned-shell-migration-batch1`
- `cd code/frontend && pnpm vitest run src/pages/contests/__tests__/ContestDetail.test.ts src/pages/contests/__tests__/contestDetailUiStrategy.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `bash scripts/check-frontend-architecture.sh --quick`

## 残余风险

- `components/contests/awd/*` 仍是下一批迁移面，这轮不收。
- raw-source 测试对 import owner 很敏感，若 feature barrel 导出不完整会直接暴露失败。
