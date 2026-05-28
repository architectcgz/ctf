> 状态：Current
> 事实源：`ContestProjectorAttackMap.vue` 当前 owner、`ContestProjector.vue` 的 projector route shell、现有 projector UI 护栏测试
> 替代：无

# Contest Projector Attack Map Decomposition Plan

## 目标

- 把 `ContestProjectorAttackMap.vue` 从“map view-model + board DOM mechanics + left/right sidebars + board shell”收口成明确的 attack map owner
- 在 `features/contest-projector` 内补齐 team sidebar / board / stats sidebar / board mechanics composable
- 保持 `ContestProjector.vue` 与现有 projector UI 测试的对外 contract 不变

## 非目标

- 本轮不改 `useContestProjectorPage()`、`useContestProjectorDerived()` 或 projector route shell 的 owner
- 本轮不改 `ContestProjectorFocusOverlay.vue` 与 `ContestProjectorAttackDetailOverlay.vue` 的交互语义
- 本轮不顺手处理 `ContestProjectorServiceMatrix.vue`、`ContestProjectorEvents.vue` 或其它 projector panel

## 输入依据

- `code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMap.vue`
- `code/frontend/src/features/contest-projector/ui/ContestProjectorAttackDetailOverlay.vue`
- `code/frontend/src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts`
- `code/frontend/src/views/platform/ContestProjector.vue`
- `code/frontend/src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts`
- `docs/reviews/frontend/2026-05-28-contest-projector-ui-cluster-feature-ui-normalization-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestProjectorAttackMap.vue` 的真实 owner 有两层：
  - map view-model owner：`visibleEdges`、`teamPanels`、`rankingRows`、detail panel open/close
  - board DOM mechanics owner：drag state、beam path、`ResizeObserver`、localStorage、team/service ref
- 当前最明显的问题不是单纯 CSS 多，而是左右侧栏、中央 board 和 DOM mechanics 全都堆在一个文件里
- 这条线已经有 `ContestProjectorAttackDetailOverlay.vue` 作为 drilldown owner，所以继续把稳定展示区块下沉不会破坏现有 projector 路由边界

## 设计边界

### `ContestProjectorAttackMap.vue` 本轮继续负责

- props contract：`rows / edges / scoreboardRows / firstBlood / latestAttackEvents / expanded / boardOnly`
- `visibleEdges`、`teamPanels`、`rankingRows`、`detailRankingRows` 等 map 展示派生
- `activeDetailPanel` 的 open / close owner
- detail overlay 的数据装配

### `ContestProjectorAttackBoard.vue` 本轮负责

- board title
- beam layer
- team node / service node 呈现
- recent event strip
- board drilldown 样式壳

### `ContestProjectorAttackMapTeamSidebar.vue` 本轮负责

- legend
- first blood block
- team list drilldown 入口

### `ContestProjectorAttackMapStatsSidebar.vue` 本轮负责

- ranking drilldown 入口
- attack stats drilldown 入口

### `useProjectorAttackBoard.ts` 本轮负责

- board `ref`
- team/service DOM ref 注册
- drag offset 持久化
- `ResizeObserver`
- beam 路径计算
- drag start / move / end / reset

### 共享 support 本轮负责

- `AttackMapDetailPanel` 类型
- service key / display name / icon helper
- 避免 board / sidebars / overlay 重复各写一份 projector attack map helper

## 任务切片

### Slice 1：抽出 board mechanics 与共享 support

- 目标：
  - 新增 `useProjectorAttackBoard.ts`
  - 视情况新增 projector attack map support
  - 保持 beam / drag / storage 行为不变
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts`
- Review focus：
  - drag / beam / observer 是否仍只有一个 owner
  - localStorage key 与 expanded / boardOnly 的行为是否没有漂移

### Slice 2：提取左右侧栏与 board shell

- 目标：
  - 新增 `ContestProjectorAttackMapTeamSidebar.vue`
  - 新增 `ContestProjectorAttackBoard.vue`
  - 新增 `ContestProjectorAttackMapStatsSidebar.vue`
  - 父 map 改为只组合 sidebars / board / detail overlay
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts src/features/contest-projector/ui/__tests__/contestProjectorAttackMapExtraction.test.ts`
- Review focus：
  - 父组件是否仍然是唯一 map view-model owner
  - 子组件是否只消费 props / emits，没有重新接管 projector page owner

### Slice 3：同步 raw-source 护栏、review 和 backlog

- 目标：
  - 新增 extraction 护栏
  - 更新 backlog 当前进展
  - 补 frontend review
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - raw-source 护栏是否已经转成聚合源码视角
  - touched surface 上的超大 attack map 债是否真的收口，而不是换成新的大子组件

## 验证计划

- `cd code/frontend && npm run test:run -- src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts src/features/contest-projector/ui/__tests__/contestProjectorAttackMapExtraction.test.ts src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `ContestProjectorAttackMap.css` 目前仍然是大体量全局样式文件；本轮先收口 owner，不额外做 CSS 文件再拆分，避免把切片范围膨胀成“逻辑 + 样式体系重排”。
- `ContestProjectorAttackDetailOverlay.vue` 会被轻触以复用 shared helper；如果共享 helper 边界不清，宁可保持 overlay 只做最小改动，也不把更多 detail 行为回灌到 parent。
