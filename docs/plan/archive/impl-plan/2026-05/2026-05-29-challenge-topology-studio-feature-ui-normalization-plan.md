> 状态：Current
> 事实源：`architectureAllowlist.ts`、`challenge-topology-studio` 当前 page / model / UI 目录、拓扑工作台 raw-source 护栏
> 替代：无

# Challenge Topology Studio Feature UI Normalization Plan

## 目标

- 清掉 `componentFeatureImportAllowlist` 中剩余的 6 条 topology studio 例外。
- 让 `challenge-topology-studio` 的专属工作台 UI 完整回到 `features/challenge-topology-studio/ui`。

## 非目标

- 本轮不改 `useChallengeTopologyStudioPage.ts` 的 workflow owner。
- 本轮不改拓扑编辑的 API / DTO / route owner。
- 本轮不做新的表现层重设计，只处理目录 owner 与测试边界。

## 输入依据

- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/features/challenge-topology-studio/ui/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts`
- `code/frontend/src/components/platform/topology/*.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/views/__tests__/asyncChunkBoundaries.test.ts`

## 当前结论

- 剩余 allowlist 不是合理长期例外，而是 feature-owned UI 没有迁干净。
- 外部真实消费面只剩 `ChallengeTopologyStudioPage.vue` 与测试，因此可以一次性迁完整组 UI，而不需要保留 bridge。

## 设计边界

### `features/challenge-topology-studio/ui` 本轮负责

- 拓扑 studio page
- challenge / template workbench
- canvas workspace、node editor、network / policy / package / summary / header 等专属工作台 UI

### `features/challenge-topology-studio/model` 本轮继续负责

- draft / graph / layout / selection / mutation / persistence workflow
- page-level async load / save / export / template action owner

### `components/platform/topology` 本轮处理

- 不保留这组 studio 专属 UI 的中间桥接状态
- 迁移完成后，原目录中的对应文件应彻底退场

## 任务切片

### Slice 1：目录 owner 收口

- 目标：
  - 把 topology studio 专属 UI 从 `components/platform/topology/` 迁到 `features/challenge-topology-studio/ui/`
  - 更新 `ChallengeTopologyStudioPage.vue` 与 UI 内部相对依赖
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts`
- Review focus：
  - 没有留下 `components -> features` 反向依赖
  - 没有引入临时 bridge 或双 owner 状态

### Slice 2：护栏与 allowlist 收口

- 目标：
  - 更新 raw-source / theme / async-chunk / 组件类型声明路径
  - 删除 topology 相关 6 条 `componentFeatureImportAllowlist`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/asyncChunkBoundaries.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 护栏是否全部切到新 owner
  - allowlist 是否真实下降且没有 stale entry

### Slice 3：backlog / review 收尾

- 目标：
  - 更新前端 backlog
  - 补本轮 review 文档
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否反映“allowlist 清尾”而不是泛泛描述
  - touched surface 上没有留下已知结构债

## 实施记录

- [x] Slice 1：`components/platform/topology/` 下 19 个 topology studio 专属 `.vue` 文件已整体迁入 `features/challenge-topology-studio/ui/`，`ChallengeTopologyStudioPage.vue` 已改为 feature 内部相对 import。
- [x] Slice 2：`ChallengeTopologyStudio.test.ts`、`sharedThemeTokenAdoption.test.ts`、`workspacePageHeaderStyles.test.ts`、`asyncChunkBoundaries.test.ts` 与 `components.d.ts` 已同步切到新路径；`componentFeatureImportAllowlist` 已清空。
- [x] Slice 3：backlog 已记录 topology studio 这一刀的收口结果，本轮 review 文档已补齐。

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/asyncChunkBoundaries.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 测试里对 raw-source 路径有大量直接引用，迁移时如果漏改一处，主要会在护栏测试里暴露。
- 当前 review 默认仍只能做到同上下文 self-review；独立 reviewer gate 需要继续显式说明。
