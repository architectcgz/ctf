# 2026-05-31 composables owner cleanup batch3 report polling plan

> 状态：Draft
> 关联 reuse decision：`.harness/reuse-decisions/composables-owner-cleanup-batch3-report-polling.md`

## 目标

把 `useReportStatusPolling` 从历史 `code/frontend/src/composables/` 收口到 `shared/model/common/`，明确它是跨 feature 共享的报告导出状态轮询 owner。

## 非目标

- 不处理 `routeNavigationTransport`
- 不处理 `routeQueryTransport`
- 不处理 `useRouteQueryTabs`
- 不处理 `useUrlSyncedTabs`
- 不调整各 feature 自己的导出成功 / 失败业务分支，只处理共享轮询 owner

## 输入事实源

- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/scripts/frontend-architecture-policy.json`
- `code/frontend/src/composables/useReportStatusPolling.ts`

## 目标归属

- `useReportStatusPolling` -> `shared/model/common/useReportStatusPolling.ts`

理由：

- 它依赖 `getReportStatus()` 与 `ReportExportData`，本质是共享报告导出状态 owner
- 它服务多个 feature，但不属于任何单一业务对象，也不是纯 `shared/lib` 级别的无 contract 基础能力

## 任务切片

### Slice 1

迁移 composable 与补最小行为测试：

- 移动 `useReportStatusPolling.ts`
- 新增 `shared/model/common/__tests__/useReportStatusPolling.test.ts`

验证：

- `cd code/frontend && timeout 180s npm run test:run -- src/shared/model/common/__tests__/useReportStatusPolling.test.ts`

### Slice 2

修正所有 feature 消费路径：

- `useContestExportFlow`
- `useUserProfilePage`
- `useClassReportExport`
- `useStudentReviewArchivePage`
- `useStudentAnalysisPage`
- `useAwdReviewDetailPage`

验证：

- `cd code/frontend && timeout 180s npm run typecheck`

### Slice 3

对齐架构文档事实：

- 更新 `01-architecture-overview.md`
- 更新 `03-state-management.md`

验证：

- `python3 scripts/check-docs-consistency.py`
- `cd code/frontend && timeout 180s npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- `git diff --check`

## 风险点

- 轮询 composable 自己会 `console.error`，测试里要避免把这个副作用误判成失败
- 多个 feature 会在 ready / failed 分支自行调用 `stopPolling()`，迁移时不能改变这一行为
- 文档里当前仍把导出任务轮询归到 `composables/`，需要和实现一起收口

## Review focus

- 新路径是否匹配“共享报告状态 owner”而不是 router / lib 语义
- 是否只发生 owner 迁移，没有混入导出流程行为变化
- 是否留下旧 `src/composables/useReportStatusPolling` 引用
