# Contest Operations Hub Panel Feature UI Normalization Plan

## 目标

- 把 `ContestOperationsHubHeroPanel.vue` 与 `ContestOperationsHubWorkspacePanel.vue` 迁入 `features/platform-contests/ui`。
- 让 `ContestOperationsHub.vue` 改为通过 `@/features/platform-contests` public API 组合 panel。
- 同步组件声明、raw-source 护栏测试和 backlog 记录。

## 非目标

- 本轮不改 `useContestOperationsHubPage()` 的分页、推荐赛事、错误处理或导航逻辑。
- 本轮不改 `ContestOperationsHubWorkspacePanel.vue` 内部表格、分页组件或文案结构。
- 本轮不重排 `ContestOperationsHub.vue` 的页面骨架样式。

## 输入依据

- `code/frontend/src/views/platform/ContestOperationsHub.vue`
- `code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts`
- `code/frontend/src/features/platform-contests/ui/index.ts`
- `code/frontend/src/components/platform/contest/ContestOperationsHubHeroPanel.vue`
- `code/frontend/src/components/platform/contest/ContestOperationsHubWorkspacePanel.vue`
- `code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts`
- `code/frontend/src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/contestOperationsHubWorkspaceExtraction.test.ts`
- `code/frontend/src/components.d.ts`

## 当前结论

- `ContestOperationsHub.vue` 是 route shell，owner 已在 `useContestOperationsHubPage()`。
- hero 与 workspace panel 只服务这一路由，属于 `platform-contests` 的 feature-owned UI。
- 最小正确落点是 `features/platform-contests/ui/*`，并通过 `features/platform-contests` public API 暴露。

## 设计边界

### `features/platform-contests/ui/*` 本轮负责

- `ContestOperationsHubHeroPanel.vue`
- `ContestOperationsHubWorkspacePanel.vue`
- `ContestOperationsHub.vue` 对这两个 panel 的 public API 组合

### `features/platform-contests/model/*` 本轮继续负责

- `useContestOperationsHubPage()` 的 route / page workflow

### `components/platform/contest/*` 本轮不再负责

- `ContestOperationsHubHeroPanel.vue`
- `ContestOperationsHubWorkspacePanel.vue`

## 任务切片

### Slice 1：feature UI owner 迁位

- 目标：
  - 新增 `features/platform-contests/ui/ContestOperationsHubHeroPanel.vue`
  - 新增 `features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue`
  - `ContestOperationsHub.vue` 改为从 `@/features/platform-contests` public API 引用 panel
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestOperationsHub.test.ts`
- Review focus：
  - route shell 是否继续只保留 page owner 与 panel 组合
  - panel 行为与 props / emits contract 是否保持不变

### Slice 2：护栏同步

- 目标：
  - 更新 `features/platform-contests/ui/index.ts`
  - 更新 `components.d.ts`
  - 更新 panel extraction / workspace extraction 测试
  - backlog 记录 ContestOperationsHub feature UI 收口进展
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts src/views/platform/__tests__/contestOperationsHubWorkspaceExtraction.test.ts`
- Review focus：
  - touched surface 是否已不再依赖旧 `components/platform/contest/ContestOperationsHub*.vue` 路径
  - `features/platform-contests` public API 是否成为唯一组合入口

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestOperationsHub.test.ts src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts src/views/platform/__tests__/contestOperationsHubWorkspaceExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只处理 contest operations hub 的 page-sized panel owner，不继续拆目录表格内部更细的 primitive。
