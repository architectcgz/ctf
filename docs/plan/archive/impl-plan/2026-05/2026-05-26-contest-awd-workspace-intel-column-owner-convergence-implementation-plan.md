> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前结构、AWD 学员战场既有 `awd/*` 子组件分层模式
> 替代：无

# Contest AWD Workspace Intel Column Owner Convergence Implementation Plan

## 目标

- 让 `ContestAWDWorkspacePanel.vue` 把右侧 intelligence rail 从父组件中抽成独立子组件。
- 保持父组件继续拥有 `useContestAWDWorkspace`、挑战选择、目标筛选、攻击提交、SSH/服务动作与 toast 行为。
- 让新子组件只承接“战场情报 / 最近战报”展示模板和局部样式 owner，不接管远端请求、路由状态或业务动作。

## 非目标

- 本轮不改 `ContestDetail.vue` 的 tab 切换、路由入口或页面级 owner。
- 本轮不处理 `ContestAWDWorkspacePanel.vue` 中间攻击区和左侧防守区。
- 本轮不引入新的 feature、store、route 或 API wrapper。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `docs/reviews/general/2026-05-05-awd-defense-content-page-review.md`
- `docs/reviews/architecture/2026-05-06-awd-defense-workspace-review.md`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`

## 当前结论

- `ContestAWDWorkspacePanel.vue` 当前约 `1161` 行，仍是 `TD-1` backlog 中的活跃超大组件。
- 防守内容页、SSH 连接面板和服务目录已经拆出独立 owner，说明该页面适合继续按稳定区块抽层，而不是整页重写。
- 当前最稳定的继续切片是右侧 `column-intel`：
  - “战场情报”只消费 `scoreboardRows` 与 `myTeam`
  - “最近战报”只消费 `workspace?.recent_events` 和已有格式化 helper
  - 不直接触发远端请求、路由同步、复制、重启或攻击提交流程

## 任务切片

### Slice 1：抽出 intelligence rail 子组件

- 目标：
  - 新建 `AWDWorkspaceIntelColumn.vue`，承接右侧“战场情报 / 最近战报”两块模板与样式。
  - `ContestAWDWorkspacePanel.vue` 只保留 props 计算、helper 和组件装配，不再内联右侧两块模板。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- 组件边界：
  - 父组件继续拥有：
    - `scoreboardRows`
    - `myTeam`
    - `workspace?.recent_events`
    - `getChallengeTitleForEvent`
    - `eventDirectionLabel`
    - `eventResultLabel`
    - `formatServiceRef`
  - 子组件只接收现成数据和 formatter，不新增异步逻辑或 local workflow owner。
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/ContestDetail.test.ts`
- Review focus：
  - 父组件是否真正退回到组合 owner，而不是把局部样式和重复 helper 留在原处。
  - 新子组件是否保持纯展示，不吸入提交、刷新或路由行为。

### Slice 2：回写 TD-1 当前事实

- 目标：
  - 把本轮 `ContestAWDWorkspacePanel.vue` 的切片进展写回前端主索引，降低后续重复扫描成本。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ContestAWDWorkspacePanel|intelligence|战场情报|最近战报" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否明确这是 “已收口的 touched surface” 还是“仍保留的 backlog”，避免把旧结论继续留成现役状态。

## 风险

- `contestAwdWorkspacePanelSource.test.ts` 当前直接对源码字符串做结构断言；抽层后需要把断言改成“父组件挂载了新子组件 + 关键文案存在于新子组件”，否则会误判为回归。
- 右侧 intelligence rail 当前和父组件共用部分局部 class；若抽层时遗漏 scoped style，会造成 UI 回归。
- 若顺手把攻击区或防守区也一起抽，会把这轮从“稳定展示区块收口”扩大成主业务链重构，超出当前切片边界。

## 回退方式

- 如 intelligence rail 抽层引入回归，可回退 `AWDWorkspaceIntelColumn.vue` 并恢复 `ContestAWDWorkspacePanel.vue` 的右侧内联模板。
- 因本轮不涉及 API、route、query、feature model 或提交动作，回退只影响前端组件层与相关测试，不涉及数据迁移。
