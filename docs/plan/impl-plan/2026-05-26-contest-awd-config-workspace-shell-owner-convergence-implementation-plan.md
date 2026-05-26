> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAwdConfig.vue` 当前 AWD 配置工作台壳、contest-awd-config feature 已存在的数据与动作 owner
> 替代：无

# ContestAwdConfig Workspace Shell Owner Convergence Implementation Plan

## 目标

- 把 `ContestAwdConfig.vue` 里的 AWD 配置工作台页面壳、目录区、编辑区、调试区、底部动作区和配套局部样式抽到独立 `ContestAwdConfigWorkspaceShell.vue`。
- 保持父页继续持有 `useContestAwdConfigPage()` 的路由、服务选择、预览、保存、错误、草稿和摘要 owner。
- 让 `ContestAwdConfig.vue` 回到 route page 组合 owner，不再承接大块工作台模板和局部样式。

## 非目标

- 本轮不改 `useContestAwdConfigPage()` 的 API、预览保存流程、checker draft 语义或服务选择逻辑。
- 本轮不改已有子组件 `ContestAwdConfigTopbar`、`ContestAwdServiceDirectory`、`ContestAwdDebugStation`、`ContestAwdConfigFooter` 的内部行为。
- 本轮不新增 AWD 配置业务能力。

## 输入依据

- `code/frontend/src/views/platform/ContestAwdConfig.vue`
- `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts`
- `.harness/reuse-decisions/contest-awd-config-workspace-shell-owner-convergence.md`

## 当前结论

- `ContestAwdConfig.vue` 当前 686 行，已经是剩余 oversized route view allowlist 的最后一页，但 owner 基本集中在 `useContestAwdConfigPage()`。
- 当前剩余重量主要来自稳定的页面模板壳和局部样式，适合继续沿用“父页保留 owner，子组件承接 workspace shell”的既有模式。

## 任务切片

### Slice 1：抽取 contest awd config workspace shell

- 目标：
  - 新增 `ContestAwdConfigWorkspaceShell.vue`，承接 AWD 配置工作台页面壳、目录区、编辑区、调试区、底部动作区和对应局部样式。
  - `ContestAwdConfig.vue` 继续保留 contest/service route owner、预览/保存动作、草稿状态和错误 owner。
- 预期改动：
  - `code/frontend/src/views/platform/ContestAwdConfig.vue`
  - `code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/platform/__tests__/ContestAwdConfig.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 父页是否仍然持有路由、服务选择、预览、保存和草稿 owner
  - 新组件是否只承接稳定工作台壳，而没有重新吸入 API 或动作逻辑
  - 路由页是否脱离 oversized allowlist

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 `ContestAwdConfig.vue` 从 oversized route view 推进到 workspace shell owner 收口的事实写回前端 review 索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否明确这是 route view 壳层收口，而不是 feature owner 迁移

## 风险

- `ContestAwdConfig` 的 props 面较宽，如果透传设计不清，容易变成“父页和 shell 双方都在看似拥有同一份 AWD 编辑状态”。
- 这一页虽然已经有多个子组件，但 view 自身仍保留大量模板分支；抽壳时必须避免把已有子组件之间的组合关系打散。

## 回退方式

- 如抽取后出现交互回归，可回退 `ContestAwdConfigWorkspaceShell.vue` 并把模板恢复到 `ContestAwdConfig.vue`。
- 本轮只影响前端视图层、测试护栏和 review 文档，不涉及后端或 API 契约。
