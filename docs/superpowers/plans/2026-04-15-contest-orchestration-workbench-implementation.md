# Contest Orchestration Workbench Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `ContestEdit` 升级成单工作台的竞赛编排页，让管理员在同一条链路里完成基础信息、统一题目池、AWD 配置、赛前检查和轮次运行。

**Architecture:** 以前端重组为主，继续复用现有 `contest_challenges`、`AWDReadiness` 和 `AWDOperationsPanel` 能力，不新增第二个 AWD 题目池，也不改后端核心模型。实现上先把 `ContestEdit` 从双标签页升级为工作台容器，再逐步把挂题、AWD 配置和 readiness 以阶段视图方式接入，并保留现有 API 契约与大部分 composable。

**Tech Stack:** Vue 3, TypeScript, Vite, Vitest, Vue Test Utils, existing admin API layer, existing AWD admin composables

---

## Execution Notes

- 前端实现遵循 `@superpowers:test-driven-development`，先补页面与交互测试，再做最小实现。
- 工作台结构与样式必须继续贴合 `@ctf-ui-theme-system`，不要引入新的页面级视觉语言。
- 本轮优先最小可行改动，避免重写 `useAdminContestAWD` 的全部职责；只在确实挡路时拆分小型 composable。
- 声称“完成”前必须执行 `@superpowers:verification-before-completion`，至少覆盖相关单测与类型检查。

## Planned File Map

### Workbench shell / route orchestration

- Modify: `code/frontend/src/views/admin/ContestEdit.vue`
  - 把当前 `基础设置 / 题目编排` 双标签升级为工作台阶段容器，并按赛事模式控制阶段显示。
- Create: `code/frontend/src/components/admin/contest/ContestWorkbenchStageRail.vue`
  - 承载工作台顶部阶段切换、阶段状态、阶段提示和 URL tab 同步。
- Create: `code/frontend/src/components/admin/contest/ContestWorkbenchSummaryStrip.vue`
  - 在页头下展示当前赛事状态、模式、题目数、准备度等统一摘要。

### Contest challenge pool

- Modify: `code/frontend/src/components/admin/contest/ContestChallengeOrchestrationPanel.vue`
  - 从轻量挂题面板演进为统一题目池，增加 AWD 摘要列、筛选和批量视图入口。
- Modify: `code/frontend/src/components/admin/contest/ContestChallengeEditorDialog.vue`
  - 保持基础挂题职责，补齐和工作台阶段文案一致的交互。
- Create: `code/frontend/src/composables/useContestChallengePool.ts`
  - 收敛题目池列表、筛选、排序视图和摘要统计，避免把页面逻辑继续堆进单文件组件。

### AWD configuration / readiness / running handoff

- Modify: `code/frontend/src/components/admin/contest/AWDChallengeConfigPanel.vue`
  - 改成工作台里的 `AWD 配置` 阶段，支持从统一题目池上下文进入，并强调连续逐题编辑。
- Modify: `code/frontend/src/components/admin/contest/AWDReadinessSummary.vue`
  - 从摘要卡演进为 `赛前检查` 主体，展示总体准备度、阻塞清单和快捷返回动作。
- Create: `code/frontend/src/components/admin/contest/ContestAwdPreflightPanel.vue`
  - 承载 `赛前检查` 阶段容器，把 readiness 摘要、阻塞列表和开赛动作整合到单面板。
- Modify: `code/frontend/src/components/admin/contest/AWDOperationsPanel.vue`
  - 改成工作台最后一段使用，未开赛时弱化显示，开赛后成为运行态入口。
- Create: `code/frontend/src/composables/useContestWorkbench.ts`
  - 管理工作台阶段列表、阶段可见性、默认聚焦逻辑，以及开赛后跳转到运行阶段的行为。

### Tests

- Modify: `code/frontend/src/views/admin/__tests__/ContestEdit.test.ts`
  - 覆盖工作台阶段结构、模式差异和 AWD 开赛后的阶段切换。
- Modify: `code/frontend/src/components/admin/__tests__/AWDOperationsPanel.test.ts`
  - 覆盖工作台运行阶段下的未开赛弱化状态与已开赛行为。
- Modify: `code/frontend/src/components/admin/__tests__/AWDReadinessSummary.test.ts`
  - 覆盖赛前检查阻塞项与快捷入口。
- Create: `code/frontend/src/components/admin/__tests__/ContestChallengeOrchestrationPanel.test.ts`
  - 覆盖统一题目池的 AWD 摘要列、筛选与摘要统计。
