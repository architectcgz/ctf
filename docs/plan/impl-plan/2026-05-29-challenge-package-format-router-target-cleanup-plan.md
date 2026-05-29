> 状态：Current
> 事实源：challenge package format route view、challenge package import feature model、前端架构 allowlist
> 替代：无

# Challenge Package Format Router Target Cleanup Plan

## 目标

- 把 `useChallengePackageFormatPage.ts` 从 route-aware push wrapper 收口为纯 route target helper。
- 让 `ChallengePackageFormat.vue` 通过 `RouterLink` 消费返回导入页的导航目标。
- 删除 `useChallengePackageFormatPage.ts -> vue-router` allowlist。

## 非目标

- 不改上传示例页的内容、文案和样式。
- 不处理 `useChallengeImportManagePage.ts` 或 `useChallengeImportPreviewPage.ts`。

## 输入依据

- `code/frontend/src/features/challenge-package-import/model/useChallengePackageFormatPage.ts`
- `code/frontend/src/features/challenge-package-import/model/index.ts`
- `code/frontend/src/features/challenge-package-import/index.ts`
- `code/frontend/src/views/platform/ChallengePackageFormat.vue`
- `code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- 这条 feature allowlist 当前只支撑一次 `router.push`，价值过低。
- route view 不能直接 `useRouter()`，但可以使用 `RouterLink`。
- 最小正确改动是保留 feature 的 route target contract，不再保留 route-aware composable。

## 设计边界

### `useChallengePackageFormatPage.ts` 本轮负责

- 暴露返回导入页的纯 route target contract

### `ChallengePackageFormat.vue` 本轮负责

- 通过 `RouterLink` 渲染返回入口
- 继续作为上传示例页的薄 route shell

## 任务切片

### Slice 1：feature helper 去掉 `vue-router`

- 目标：
  - `useChallengePackageFormatPage.ts` 改成纯 route target helper
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- Review focus：
  - helper 是否不再持有 `useRouter` / `router.push`

### Slice 2：route view 切到 `RouterLink`

- 目标：
  - `ChallengePackageFormat.vue` 改为直接消费纯 target contract
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- Review focus：
  - 视图是否保持薄壳，不重新引入 `useRouter`

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 删除 allowlist、补 backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengePackageFormat.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - 这条 allowlist 是否被真正删除

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/challenge-package-format-router-target-cleanup.md docs/plan/impl-plan/2026-05-29-challenge-package-format-router-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-challenge-package-format-router-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/challenge-package-import/model/useChallengePackageFormatPage.ts code/frontend/src/features/challenge-package-import/model/index.ts code/frontend/src/features/challenge-package-import/index.ts code/frontend/src/views/platform/ChallengePackageFormat.vue code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收一条极薄的 router push wrapper，不继续动 challenge package import 其余 page owner。
