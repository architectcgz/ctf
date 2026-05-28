> 状态：Current
> 事实源：本次工作树 diff、相关 layout / feature / guardrail 测试结果
> 替代：无

# Layout Feature Boundary Cleanup Review

## Review target

- Repository: `ctf`
- Branch: `main`
- Diff source: 当前工作树未提交改动
- Files reviewed:
  - `code/frontend/src/widgets/layout-shell/index.ts`
  - `code/frontend/src/widgets/layout-shell/model/useLayoutNotificationDrawerBridge.ts`
  - `code/frontend/src/widgets/layout-shell/model/useLayoutNotificationRealtimeBridge.ts`
  - `code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.ts`
  - `code/frontend/src/components/layout/AppLayout.vue`
  - `code/frontend/src/components/layout/NotificationDrawer.vue`
  - `code/frontend/src/components/layout/TopNav.vue`
  - `code/frontend/src/features/auth/model/useAuth.ts`
  - `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
  - `code/frontend/src/features/notifications/model/useNotificationRealtime.ts`
  - `code/frontend/src/components/layout/__tests__/AppLayout.test.ts`
  - `code/frontend/src/components/layout/__tests__/TopNav.test.ts`
  - `code/frontend/src/features/notifications/model/useNotificationDrawer.test.ts`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Classification check

- 结论：同意本次改动按 non-trivial frontend refactor 处理。
- 原因：改动同时触达 layout、feature model、架构护栏与 backlog 记录，且目标是结构边界收口，不是单文件小修。

## Gate verdict

- 结论：`pass with minor issues`
- 说明：本次同上下文自审未发现 material finding，验证也齐全；但按 pipeline 标准，独立 review gate 仍未满足。

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 当前方案保持了通知与登出 workflow 的 feature owner，同时通过显式 `widgets/layout-shell` bridge 清掉了 layout 对 feature 的直接 import，满足了当前 guardrail 的边界约束。
- 这次用 `widgets/*` bridge 作为 layout 与 feature 之间的适配层，blast radius 小，现有行为与测试基本不动，属于这轮最稳妥的最小改动。
- 更激进的替代方案是引入更明确的 `widgets/*` bridge 或重新定义 layout shell / cross-cutting workflow owner；这会是下一层架构整理，不适合混进这次只清 3 条 allowlist 的切片。

## Required re-validation

- 已执行，无额外 required re-validation。
- 如果后续继续下钻 layout owner，优先重跑：
  - `cd code/frontend && npm run test:run -- src/components/layout/__tests__/AppLayout.test.ts src/components/layout/__tests__/NotificationDrawer.test.ts src/components/layout/__tests__/TopNav.test.ts src/__tests__/architectureBoundaries.test.ts`

## Residual risk

- `widgets/layout-shell/*` 目前是薄桥接层，减少了直接反向依赖，但没有改变 feature 仍是 workflow owner 的事实。
- `Sidebar.vue`、`NotificationDrawer.vue`、`TopNav.vue` 的大组件体量问题仍然存在，这轮只处理结构边界，不继续做更深层瘦身。
- 独立 review gate 未满足：当前工具策略要求用户显式授权委派后才能拉子 agent reviewer，本轮没有这项授权，因此这里只能归档同上下文自审结果。

## Touched known-debt status

- 已触达的已知债务：`componentFeatureImportAllowlist` 中 layout 相关 3 条。
- 收口结果：已完全清掉这 3 条 direct import 债务。
- 未在本轮收口的相邻债务：layout 组件体量与更深层交互 owner，已继续留在 `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`。