- Create: `code/frontend/src/components/admin/__tests__/ContestWorkbenchStageRail.test.ts`
  - 覆盖阶段显隐、模式切换和键盘导航。
- Create: `code/frontend/src/composables/__tests__/useContestWorkbench.test.ts`
  - 覆盖工作台阶段计算、默认焦点和模式边界。

## Task 1: 锁定工作台骨架的 RED 测试

**Files:**
- Modify: `code/frontend/src/views/admin/__tests__/ContestEdit.test.ts`
- Create: `code/frontend/src/components/admin/__tests__/ContestWorkbenchStageRail.test.ts`
- Create: `code/frontend/src/composables/__tests__/useContestWorkbench.test.ts`

- [ ] **Step 1: 给 `ContestEdit` 增加工作台阶段断言**

在 `ContestEdit.test.ts` 增加至少 3 个用例：

- `应该在普通赛下只展示基础信息与题目池阶段`
- `应该在 AWD 赛事下展示基础信息、题目池、AWD 配置、赛前检查与轮次运行`
- `应该在 AWD 赛事已开赛时默认聚焦轮次运行阶段`

最小断言示例：

```ts
expect(wrapper.text()).toContain('题目池')
expect(wrapper.text()).toContain('赛前检查')
expect(wrapper.text()).toContain('轮次运行')
expect(wrapper.text()).not.toContain('AWD 配置')
```

- [ ] **Step 2: 给阶段 rail 单测锁定结构和导航**

在 `ContestWorkbenchStageRail.test.ts` 补至少 2 个用例：

- `应该按传入阶段渲染按钮并高亮当前阶段`
- `应该跳过 disabled 阶段并保持 roving tabindex`

最小断言示例：

```ts
expect(wrapper.get('[role=\"tab\"][aria-selected=\"true\"]').text()).toContain('题目池')
expect(wrapper.findAll('[role=\"tab\"]')).toHaveLength(5)
```

- [ ] **Step 3: 给 `useContestWorkbench` 锁定模式与默认阶段逻辑**

在 `useContestWorkbench.test.ts` 增加：

- `jeopardy 模式仅返回基础阶段`
- `awd + running 状态默认阶段为 operations`
- `awd + registering 状态默认阶段为 basics 或 pool`

最小断言示例：

```ts
expect(result.visibleStages.map((item) => item.key)).toEqual(['basics', 'pool'])
expect(result.defaultStage).toBe('operations')
```

- [ ] **Step 4: 运行定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/spec-contest-orchestration-workbench/code/frontend
npx vitest run src/views/admin/__tests__/ContestEdit.test.ts src/components/admin/__tests__/ContestWorkbenchStageRail.test.ts src/composables/__tests__/useContestWorkbench.test.ts
```

预期：FAIL，失败点集中在新阶段结构和新 composable 尚不存在。

- [ ] **Step 5: 提交工作台骨架测试基线**

```bash
git add code/frontend/src/views/admin/__tests__/ContestEdit.test.ts code/frontend/src/components/admin/__tests__/ContestWorkbenchStageRail.test.ts code/frontend/src/composables/__tests__/useContestWorkbench.test.ts
git commit -m "test(竞赛): 补充工作台阶段骨架断言"
```

## Task 2: 实现工作台容器与阶段 rail

**Files:**
- Modify: `code/frontend/src/views/admin/ContestEdit.vue`
- Create: `code/frontend/src/components/admin/contest/ContestWorkbenchStageRail.vue`
- Create: `code/frontend/src/components/admin/contest/ContestWorkbenchSummaryStrip.vue`
- Create: `code/frontend/src/composables/useContestWorkbench.ts`
- Modify: `code/frontend/src/views/admin/__tests__/ContestEdit.test.ts`
- Create: `code/frontend/src/components/admin/__tests__/ContestWorkbenchStageRail.test.ts`
- Create: `code/frontend/src/composables/__tests__/useContestWorkbench.test.ts`

- [ ] **Step 1: 先写 `useContestWorkbench` 最小实现**

输出至少包含：

```ts
export type ContestWorkbenchStageKey =
  | 'basics'
  | 'pool'
  | 'awd-config'
  | 'preflight'
  | 'operations'

export function useContestWorkbench(contest: Readonly<Ref<ContestDetailData | null>>) {
  // return { visibleStages, defaultStage, summaryItems }
}
```

- [ ] **Step 2: 实现阶段 rail 组件**

要求：

- 使用现有 tab 语义
- 接受 `stages`、`activeStage`、`selectStage`
- disabled 阶段不可选

最小模板结构：

```vue
<nav class="top-tabs" role="tablist" aria-label="竞赛工作台阶段切换">
  <button v-for="stage in stages" :key="stage.key" role="tab">
    {{ stage.label }}
  </button>
