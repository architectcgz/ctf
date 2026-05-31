# 2026-05-31 composables owner cleanup batch5 theme plan

> 状态：Draft
> 关联 reuse decision：`.harness/reuse-decisions/composables-owner-cleanup-batch5-theme.md`

## 目标

把 `useTheme` 从历史 `code/frontend/src/composables/` 收口到 `shared/model/theme/`，明确它是全局主题与品牌 owner。

## 非目标

- 不调整主题默认值
- 不调整主题品牌枚举
- 不改图表主题计算逻辑
- 不处理 router / query transport 组
- 不处理 `useWebSocket`
- 不处理 `useProbeEasterEggs`

## 输入事实源

- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/src/composables/useTheme.ts`
- `code/frontend/src/shared/model/layout/topnav/useTopNavViewState.ts`
- `code/frontend/src/shared/ui/charts/*.vue`

## 目标归属

- `useTheme` -> `shared/model/theme/useTheme.ts`

理由：

- `useTheme` 自己维护全局 `theme` / `brand` 单例状态、DOM 主题属性和持久化
- 它是 shared model owner，而不是纯浏览器工具或某个 feature 的局部流程

## 任务切片

### Slice 1

迁移主题 owner 与测试：

- 移动 `useTheme.ts`
- 移动 `useTheme.test.ts`

验证：

- `cd code/frontend && timeout 180s npm run test:run -- src/shared/model/theme/__tests__/useTheme.test.ts`

### Slice 2

修正全部消费路径：

- `App.vue`
- `useTopNavViewState.ts`
- `TopNavBrandPicker.vue`
- 图表组件与图表测试

验证：

- `cd code/frontend && timeout 180s npm run typecheck`
- `cd code/frontend && timeout 180s npm run test:run -- src/shared/ui/charts/__tests__/EChartsMountGate.test.ts`

### Slice 3

对齐架构文档事实：

- 更新 `01-architecture-overview.md`
- 更新 `03-state-management.md`

验证：

- `python3 scripts/check-docs-consistency.py`
- `cd code/frontend && timeout 180s npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- `git diff --check`

## 风险点

- `useTheme` 的状态是模块级单例，测试迁移时不能破坏现有 reset 行为
- `TopNavBrandPicker.vue` 依赖 `ThemeBrand` 类型，路径更新要一起改
- 图表测试会直接 import `useTheme`，要确认 mock / 调用面仍然稳定

## Review focus

- `shared/model/theme` 是否比历史 `composables/` 更清楚地表达 owner
- 是否只发生 owner 迁移，没有混入主题行为变化
- 是否留下旧 `src/composables/useTheme` 引用
