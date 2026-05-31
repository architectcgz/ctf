# 2026-05-31 report status polling reporting owner tightening plan

> 状态：Draft
> 关联 reuse decision：`.harness/reuse-decisions/report-status-polling-reporting-owner-tightening.md`

## 目标

把 `useReportStatusPolling` 从 `shared/model/common/` 收紧到 `shared/model/reporting/`，让目录语义和实际 owner 一致。

## 非目标

- 不回退到历史 `src/composables/`
- 不新建独立定时器工具
- 不改变当前注入式实现
- 不改各 feature 的导出成功 / 失败处理逻辑

## 输入事实源

- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/src/shared/model/common/useReportStatusPolling.ts`
- `code/frontend/scripts/frontend-architecture-policy.json`

## 目标归属

- `useReportStatusPolling` -> `shared/model/reporting/useReportStatusPolling.ts`

理由：

- 它是共享 reporting workflow owner，而不是 `common` 级的泛化共享状态
- 当前所有调用点都稳定围绕报告导出 / 报告生成状态，没有更宽的复用语义

## 任务切片

### Slice 1

迁移共享 owner 与测试：

- 移动 `useReportStatusPolling.ts`
- 移动对应测试到 `shared/model/reporting/__tests__/`

验证：

- `cd code/frontend && timeout 180s npm run test:run -- src/shared/model/reporting/__tests__/useReportStatusPolling.test.ts`

### Slice 2

修正全部消费路径：

- `useAwdReviewDetailPage`
- `useContestExportFlow`
- `useUserProfilePage`
- `useClassReportExport`
- `useStudentAnalysisPage`
- `useStudentReviewArchivePage`

验证：

- `cd code/frontend && timeout 180s npm run typecheck`
- `cd code/frontend && timeout 180s npm run test:run -- src/__tests__/architectureBoundaries.test.ts`

### Slice 3

对齐架构文档事实：

- 更新 `01-architecture-overview.md`
- 更新 `03-state-management.md`

验证：

- `python3 scripts/check-docs-consistency.py`
- `git diff --check`

## 风险点

- 文档和计划里现在还记录着 `shared/model/common` 的旧落点，需要一起更新
- 这是 owner tightening，不应该混入实现行为变化

## Review focus

- `reporting` 目录命名是否比 `common` 更准确
- 是否只发生 owner 收紧，没有把已有共享状态又打散成无谓的新层级