</nav>
```

- [ ] **Step 3: 在 `ContestEdit.vue` 接入新工作台骨架**

把原有 `editPanels` 升级为工作台阶段：

- 保留 `AdminContestFormPanel`
- 把 `ContestChallengeOrchestrationPanel` 作为 `pool`
- 先为 `awd-config / preflight / operations` 留阶段挂载点
- AWD 未开赛时默认停在编排阶段，已开赛后默认切到 `operations`

- [ ] **Step 4: 增加统一摘要条**

在 `ContestWorkbenchSummaryStrip.vue` 至少展示：

- 赛事模式
- 当前状态
- 已关联题目数
- AWD 准备度摘要（仅 AWD）

- [ ] **Step 5: 运行测试确认通过**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/spec-contest-orchestration-workbench/code/frontend
npx vitest run src/views/admin/__tests__/ContestEdit.test.ts src/components/admin/__tests__/ContestWorkbenchStageRail.test.ts src/composables/__tests__/useContestWorkbench.test.ts
```

预期：PASS。

- [ ] **Step 6: 提交工作台骨架实现**

```bash
git add code/frontend/src/views/admin/ContestEdit.vue code/frontend/src/components/admin/contest/ContestWorkbenchStageRail.vue code/frontend/src/components/admin/contest/ContestWorkbenchSummaryStrip.vue code/frontend/src/composables/useContestWorkbench.ts code/frontend/src/views/admin/__tests__/ContestEdit.test.ts code/frontend/src/components/admin/__tests__/ContestWorkbenchStageRail.test.ts code/frontend/src/composables/__tests__/useContestWorkbench.test.ts
git commit -m "feat(竞赛): 搭建竞赛编排工作台骨架"
```

## Task 3: 把轻量挂题面板升级为统一题目池

**Files:**
- Modify: `code/frontend/src/components/admin/contest/ContestChallengeOrchestrationPanel.vue`
- Modify: `code/frontend/src/components/admin/contest/ContestChallengeEditorDialog.vue`
- Create: `code/frontend/src/composables/useContestChallengePool.ts`
- Create: `code/frontend/src/components/admin/__tests__/ContestChallengeOrchestrationPanel.test.ts`
- Modify: `code/frontend/src/views/admin/__tests__/ContestEdit.test.ts`

- [ ] **Step 1: 先写题目池测试**

在 `ContestChallengeOrchestrationPanel.test.ts` 增加至少 3 个用例：

- `应该显示基础编排字段`
- `应该在 AWD 模式下显示 checker / SLA / 防守分 / 验证状态摘要列`
- `应该支持按未配置 AWD 和预检失败筛选`

最小断言示例：

```ts
expect(wrapper.text()).toContain('Checker')
expect(wrapper.text()).toContain('SLA')
expect(wrapper.text()).toContain('待重新验证')
```

- [ ] **Step 2: 实现 `useContestChallengePool`**

收敛：

- 排序后的 `challengeLinks`
- 题目池 summaryItems
- AWD 筛选视图
- 按状态过滤后的列表

最小接口：

```ts
export function useContestChallengePool(challengeLinks: Ref<AdminContestChallengeData[]>, contestMode: Ref<'jeopardy' | 'awd' | null>) {
  // return { visibleItems, summaryItems, activeFilter, setFilter }
}
```

- [ ] **Step 3: 在题目池面板接入 AWD 摘要列**

仅 AWD 赛事下展示：

- `Checker`
- `验证状态`
- `SLA / 防守分`
- `最近试跑`

但不要在这里展开完整表单。

- [ ] **Step 4: 保持基础挂题弹层只做基础编排**

`ContestChallengeEditorDialog.vue` 继续只负责：

- 关联题目
- 顺序
- 分值
- 可见性

把说明文案改成明确的“AWD 深度配置在下一阶段完成”。

