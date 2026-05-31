# platform cheat detection data split 计划

## Objective

- 在 `platform/cheat-detection` 内新增 `useCheatDetectionData`，承接平台作弊检测的数据 owner。
- 让 `useCheatDetectionPage` 只保留审计 route target、快捷操作和展示辅助函数。

## Non-goals

- 不改 `CheatDetectionWorkspacePanel` 的 UI。
- 不改审计 route target contract。
- 不把作弊检测页面和其他平台概览页合成 shared page owner。

## Source Inputs

- `code/frontend/src/features/platform/cheat-detection/model/useCheatDetectionPage.ts`
- `code/frontend/src/pages/platform/__tests__/CheatDetection.test.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewData.ts`
- `code/frontend/src/features/platform/overview/model/usePlatformOverviewPage.ts`

## Plan Review Result

- `platform/cheat-detection` 适合做 `page + data` 拆分。
- data owner 负责请求、loading/error 和初始化。
- page model 保留审计跳转、快捷操作和日期格式化函数。

## Task Slices

### Slice 1: 新建 useCheatDetectionData

- 目标：收口风险检测请求、loading/error 状态。
- 风险：
  - 如果把审计 route 一起搬走，会重新模糊 page owner。

### Slice 2: useCheatDetectionPage 改为消费 data owner

- 目标：保留工作台编排和 route target 语义。
- 风险：
  - 如果 page 继续直接依赖 `getCheatDetection`，就没有真正收口异步 owner。

### Slice 3: 更新源码级和行为测试

- 目标：给新 data owner 补直测，并更新平台端源码断言。
- 风险：
  - 不补失败态测试，异步错误 owner 还会回流进 page model。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision platform-cheat-detection-data-split`
- `npm run test:run -- src/features/platform/cheat-detection/model/useCheatDetectionData.test.ts src/pages/platform/__tests__/CheatDetection.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `useCheatDetectionData` 是否只承接风险检测数据 owner。
- `useCheatDetectionPage` 是否只剩 route target、快捷操作和展示辅助逻辑。
- route page 是否继续只负责组合。

## Rollback / Recovery

- 如果 `useCheatDetectionData` 的返回接口不顺手，可以调整字段组织，但风险检测数据加载 owner 仍必须留在新 composable。
