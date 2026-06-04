# Reuse Decision

## Change type
component styling / loading state polish

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## Similar implementations found
- `StudentInsightPanel.vue` 是学生分析各 tab 的公共 loading owner，六维分布、推荐任务、题解列表切换学生时都会先走这层 skeleton。
- `StudentInsightManualReviewSection.vue` 还单独拥有右侧详情 loading 壳，这里需要单独去掉 panel 感。

## Decision
extend_existing

## Reason
- 用户这次指出的是 loading 态残留玻璃屏，不是已加载内容。
- 最小正确改动是把学生分析公共 skeleton 和人工审核详情 loading 壳改成平面骨架容器，不动数据流和已加载内容结构。

## Files to modify
- `.harness/reuse-decisions/student-analysis-loading-remove-glass.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## After implementation
- 六维分布、推荐任务、发布题解列表等学生分析 tab 在 loading 时不再出现玻璃卡片。
- 人工审核右侧详情在 loading 时也改为平面骨架壳。
