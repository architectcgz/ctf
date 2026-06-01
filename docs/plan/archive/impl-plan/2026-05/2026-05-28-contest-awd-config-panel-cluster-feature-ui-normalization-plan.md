> 状态：Current
> 事实源：`ContestAwdConfigWorkspaceShell.vue` 当前 owner、AWD config panel cluster 当前依赖关系
> 替代：无

# Contest AWD Config Panel Cluster Feature UI Normalization Plan

## 目标

- 把 AWD 配置 workspace shell 当前依赖的 panel / fields / UI types cluster 迁入 `features/contest-awd-config/ui`。
- 让 `ContestAwdConfigWorkspaceShell.vue` 改为 feature 内部相对 import。
- 同步更新组件声明、raw-source 测试和 backlog 记录。

## 非目标

- 本轮不改 `features/contest-awd-config/model/*` 的 draft、preview、save、load owner。
- 本轮不触碰 `awdCheckerConfigSupport.ts` 等已经位于 `features/contest-awd-config/model` 的能力。
- 本轮不重写 AWD 配置页的布局、交互或 checker 参数表单逻辑。

## 输入依据

- `code/frontend/src/features/contest-awd-config/ui/ContestAwdConfigWorkspaceShell.vue`
- `code/frontend/src/components/platform/contest/ContestAwdConfigTopbar.vue`
- `code/frontend/src/components/platform/contest/ContestAwdConfigFooter.vue`
- `code/frontend/src/components/platform/contest/ContestAwdDebugStation.vue`
- `code/frontend/src/components/platform/contest/ContestAwdEditorHeader.vue`
- `code/frontend/src/components/platform/contest/ContestAwdScoreWeights.vue`
- `code/frontend/src/components/platform/contest/ContestAwdServiceDirectory.vue`
- `code/frontend/src/components/platform/contest/ContestAwdCheckerConfigSection.vue`
- `code/frontend/src/components/platform/contest/ContestAwdHttpStandardFields.vue`
- `code/frontend/src/components/platform/contest/ContestAwdLegacyProbeFields.vue`
- `code/frontend/src/components/platform/contest/ContestAwdScriptCheckerFields.vue`
- `code/frontend/src/components/platform/contest/ContestAwdTcpStandardFields.vue`
- `code/frontend/src/components/platform/contest/contestAwdConfigTypes.ts`
- `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- AWD 配置 route shell 已经归 `features/contest-awd-config/ui` 持有。
- 当前 panel / fields / UI types cluster 只被 AWD 配置 feature 消费，没有 shared consumer。
- 最小正确落点是把这组 cluster 一并迁到 `features/contest-awd-config/ui`，而不是只迁最外层 7 个 panel。

## 设计边界

### `features/contest-awd-config/ui/*` 本轮负责

- workspace shell
- topbar / footer / debug station / header / score weights / service directory / checker config section
- checker config section 依赖的 HTTP/TCP/script/legacy fields
- 仅供这组 UI 使用的 `contestAwdConfigTypes.ts`

### `features/contest-awd-config/model/*` 本轮继续负责

- draft hydration / save / preview / load
- checker config build rules 和 labels
- route / query / data fetch owner

### `components/platform/contest/*` 本轮不再负责

- AWD config panel cluster 本身

## 任务切片

### Slice 1：panel cluster 迁位

- 目标：
  - 迁移 7 个 panel、4 个 fields 子件和 `contestAwdConfigTypes.ts`
  - `ContestAwdConfigWorkspaceShell.vue` 改为 feature 内部相对 import
  - 保持 fields 和 types 的相对依赖自洽
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- Review focus：
  - shell 继续只组合 UI，不回流 model owner
  - checker config section / fields 的 props contract 保持不变

### Slice 2：护栏与引用同步

- 目标：
  - 更新 `components.d.ts`
  - 更新 `ui/index.ts`
  - 更新 AWD config raw-source 测试
  - backlog 记录 AWD config feature UI cluster 收口进展
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 旧 `components/platform/contest/*` 这组 AWD config panel 路径是否已从 touched surface 消失
  - `contest-awd-config` public / internal import 方向是否仍清楚

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只做 AWD config UI owner 收口，不继续拆 `ContestAwdConfigWorkspaceShell.vue` 的布局或逻辑 owner；如果后续这页继续增长，下一刀应继续按 capability 拆 shell 内部职责，而不是再把 panel 压回 `components/`。
