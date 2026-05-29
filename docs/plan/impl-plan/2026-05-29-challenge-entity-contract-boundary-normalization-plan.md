> 状态：Current
> 事实源：`architectureAllowlist.ts`、`entities/challenge` 当前 model / ui、challenge entity 消费面
> 替代：无

# Challenge Entity Contract Boundary Normalization Plan

## 目标

- 让 `entities/challenge` 从 `@/api/contracts` 脱钩，改为消费本地 entity 展示类型。
- 收掉 `commonForbiddenImportAllowlist` 中 `entities/challenge` 的 9 条例外。

## 非目标

- 本轮不处理 `components/common/InstancePanel.vue`。
- 本轮不改 challenge 相关 API / DTO 本身。
- 本轮不重写 challenge entity 的展示逻辑，只调整 contract owner。

## 输入依据

- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/entities/challenge/model/index.ts`
- `code/frontend/src/entities/challenge/model/presentation.ts`
- `code/frontend/src/entities/challenge/ui/*.vue`

## 当前结论

- `entities/challenge` 当前直接使用 API contract 类型，但真正需要的只是本地展示语义和少量字段。
- 这组边界继续保留 allowlist 没有必要，适合直接改为本地 entity type。

## 设计边界

### `entities/challenge/model` 本轮负责

- category / difficulty / status / instance sharing 的本地展示类型
- category / difficulty / status label / color / normalize helper

### `entities/challenge/ui` 本轮负责

- challenge entity 组件真正依赖的最小字段接口
- 不再直接 import `@/api/contracts`

### feature / page 消费面本轮约定

- 继续直接把结构兼容的 DTO 或 DTO 子集传入 entity UI
- 不新增 feature 级 adapter

## 任务切片

### Slice 1：本地 challenge entity type

- 目标：
  - 在 `entities/challenge/model` 内定义本地展示类型
  - `presentation.ts` 改为只依赖本地类型
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 类型边界是否真的回到 entity 内
  - 没有把 API DTO owner 换个文件名继续留在 entity 层

### Slice 2：entity UI contract 切换

- 目标：
  - `entities/challenge/ui/*` 改为使用本地展示类型
  - 删除 allowlist 对应 9 条例外
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 外部消费面是否不需要额外适配
  - allowlist 是否真实下降

### Slice 3：文档 / review 收尾

- 目标：
  - 更新 backlog 与本轮 review 文档
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这组组件依赖结构兼容的 DTO 输入；如果某些调用方暗含更宽松的空值语义，可能需要在类型里显式保留可选字段。
- review 默认仍是同上下文 self-review，独立 reviewer gate 需要继续说明。

## 实施记录

- [x] Slice 1：已新增 `entities/challenge/model/types.ts`，并让 `presentation.ts` 改为使用本地 `ChallengeCategory`、`ChallengeDifficulty`、`ChallengeStatus`、`ChallengeInstanceSharing`。
- [x] Slice 2：`entities/challenge/ui/*` 中命中 allowlist 的 8 个组件已全部切到本地 entity 类型；`commonForbiddenImportAllowlist` 已删掉 challenge entity 对应 9 条例外。
- [x] Slice 3：backlog 与 review 文档已同步记录本轮收口结果。