- [ ] **Step 5: 跑测试确认通过**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/spec-contest-orchestration-workbench/code/frontend
npx vitest run src/components/admin/__tests__/ContestChallengeOrchestrationPanel.test.ts src/views/admin/__tests__/ContestEdit.test.ts
```

预期：PASS。

- [ ] **Step 6: 提交统一题目池改动**

```bash
git add code/frontend/src/components/admin/contest/ContestChallengeOrchestrationPanel.vue code/frontend/src/components/admin/contest/ContestChallengeEditorDialog.vue code/frontend/src/composables/useContestChallengePool.ts code/frontend/src/components/admin/__tests__/ContestChallengeOrchestrationPanel.test.ts code/frontend/src/views/admin/__tests__/ContestEdit.test.ts
git commit -m "feat(竞赛): 将挂题面板升级为统一题目池"
```

## Task 4: 把 AWD 配置与赛前检查接入工作台

**Files:**
- Modify: `code/frontend/src/components/admin/contest/AWDChallengeConfigPanel.vue`
- Modify: `code/frontend/src/components/admin/contest/AWDReadinessSummary.vue`
- Create: `code/frontend/src/components/admin/contest/ContestAwdPreflightPanel.vue`
- Modify: `code/frontend/src/views/admin/ContestEdit.vue`
- Modify: `code/frontend/src/components/admin/__tests__/AWDReadinessSummary.test.ts`
- Modify: `code/frontend/src/views/admin/__tests__/ContestEdit.test.ts`

- [ ] **Step 1: 先补赛前检查测试**

在 `AWDReadinessSummary.test.ts` 或新加的 `ContestAwdPreflightPanel` 测试里覆盖：

- `应该显示可开赛 / 可强制开赛 / 不可开赛结论`
- `应该列出阻塞项并显示快捷入口文案`
- `应该保留强制开赛弹层入口`

最小断言示例：

```ts
expect(wrapper.text()).toContain('不可开赛')
expect(wrapper.text()).toContain('Challenge 101')
expect(wrapper.text()).toContain('返回 AWD 配置')
```

- [ ] **Step 2: 实现 `ContestAwdPreflightPanel.vue`**

职责：

- 组合 `AWDReadinessSummary`
- 渲染阻塞项列表
- 暴露跳转到 `awd-config` 阶段的事件

最小接口：

```vue
<ContestAwdPreflightPanel
  :readiness="readiness"
  @navigate:challenge="handleNavigateChallenge"
  @navigate:stage="selectStage('awd-config')"
/>
```

- [ ] **Step 3: 调整 `AWDChallengeConfigPanel.vue` 成为工作台阶段**

至少补齐：

- 连续逐题编辑的上下文文案
- 从题目池/赛前检查回跳时的当前题高亮
- `上一题 / 下一题` 的事件接口占位

- [ ] **Step 4: 在 `ContestEdit.vue` 串起 `awd-config -> preflight`**

要求：

- AWD 模式下显示两个新阶段
- 赛前检查可以把管理员带回对应配置阶段
- 仍复用现有 `AWDReadinessOverrideDialog`

- [ ] **Step 5: 跑测试确认通过**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/spec-contest-orchestration-workbench/code/frontend
npx vitest run src/components/admin/__tests__/AWDReadinessSummary.test.ts src/views/admin/__tests__/ContestEdit.test.ts
```

预期：PASS。

- [ ] **Step 6: 提交 AWD 配置与赛前检查集成**

```bash
git add code/frontend/src/components/admin/contest/AWDChallengeConfigPanel.vue code/frontend/src/components/admin/contest/AWDReadinessSummary.vue code/frontend/src/components/admin/contest/ContestAwdPreflightPanel.vue code/frontend/src/views/admin/ContestEdit.vue code/frontend/src/components/admin/__tests__/AWDReadinessSummary.test.ts code/frontend/src/views/admin/__tests__/ContestEdit.test.ts
git commit -m "feat(竞赛): 将 AWD 配置与赛前检查接入工作台"
```

## Task 5: 把运行态承接到工作台最后一段

**Files:**
- Modify: `code/frontend/src/components/admin/contest/AWDOperationsPanel.vue`
- Modify: `code/frontend/src/views/admin/ContestEdit.vue`
- Modify: `code/frontend/src/components/admin/__tests__/AWDOperationsPanel.test.ts`
- Modify: `code/frontend/src/views/admin/__tests__/ContestManage.test.ts`
- Modify: `code/frontend/src/views/admin/__tests__/ContestEdit.test.ts`

- [ ] **Step 1: 先补运行态承接测试**

在 `AWDOperationsPanel.test.ts` 和 `ContestEdit.test.ts` 至少补 3 个用例：

