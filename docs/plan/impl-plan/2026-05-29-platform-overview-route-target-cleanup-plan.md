> 状态：Current
> 事实源：platform overview page models、overview/cheat route views、前端架构 allowlist
> 替代：无

# Platform Overview Route Target Cleanup Plan

## 目标

- 把 `usePlatformOverviewPage.ts` 与 `useCheatDetectionPage.ts` 从单次 `router.push()` 收口为纯 route target contract。
- 保持数据 owner 不变，同时再清掉 2 条 `featureRouterImportAllowlist`。

## 非目标

- 不处理 `useAuditLogPage.ts`
- 不改 overview / cheat detection 的加载、展示或刷新 owner
- 不继续做大组件拆分

## 输入依据

- `code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts`
- `code/frontend/src/features/platform-overview/model/useCheatDetectionPage.ts`
- `code/frontend/src/features/platform-overview/ui/PlatformOverviewPage.vue`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue`
- `code/frontend/src/components/platform/cheat/CheatDetectionWorkspacePanel.vue`
- `code/frontend/src/components/platform/cheat/CheatDetectionHeroPanel.vue`
- `code/frontend/src/components/platform/cheat/CheatDetectionReviewPanels.vue`
- `code/frontend/src/views/platform/PlatformOverview.vue`
- `code/frontend/src/views/platform/CheatDetection.vue`
- `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
- `code/frontend/src/views/platform/__tests__/CheatDetection.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- 两个 page model 都不需要 route-aware owner，只需要 route target contract。
- 在同一个 feature 内一起收口，比只改其中一条更合理，避免 feature 内留下重复模式。
- route view 和下游 panel 直接消费 `RouterLink`，能保住“view 不直接 useRouter”的边界。

## 设计边界

### `platformOverviewRoutes.ts` 本轮负责

- 生成 overview / cheat detection 所需的审计日志 route target
- 生成作弊检测页 route target

### `usePlatformOverviewPage()` / `useCheatDetectionPage()` 本轮负责

- 数据加载、错误状态、快捷动作数据 owner
- 返回 route target，不再直接导航

### `PlatformOverviewPage.vue` / `CheatDetection*` 本轮负责

- 通过 `RouterLink` 消费 route target
- 保持刷新等事件 owner 不变

## 任务切片

### Slice 1：抽 route target contract

- 目标：
  - 新增 `platformOverviewRoutes.ts`
  - 两个 page model 去掉 `vue-router`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - allowlist 是否可净减少 2 条

### Slice 2：overview / cheat UI 切到 RouterLink

- 目标：
  - hero / cheat panel 直接消费 route target
  - route view 继续保持薄组合层
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/PlatformOverview.test.ts src/views/platform/__tests__/CheatDetection.test.ts`
- Review focus：
  - UI contract 是否清楚，不把刷新 owner 也顺手打散

### Slice 3：allowlist / review / backlog 收尾

- 目标：
  - 更新 allowlist、review 和 backlog
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `platform-overview` feature 内这两条 router 例外是否都收掉

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/PlatformOverview.test.ts src/views/platform/__tests__/CheatDetection.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-overview-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-platform-overview-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-platform-overview-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-overview/model/index.ts code/frontend/src/features/platform-overview/model/platformOverviewRoutes.ts code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts code/frontend/src/features/platform-overview/model/useCheatDetectionPage.ts code/frontend/src/features/platform-overview/ui/PlatformOverviewPage.vue code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue code/frontend/src/components/platform/cheat/CheatDetectionWorkspacePanel.vue code/frontend/src/components/platform/cheat/CheatDetectionHeroPanel.vue code/frontend/src/components/platform/cheat/CheatDetectionReviewPanels.vue code/frontend/src/views/platform/PlatformOverview.vue code/frontend/src/views/platform/CheatDetection.vue code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts code/frontend/src/views/platform/__tests__/CheatDetection.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `AuditLog` 自身的 route owner 不在这轮范围内。
