> 状态：Current
> 事实源：`AWDChallengeConfigPanel.vue` 当前 owner、`ContestEditWorkspacePanel.vue` 的 workspace shell owner、现有 contest edit feature UI 收口模式
> 替代：无

# AWD Challenge Config Panel Feature UI Normalization Plan

## 目标

- 把 `AWDChallengeConfigPanel.vue` 迁入 `features/platform-contests/ui`。
- 让 `ContestEditWorkspacePanel.vue` 改为通过 feature 内部 UI 组合 AWD 配置阶段。
- 收掉 `architectureAllowlist.ts` 里 `AWDChallengeConfigPanel.vue -> @/features/awd-inspector` 这条历史例外。

## 非目标

- 本轮不拆 `AWDChallengeConfigPanel.vue` 的内部表格、summary card 或 checker 摘要布局。
- 本轮不改 `useAwdCheckResultPresentation()` 的 model owner。
- 本轮不顺手调整 `ContestAwdPreflightPanel.vue`、`PlatformContestFormPanel.vue` 或 `ContestEditWorkspacePanel.vue` 的其它 stage owner。

## 输入依据

- `code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue`
- `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
- `code/frontend/src/features/platform-contests/ui/ContestEditTopbarPanel.vue`
- `code/frontend/src/features/awd-inspector/model/useAwdCheckResultPresentation.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `AWDChallengeConfigPanel.vue` 只服务 contest edit 的 `awd-config` stage，不是独立的 cross-page shared UI。
- `features/awd-inspector` 当前只承接 AWD 结果 presentation / inspector model，不适合作为这块 contest edit stage UI 的目录 owner。
- 最小正确落点是 `features/platform-contests/ui/AWDChallengeConfigPanel.vue`，并让 `ContestEditWorkspacePanel.vue` 走 feature 内部相对 import。

## 设计边界

### `features/platform-contests/ui/*` 本轮负责

- AWD challenge config panel
- contest edit workspace 对 AWD config stage 的 UI 组合

### `features/awd-inspector/model/*` 本轮继续负责

- AWD 校验结果的 label / access URL presentation helper

### `views/platform/ContestEdit.vue` 本轮不负责

- AWD config stage 的局部目录展示 owner

## 任务切片

### Slice 1：panel 落位与 import 收口

- 目标：
  - 新增 `features/platform-contests/ui/AWDChallengeConfigPanel.vue`
  - `ContestEditWorkspacePanel.vue` 改为 feature 内部相对 import
  - `platform-contests/ui/index.ts` 暴露 panel
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- Review focus：
  - panel 是否仍只保留 props / emits contract
  - `ContestEditWorkspacePanel.vue` 是否没有把 AWD config owner 抬回 route view

### Slice 2：护栏与历史例外同步

- 目标：
  - 更新 raw-source / theme token / ui primitive 测试
  - 删除 `architectureAllowlist.ts` 中的旧 allowlist
  - 去掉 feature UI 中显式 `RouterLink` import
- 验证：
  - `npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - feature UI 是否不再直接 import `vue-router`
  - allowlist 是否在 touched surface 内被实际删掉

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `AWDChallengeConfigPanel.vue` 迁位后仍然是偏大的展示壳；如果后续继续在同一 surface 上堆表格、summary 或新动作，再按目录 header / summary / table row detail 继续拆，而不是把它重新放回 `components/`。
