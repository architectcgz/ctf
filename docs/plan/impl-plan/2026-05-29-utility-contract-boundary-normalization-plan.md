> 状态：Current
> 事实源：`architectureAllowlist.ts`、3 个 `utils/*.ts` 文件及其现有 consumer
> 替代：无

# Utility Contract Boundary Normalization Plan

## 目标

- 让 `utils/contest.ts`、`utils/platformContestAwdChallengeLinks.ts`、`utils/skillProfile.ts` 从 `@/api/contracts` 脱钩。
- 清空 `utilityBoundaryImportAllowlist`。

## 非目标

- 本轮不调整这些 utility 的 consumer import 路径。
- 本轮不新建 `entities/contest`、`entities/skill-profile` 等更大结构。
- 本轮不处理 `featureRouterImportAllowlist` 和 `composableMultiBoundaryAllowlist`。

## 输入依据

- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/utils/contest.ts`
- `code/frontend/src/utils/platformContestAwdChallengeLinks.ts`
- `code/frontend/src/utils/skillProfile.ts`
- 对应 API wrapper / feature consumer

## 当前结论

- 这 3 个 utility 的问题不是功能位置错误，而是 contract owner 还停留在 `@/api/contracts`。
- 最小可审查改动是把它们的最小字段接口和枚举语义收回到文件本地，不扩大到 consumer 改路由式迁移。

## 设计边界

### `utils/contest.ts` 本轮负责

- contest status / mode 的本地展示类型与文案 helper

### `utils/platformContestAwdChallengeLinks.ts` 本轮负责

- AWD service 到 challenge link 的本地映射输入 / 输出 contract

### `utils/skillProfile.ts` 本轮负责

- skill profile / recommendation 的本地归一化 contract
- radar / weak dimension 的展示辅助

### consumer 本轮约定

- 继续通过结构兼容对象消费这些 utility
- 不改现有 import 路径

## 任务切片

### Slice 1：contest / skill profile contract 本地化

- 目标：
  - `contest.ts` 与 `skillProfile.ts` 改为本地最小类型
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 展示 helper 的类型语义是否仍与现有 consumer 兼容

### Slice 2：AWD challenge link mapper contract 本地化

- 目标：
  - `platformContestAwdChallengeLinks.ts` 改为本地输入 / 输出 contract
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - AWD challenge link 输出字段是否仍覆盖下游需要的所有字段

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 清空 `utilityBoundaryImportAllowlist`
  - 同步 backlog 与 review 文档
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/utility-contract-boundary-normalization.md docs/plan/impl-plan/2026-05-29-utility-contract-boundary-normalization-plan.md docs/reviews/frontend/2026-05-29-utility-contract-boundary-normalization-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/utils/contest.ts code/frontend/src/utils/platformContestAwdChallengeLinks.ts code/frontend/src/utils/skillProfile.ts code/frontend/src/__tests__/architectureAllowlist.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `skillProfile.ts` 和 `platformContestAwdChallengeLinks.ts` 仍属于 product-specific utility；本轮只解决 contract owner，不额外做目录归属重排。
- 这轮 review 默认仍是同上下文 self-review，独立 reviewer gate 仍需单独说明。

## 实施记录

- [x] Slice 1：`utils/contest.ts` 与 `utils/skillProfile.ts` 已改为本地最小类型；`sharedThemeTokenAdoption.test.ts` 也补上了 `skillProfile` 的 `@/api/contracts` 负向断言。
- [x] Slice 2：`utils/platformContestAwdChallengeLinks.ts` 已改为本地 AWD service / challenge link contract，不再直接依赖 API DTO owner。
- [x] Slice 3：`utilityBoundaryImportAllowlist` 已清空，backlog 与 review 文档已同步。
