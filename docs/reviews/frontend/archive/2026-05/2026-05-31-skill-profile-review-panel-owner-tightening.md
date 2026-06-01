# Skill Profile Panel Owner Tightening Review

## Review target

- Repository: `/home/azhi/workspace/projects/ctf`
- Branch: `main`
- Diff source: working tree uncommitted change set `skill-profile panel owner tightening`
- Files reviewed:
  - `.harness/reuse-decisions/skill-profile-panel-owner-tightening.md`
  - `docs/plan/impl-plan/2026-05-31-skill-profile-panel-owner-tightening-plan.md`
  - `code/frontend/src/features/skill-profile/model/skillProfilePanelRoute.ts`
  - `code/frontend/src/features/skill-profile/model/index.ts`
  - `code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts`
  - `code/frontend/src/pages/profile/SkillProfileRoutePage.vue`
  - `code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Validation actually run:
  - `bash scripts/check-task-intake.sh`
  - `git diff --check -- .harness/reuse-decisions/skill-profile-panel-owner-tightening.md docs/plan/impl-plan/2026-05-31-skill-profile-panel-owner-tightening-plan.md code/frontend/src/features/skill-profile/model/skillProfilePanelRoute.ts code/frontend/src/features/skill-profile/model/index.ts code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts code/frontend/src/pages/profile/SkillProfileRoutePage.vue code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `cd code/frontend && npm run test:run -- src/pages/profile/__tests__/SkillProfile.test.ts`

## Classification check

- 结论：同意按 `HARNESS / frontend non-trivial refactor review` 审查，不需要升级成更大范围结构改造。
- 依据：本次改动只触达 skill profile 页面自己的 panel query owner、route page 组合层和对应护栏测试，没有扩散到数据加载 owner、推荐 workflow 或共享导航基础设施。

## Gate verdict

- `pass with minor issues`
- blocker：无

## Findings

1. `Non-blocking` `code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts:248-261`
   当前测试没有直接证明 `?panel=recommendations` 的初始恢复真的生效。这里先挂载带 query 的页面，再直接在推荐列表里找链接并触发点击；但 `SkillProfileWorkspaceShell.vue:253-290` 的推荐面板使用的是 `v-show`，即使 `activePanel` 恢复回归，推荐链接仍然会留在 DOM 里，测试依旧可能通过。建议在 `mountPage('/profile?panel=recommendations')` 之后，直接断言当前 tab 的 `aria-selected`、推荐面板的 `aria-hidden/class active`，以及分析面板已经隐藏，再继续验证推荐题目跳转。

## Material findings

- 无 material findings。

## Senior implementation assessment

- 这次 owner 收口方向是对的，且改动面控制得比较干净。`useSkillProfilePage.ts:27-45,138-143` 现在统一持有 query 读取、panel 解析和 query 回写，route/query owner 已经单点收口到 page model。
- `skillProfilePanelRoute.ts:1-25` 保持纯 helper，只负责 panel 归一和 query 构建，没有把 route transport 或 UI 细节混进去，这和教师概览、用户治理、赛事目录最近几轮模式一致。
- `SkillProfileRoutePage.vue:1-23,53-83` 已退回组合层：它只组装 page model 输出、tab 清单和 `useTabKeyboardNavigation()`。route page 不再直接 import `useUrlSyncedTabs()`，也不再自己维护 route-aware panel state。
- UI keyboard owner 目前仍然清晰，保留在 route page 这一层是合理的：键盘焦点与 tab 顺序是纯 UI concerns，`useTabKeyboardNavigation()` 只消费 `orderedTabs + selectTab`，没有重新把 route/query owner 拉回页面壳。

## Required re-validation

- 补强初始恢复断言后，重新执行：
  - `cd code/frontend && npm run test:run -- src/pages/profile/__tests__/SkillProfile.test.ts`
- 建议新增的断言路径：
  - 以 `/profile?panel=recommendations` 挂载页面。
  - 断言 `#skill-profile-tab-recommendations` 为 `aria-selected="true"`。
  - 断言 `#skill-profile-panel-recommendations` 为可见态，`#skill-profile-panel-analysis` 为隐藏态。

## Residual risk

- 当前代码路径本身没有发现 owner 混写或 keyboard 归属漂移，剩余风险主要在回归护栏强度：如果后续有人改坏初始 query hydrate，现有测试更像“推荐链接仍可跳转”而不是“页面初始 panel 恢复正确”。
- 本轮只跑了 skill profile 定向测试，没有额外跑全量前端测试；对这个切片来说证据已基本充分，但跨页面共享导航 helper 的全局回归仍依赖仓库既有测试矩阵。

## Touched known-debt status

- 已触达并实质收口 `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md:345` 记录的 `skill profile panel owner tightening` 债务。
- 当前 touched surface 上，route/query owner 已从 `SkillProfileRoutePage.vue` 收回 `useSkillProfilePage.ts`，`useUrlSyncedTabs()` 也已退出 route page。剩下的问题是测试护栏偏弱，不属于这条 owner debt 未收口。