- `未开赛时运行段应显示尚未进入运行阶段`
- `已开赛时工作台默认切到运行段`
- `ContestManage` 仍然可以跳到编辑页，而不是维持平行 AWD 运维入口`

最小断言示例：

```ts
expect(wrapper.text()).toContain('尚未进入运行阶段')
expect(wrapper.text()).toContain('轮次态势')
```

- [ ] **Step 2: 调整 `AWDOperationsPanel.vue` 的未开赛降级展示**

要求：

- 赛事未到 `running/frozen` 时，保留面板骨架
- 禁用会改动赛时数据的按钮
- 明确提示“需先通过赛前检查并开赛”

- [ ] **Step 3: 在 `ContestEdit.vue` 接入运行段默认聚焦**

当满足：

- `contest.mode === 'awd'`
- `contest.status === 'running' || contest.status === 'frozen'`

则默认活动阶段为 `operations`。

- [ ] **Step 4: 清理 `ContestManage` 中平行 AWD 运维的文案依赖**

最小改动即可：

- 保留赛事管理台现有结构
- 但让“进入 AWD 运维”类文案更明确地指向“进入竞赛工作台”

- [ ] **Step 5: 运行相关测试确认通过**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/spec-contest-orchestration-workbench/code/frontend
npx vitest run src/components/admin/__tests__/AWDOperationsPanel.test.ts src/views/admin/__tests__/ContestManage.test.ts src/views/admin/__tests__/ContestEdit.test.ts
```

预期：PASS。

- [ ] **Step 6: 提交运行态承接改动**

```bash
git add code/frontend/src/components/admin/contest/AWDOperationsPanel.vue code/frontend/src/views/admin/ContestEdit.vue code/frontend/src/components/admin/__tests__/AWDOperationsPanel.test.ts code/frontend/src/views/admin/__tests__/ContestManage.test.ts code/frontend/src/views/admin/__tests__/ContestEdit.test.ts
git commit -m "feat(竞赛): 将 AWD 运行态承接到工作台"
```

## Task 6: 做最终验证并补文档回链

**Files:**
- Modify: `docs/architecture/features/2026-04-14-contest-orchestration-workbench-design.md`
- Modify: `docs/superpowers/plans/2026-04-15-contest-orchestration-workbench-implementation.md`

- [ ] **Step 1: 运行最小充分验证**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/spec-contest-orchestration-workbench/code/frontend
npx vitest run src/views/admin/__tests__/ContestEdit.test.ts src/views/admin/__tests__/ContestManage.test.ts src/components/admin/__tests__/ContestChallengeOrchestrationPanel.test.ts src/components/admin/__tests__/AWDReadinessSummary.test.ts src/components/admin/__tests__/AWDOperationsPanel.test.ts src/components/admin/__tests__/ContestWorkbenchStageRail.test.ts src/composables/__tests__/useContestWorkbench.test.ts
npx vue-tsc --noEmit
```

预期：

- 所有相关 Vitest 用例通过
- `vue-tsc --noEmit` 通过

- [ ] **Step 2: 回写 spec / plan 的实施状态**

在 spec 或 plan 里补最小必要说明：

- 哪些阶段本轮已落地
- 哪些交互保留到后续 phase

- [ ] **Step 3: 检查 git 差异仅包含本轮工作台改动**

运行：

```bash
git status --short
git diff --stat
```

预期：只包含竞赛工作台相关文件和对应测试。

- [ ] **Step 4: 提交最终验证与文档回链**

```bash
git add docs/architecture/features/2026-04-14-contest-orchestration-workbench-design.md docs/superpowers/plans/2026-04-15-contest-orchestration-workbench-implementation.md
git commit -m "docs(竞赛): 更新编排工作台实施记录"
```

## Implementation Backlink（2026-04-15）

- 最小充分验证已执行：
  - `npx vitest run src/views/admin/__tests__/ContestEdit.test.ts src/views/admin/__tests__/ContestManage.test.ts src/components/admin/__tests__/ContestChallengeOrchestrationPanel.test.ts src/components/admin/__tests__/AWDReadinessSummary.test.ts src/components/admin/__tests__/AWDOperationsPanel.test.ts src/components/admin/__tests__/ContestWorkbenchStageRail.test.ts src/composables/__tests__/useContestWorkbench.test.ts`
    - 结果：`7` 个测试文件、`45` 个用例全部通过
  - `npx vue-tsc --noEmit`
    - 结果：通过，退出码 `0`
- 本轮已落地阶段：`基础信息`、`题目池`、`AWD 配置`、`赛前检查`、`轮次运行`
- 明确保留到后续 phase 的交互：
  - 题目池批量操作与复杂拖拽编排
  - `mode=jeopardy` 下的 `赛前检查` 扩展
  - 已开赛后对前置阶段高风险字段的更强限制或风险提示
