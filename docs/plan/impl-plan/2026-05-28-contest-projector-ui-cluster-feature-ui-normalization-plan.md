> 状态：Current
> 事实源：`ContestProjector.vue` 当前 UI owner、`features/contest-projector` 既有 page model / types / formatters
> 替代：无

# Contest Projector UI Cluster Feature UI Normalization Plan

## 目标

- 把 `components/platform/contest/projector/*` 下的 projector UI cluster 迁入 `features/contest-projector/ui`
- 让 `ContestProjector.vue` 改为通过 `@/features/contest-projector` public API 组合 projector page model 与 UI
- 同步更新组件声明、raw-source / component 测试与 backlog 记录

## 非目标

- 本轮不调整 `useContestProjectorPage()` 的 route page owner、fullscreen 流程、自动刷新策略或 toast 策略
- 本轮不继续拆 `ContestProjectorAttackMap.vue` 的内部大块展示区或交互 state
- 本轮不新增新的 shared projector capability，也不把 projector UI 抽到 `components/common` 或 `widgets`

## 输入依据

- `code/frontend/src/views/platform/ContestProjector.vue`
- `code/frontend/src/features/contest-projector/index.ts`
- `code/frontend/src/features/contest-projector/model/useContestProjectorPage.ts`
- `code/frontend/src/features/contest-projector/model/projectorTypes.ts`
- `code/frontend/src/features/contest-projector/model/projectorFormatters.ts`
- `code/frontend/src/components/platform/contest/projector/*`
- `code/frontend/src/components/platform/contest/projector/__tests__/*`
- `code/frontend/src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestProjector.vue` 已经通过 `useContestProjectorPage()` 把大屏 route 的数据装配、fullscreen、轮次跟随和生命周期 owner 收到 `features/contest-projector/model`
- 但 route view 仍直接引用旧 `components/platform/contest/projector/*` UI cluster，导致 `contest-projector` 只有 model owner，没有 UI owner
- `contestProjectorTypes.ts` 与 `contestProjectorFormatters.ts` 在 feature model 已经有对应事实源，本轮不应再继续保留旧 `components/*` 版本作为 UI 依赖

## 设计边界

### `features/contest-projector/ui/*` 本轮负责

- projector toolbar
- projector hero
- projector attack map / overlay / styles
- projector focus overlay
- projector leaderboard / service matrix / traffic / events

### `features/contest-projector/model/*` 本轮继续负责

- contest / round 数据加载
- 投屏生命周期、自动刷新和 fullscreen owner
- projector derived data、types 与 formatter

### `views/platform/ContestProjector.vue` 本轮继续负责

- route view composition shell
- `AppLoading` / `AppEmpty` 分支与 feature 组合

### 本轮不动

- projector page model API 形状
- projector route path
- `ContestProjectorAttackMap.vue` 的内部继续拆分

## 任务切片

### Slice 1：projector UI cluster 迁位

- 目标：
  - 把 `projector/*` UI 组件和样式迁入 `features/contest-projector/ui`
  - 组件内部改为依赖 `../model/projectorTypes`、`../model/projectorFormatters`
  - 去掉旧 `components/platform/contest/projector/contestProjectorTypes.ts` 与 `contestProjectorFormatters.ts` 的 UI 依赖
- 验证：
  - `npm run test:run -- src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts src/features/contest-projector/ui/__tests__/ContestProjectorFocusOverlay.test.ts src/features/contest-projector/ui/__tests__/ContestProjectorServiceMatrix.test.ts`
- Review focus：
  - projector UI cluster 是否彻底离开旧 components owner
  - UI 与 model 的边界是否清楚，没有把 page owner 重新塞回 UI

### Slice 2：route / public API / 护栏同步

- 目标：
  - `features/contest-projector/index.ts` 暴露 UI public API
  - `ContestProjector.vue` 改为从 `@/features/contest-projector` 取 UI
  - 更新 `components.d.ts`、page raw-source test 和 backlog 记录
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - route view 是否不再直连旧 `components/platform/contest/projector/*`
  - public API 是否足够收口，不需要 route 深入 feature 内部路径

## 验证计划

- `cd code/frontend && npm run test:run -- src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts src/features/contest-projector/ui/__tests__/ContestProjectorFocusOverlay.test.ts src/features/contest-projector/ui/__tests__/ContestProjectorServiceMatrix.test.ts src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `ContestProjectorAttackMap.vue` 本体仍然很大，这轮只收 UI owner，不继续做内部展示壳切片；如果后续继续下钻，需要按 attack map 内部分区单独再开一刀
- 当前 route view 仍保留较多模板分支和 attack map focus 组合，后续如果 projector route 自身继续膨胀，再单独判断 page shell 是否需要继续下沉
