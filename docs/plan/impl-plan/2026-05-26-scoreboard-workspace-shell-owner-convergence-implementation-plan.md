> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ScoreboardView.vue` 当前剩余 workspace 壳、scoreboard feature 已存在的 route/data/filter owner
> 替代：无

# Scoreboard Workspace Shell Owner Convergence Implementation Plan

## 目标

- 把 `ScoreboardView.vue` 里的顶部 tab rail、contest / points 两块 workspace 面板和配套局部样式抽到独立 `ScoreboardWorkspaceShell.vue`。
- 保持父页继续持有 `useScoreboardRoutePage()` 和 `useScoreboardView()` 的 route/query、筛选、分页、刷新、数据和错误 owner。
- 让 `ScoreboardView.vue` 回到 route page 组合 owner，不再承接大块 workspace 模板。

## 非目标

- 本轮不改 `useScoreboardView()`、`useScoreboardContestDirectoryPage()` 或 `useScoreboardRoutePage()` 的数据流与 API 契约。
- 本轮不改排行榜业务语义、筛选规则、实时刷新或分页逻辑。
- 本轮不改 `ScoreboardDetail.vue`。

## 输入依据

- `code/frontend/src/views/scoreboard/ScoreboardView.vue`
- `code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts`
- `code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
- `.harness/reuse-decisions/scoreboard-workspace-shell-owner-convergence.md`

## 当前结论

- `ScoreboardView.vue` 当前 591 行，超出 route view 阈值，但脚本侧已经比较干净，主要剩余重量集中在模板壳层和局部样式。
- 这一页适合继续沿用“父页保留 owner，子组件承接稳定 workspace shell”的既有模式。

## 任务切片

### Slice 1：抽取 scoreboard workspace shell

- 目标：
  - 新增 `ScoreboardWorkspaceShell.vue`，承接顶部 tab rail、contest / points 面板模板和对应局部样式。
  - `ScoreboardView.vue` 继续保留 route/query、筛选、刷新、加载和分页 owner。
- 预期改动：
  - `code/frontend/src/views/scoreboard/ScoreboardView.vue`
  - `code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/__tests__/pageTabsStyles.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 父页是否仍然持有 route/filter/data owner
  - 新组件是否只承接 workspace shell，而没有吸入请求、筛选或分页逻辑
  - 路由页是否低于 view 阈值

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 `ScoreboardView.vue` 从“共享样式对齐”推进到“workspace shell owner 收口”的事实写回前端 review 索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否明确这是路由页壳层收口，而不是 feature owner 迁移

## 风险

- tab rail 抽到子组件后，键盘导航 ref 和 route-tab owner 可能被误迁到壳层组件，导致 query sync 边界变模糊。
- contest / points 两块面板如果 props 设计过宽，容易把整个 scoreboard state 打包透传，变成只有“减行数”没有“收 owner”的伪拆分。

## 回退方式

- 如抽取后出现交互回归，可回退 `ScoreboardWorkspaceShell.vue` 并把模板恢复到 `ScoreboardView.vue`。
- 本轮只影响前端视图层、测试护栏和 review 文档，不涉及后端或 API 契约。
