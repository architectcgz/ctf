> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestDetail.vue` 当前剩余模板结构、contest 已抽出的 overview / challenge workspace 组件
> 替代：无

# Contest Detail Shell Owner Convergence Implementation Plan

## 目标

- 把 `ContestDetail.vue` 剩余的公告 section、队伍 section、队伍对话框壳抽到独立 `components/contests` 组件。
- 保持父页继续持有 route/query 同步、tab owner、远端数据、创建/加入/踢出队伍动作和异常处理。
- 让 `ContestDetail.vue` 进一步收敛到 route page 组合 owner，而不是继续保留大量局部模板壳。

## 非目标

- 本轮不改 `useContestDetailRoutePage` 的异步流、状态 shape 或路由协议。
- 本轮不改 `ContestOverviewPanel.vue`、`ContestChallengeWorkspacePanel.vue`、`ContestAWDWorkspacePanel.vue` 的业务逻辑。
- 本轮不改竞赛详情的视觉方向和主题变量体系。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/components/contests/ContestAnnouncementsPanel.vue`
- `code/frontend/src/components/contests/ContestTeamPanel.vue`
- `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/views/contests/__tests__/contestDetailPanelExtraction.test.ts`
- `.harness/reuse-decisions/contest-detail-shell-owner-convergence.md`

## 当前结论

- `ContestDetail.vue` 已经完成 overview 和普通题目工作区抽取，当前剩余重量不在主业务 flow，而在仍然留在路由页里的 section shell 和 dialog shell。
- 这些壳层没有新的页面级 owner 语义，继续留在父页只会维持大模板而不会提升可读性。
- 最小安全切片是：只迁移壳层，不迁移 route model、请求、表单动作或队伍工作流状态。

## 任务切片

### Slice 1：收回公告 / 队伍 / 队伍对话框壳

- 目标：
  - `ContestAnnouncementsWorkspaceSection.vue` 持有公告 tab 的 section heading 与数量提示壳。
  - `ContestTeamWorkspaceSection.vue` 持有队伍 tab 的 section heading、当前队伍摘要和 `ContestTeamPanel` 装配壳。
  - `ContestTeamDialogs.vue` 持有创建/加入队伍两组 `CFocusedInputDialog` 壳。
- 预期改动：
  - `code/frontend/src/views/contests/ContestDetail.vue`
  - `code/frontend/src/components/contests/ContestAnnouncementsWorkspaceSection.vue`
  - `code/frontend/src/components/contests/ContestTeamWorkspaceSection.vue`
  - `code/frontend/src/components/contests/ContestTeamDialogs.vue`
- 验证：
  - `git diff --check -- code/frontend/src/views/contests/ContestDetail.vue code/frontend/src/components/contests/ContestAnnouncementsWorkspaceSection.vue code/frontend/src/components/contests/ContestTeamWorkspaceSection.vue code/frontend/src/components/contests/ContestTeamDialogs.vue`
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/contestDetailPanelExtraction.test.ts src/views/contests/__tests__/ContestDetail.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 父页是否继续持有 tab owner、远端数据和主动作 owner
  - 新子组件是否只承接 section / dialog 壳，而没有吸入 route 或请求逻辑

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 `ContestDetail.vue` 这一刀从“继续减重”推进到“剩余 shell owner 收口”的进展写回前端 review 索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ContestDetail.vue|队伍对话框|公告 section|队伍 section" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否明确这是 route page shell 收口，而不是把 route owner 错迁到子组件

## 风险

- 队伍对话框抽走后，`v-model` 风格的打开状态和 `close` 行为容易出现第二份 owner。
- 公告 / 队伍 tab section 抽走后，源码护栏如果还盯父页 raw source，容易把合理抽取误判成回归。

## 回退方式

- 如抽取后出现交互回归，可回退对应壳组件，把模板恢复到 `ContestDetail.vue`。
- 本轮只影响 contest 前端视图层、测试护栏和 review 文档，不涉及 API、权限或服务端逻辑。
