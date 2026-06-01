> 状态：Current
> 事实源：`ContestAwdPreflightPanel.vue` 当前 owner、`platform-contests` 既有 feature-owned UI 收口模式
> 替代：无

# Contest AWD Preflight Panel Feature UI Normalization Plan

## 目标

- 把 `ContestAwdPreflightPanel.vue` 迁入 `features/platform-contests/ui`。
- 让 `ContestEditWorkspacePanel.vue` 改为 feature 内部相对 import。
- 同步更新组件声明、raw-source 测试和 backlog 记录。

## 非目标

- 本轮不迁 `AWDReadinessChecklist.vue`、`AWDReadinessDecisionHUD.vue`、`AWDReadinessSummary.vue`。
- 本轮不改 AWD readiness 数据来源、导航事件或 contest edit stage owner。
- 本轮不顺手处理 `ContestAwdConfigWorkspaceShell.vue` 那组更大颗粒度的 AWD 配置壳体。

## 输入依据

- `code/frontend/src/components/platform/contest/ContestAwdPreflightPanel.vue`
- `code/frontend/src/components/platform/contest/AWDReadinessChecklist.vue`
- `code/frontend/src/components/platform/contest/AWDReadinessDecisionHUD.vue`
- `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase18.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestAwdPreflightPanel.vue` 只服务 `ContestEditWorkspacePanel.vue`，是单一 feature 的 AWD 赛前检查展示面板。
- `AWDReadinessChecklist.vue` / `AWDReadinessDecisionHUD.vue` 当前更像 readiness primitive；这轮只迁 route shell 级 panel，不扩大 touched surface。
- 最小正确落点是 `features/platform-contests/ui/ContestAwdPreflightPanel.vue`。

## 设计边界

### `features/platform-contests/ui/*` 本轮负责

- AWD preflight panel
- contest edit workspace 对 preflight panel 的 feature 内部组合

### `components/platform/contest/*` 本轮继续负责

- `AWDReadinessChecklist.vue`
- `AWDReadinessDecisionHUD.vue`
- 其它 AWD readiness primitive 和 summary surface

### `features/platform-contests/model/*` 本轮继续负责

- contest edit stage 切换
- `retry:preflight`、`navigate:awd-challenge-from-preflight` 等 workflow owner

## 任务切片

### Slice 1：preflight panel 迁位

- 目标：
  - 新增 `features/platform-contests/ui/ContestAwdPreflightPanel.vue`
  - `ContestEditWorkspacePanel.vue` 改为 feature 内部相对 import
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase18.test.ts`
- Review focus：
  - panel props / emits contract 是否保持不变
  - contest edit stage owner 是否仍留在 workspace shell / feature model

### Slice 2：护栏与引用同步

- 目标：
  - 更新 `components.d.ts`
  - 更新 raw-source / theme token 测试引用路径
  - backlog 记录 `platform-contests` 剩余 UI surface 进展
- 验证：
  - `npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 旧 `components/platform/contest/ContestAwdPreflightPanel.vue` 路径是否已经从 touched surface 消失
  - readiness primitive 是否没有被意外拖进 feature root barrel

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase18.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `ContestAwdPreflightPanel.vue` 迁位后仍显式依赖旧目录下的 readiness primitive；如果后续要继续清这条线，应把 `AWDReadinessChecklist.vue` / `AWDReadinessDecisionHUD.vue` 归到更清晰的 AWD readiness owner，而不是在多个 feature 里各自复制一份。
