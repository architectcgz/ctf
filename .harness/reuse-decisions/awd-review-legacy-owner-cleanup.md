# Reuse Decision

## Change type
+component / cleanup / docs

## Existing code searched
- `code/frontend/src/components/platform/awd-review/AwdReviewHeroPanel.vue`
- `code/frontend/src/components/platform/awd-review/AwdReviewDirectoryPanel.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewHeroPanel.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewDirectoryPanel.vue`
- `code/frontend/src/pages/awd-review/PlatformAwdReviewIndexRoutePage.vue`
- `code/frontend/src/pages/awd-review/__tests__/PlatformAWDReviewIndex.test.ts`
- `code/frontend/src/pages/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`

## Similar implementations found
- 平台 AWD 复盘页已经只从 `widgets/awd-review-workspace/*` 导入 `AwdReviewHeroPanel` 和 `AwdReviewDirectoryPanel`。
- `components/platform/awd-review/*` 与 `widgets/awd-review-workspace/*` 中对应文件内容已经是重复副本，不再承担运行时 owner。
- 相关 raw-source 测试与页面组合面也都已经转向 `widgets/awd-review-workspace/*`。

## Decision
refactor_existing

## Reason
- 这次不做“旧目录保留一份镜像副本”的中间态。
- `AwdReviewHeroPanel` 和 `AwdReviewDirectoryPanel` 当前唯一合理 owner 是 `widgets/awd-review-workspace/*`。
- 最小正确动作是删除 `components/platform/awd-review/*` 的旧副本，避免同名双落点继续制造边界漂移。

## Files to modify
- `.harness/reuse-decisions/awd-review-legacy-owner-cleanup.md`
- `code/frontend/src/components/platform/awd-review/AwdReviewHeroPanel.vue`
- `code/frontend/src/components/platform/awd-review/AwdReviewDirectoryPanel.vue`

## After implementation
- 平台 AWD 复盘页面与相关测试继续只认 `widgets/awd-review-workspace/*` 为唯一 owner。
- `components/platform/awd-review/*` 目录退出主路径，不保留桥接壳或重复副本。
