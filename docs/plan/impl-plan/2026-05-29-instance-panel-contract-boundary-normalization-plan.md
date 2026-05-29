> 状态：Current
> 事实源：`architectureAllowlist.ts`、`InstancePanel.vue`、`InstancePanel.test.ts`
> 替代：无

# Instance Panel Contract Boundary Normalization Plan

## 目标

- 让 `components/common/InstancePanel.vue` 从 `@/api/contracts` 脱钩，改为消费本地最小展示类型。
- 清空 `commonForbiddenImportAllowlist`。

## 非目标

- 本轮不改 `features/instance-list` 的 workflow owner。
- 本轮不统一 challenge detail / instance list / instance panel 的全部 instance 展示 helper。
- 本轮不调整 `InstancePanel` 的交互行为和视觉结构。

## 输入依据

- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components/common/InstancePanel.vue`
- `code/frontend/src/components/common/__tests__/InstancePanel.test.ts`
- `code/frontend/src/features/instance-list/model/useInstanceListPage.ts`

## 当前结论

- `InstancePanel.vue` 属于 shared/common 层，直接引用 API DTO owner 不合理。
- 这条例外已经收敛到单文件，适合用本地最小 contract 一次性收口，不保留中间 bridge。

## 设计边界

### `components/common/instancePanel.types.ts` 本轮负责

- `InstancePanel` 真实需要的状态、分享范围和最小展示字段

### `InstancePanel.vue` 本轮负责

- 面板倒计时、状态标签、expiring soon 提示等本地展示逻辑
- 不再直接 import `@/api/contracts`

### 调用方本轮约定

- 继续直接传结构兼容对象
- 不新增 feature / entity adapter

## 任务切片

### Slice 1：本地 panel contract

- 目标：
  - 为 `InstancePanel` 新增本地类型文件
  - 组件 props / emits 切到本地 contract
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/common/__tests__/InstancePanel.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - common 组件的 contract owner 是否已经本地化
  - 没有把 API DTO 仅仅换个文件继续透传

### Slice 2：allowlist / backlog / review 收尾

- 目标：
  - 删除 `commonForbiddenImportAllowlist` 最后 1 条
  - 同步 backlog 与 review 文档
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/common/__tests__/InstancePanel.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/instance-panel-contract-boundary-normalization.md docs/plan/impl-plan/2026-05-29-instance-panel-contract-boundary-normalization-plan.md docs/reviews/frontend/2026-05-29-instance-panel-contract-boundary-normalization-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/components/common/InstancePanel.vue code/frontend/src/components/common/instancePanel.types.ts code/frontend/src/components/common/__tests__/InstancePanel.test.ts code/frontend/src/__tests__/architectureAllowlist.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `InstancePanel` 目前保留本地状态标签映射；后续如果 instance 展示 helper 继续分散增长，可以再单独评估是否收敛到 `entities/instance` 或现有 instance feature 的共享表示层。
- 这轮 review 默认仍是同上下文 self-review，独立 reviewer gate 仍需单独说明。

## 实施记录

- [x] Slice 1：已新增 `components/common/instancePanel.types.ts`，`InstancePanel.vue` 改为消费本地 `InstancePanelItem` / `InstancePanelStatus`。
- [x] Slice 2：`commonForbiddenImportAllowlist` 已清空，`InstancePanel.test.ts` 与 backlog / review 文档已同步更新。
