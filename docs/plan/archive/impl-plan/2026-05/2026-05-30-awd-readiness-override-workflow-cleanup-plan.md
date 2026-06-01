> 状态：Current
> 事实源：`contest-awd-admin` readiness override workflow 调用链、AWD panel / composable 测试
> 替代：无

# AWD Readiness Override Workflow Cleanup Plan

## 目标

- 继续收口 `useAwdReadinessDecision.ts` 的更深层 workflow owner
- 避免 readiness 摘要刷新失败时把 override dialog 打开链路直接炸穿
- 让 override action 执行逻辑从大段条件分支收口成更明确的 executor

## 非目标

- 本轮不继续调整 runtime stage / round gate / auto refresh policy
- 本轮不变更 `AWDReadinessOverrideDialog.vue` 的 props / emits 契约
- 本轮不重排 `usePlatformContestAwd.ts` 的组合结构

## 输入依据

- `code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue`
- `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`

## 当前结论

- `useAwdReadinessDecision.ts` 已经是 readiness summary、override dialog state、override execute 的唯一 owner，边界基本正确。
- 但 `openOverrideDialog()` 目前把 refresh 与 open 绑死，且对 refresh failure 没有本地错误处理。
- `confirmOverrideAction()` 中的 `create_round` / `run_current_round_check` override 执行逻辑已经开始表现出“action executor 与 dialog state owner 混写”的趋势。

## 设计边界

### `useAwdReadinessDecision.ts` 本轮负责

- readiness summary refresh
- override dialog state
- override action execute / retry
- open dialog 之前的 readiness snapshot refresh failure handling

### `useAwdReadinessDecision.ts` 本轮不负责

- runtime stage / selected round / auto refresh 规则
- round create / round check 常规 mutation owner
- override dialog 的视图表现和输入控件

## 任务切片

### Slice 1：收口 readiness override workflow

- 更新：
  - `useAwdReadinessDecision.ts`
- 目标：
  - 给 `openOverrideDialog()` 增加 refresh failure 的本地错误兜底
  - 把 `confirmOverrideAction()` 的 action 分支抽成明确的 override executor

### Slice 2：同步测试与 backlog / review

- 更新：
  - `usePlatformContestAwd.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-30-awd-readiness-override-workflow-cleanup-review.md`
- 目标：
  - 覆盖 refresh failure 时不应错误打开 override dialog
  - 保持已有 create round / current round check override success path

## 验证

- `cd code/frontend && npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 当前 action 数量仍只有两条，本轮只把 executor owner 明确化，不额外引入更重的 command registry 结构。
- 如果后续 override action 再增长到三条以上，可能需要把 action executor 进一步收成独立 contract / helper。
