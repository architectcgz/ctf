> 状态：Current
> 事实源：skill profile page model、skill profile workspace shell、前端架构 allowlist
> 替代：无

# Skill Profile Route Target Cleanup Plan

## 目标

- 把 `useSkillProfilePage.ts` 从“去做题 / 去题目详情”的 `router.push()` 收口成纯 route target contract。
- 保持六维画像的数据加载、错误态、教师学员选择和刷新 owner 不变，同时再清掉 1 条 `featureRouterImportAllowlist`。

## 非目标

- 不处理 `SkillProfile.vue` 的 tab query owner。
- 不迁移 `SkillProfileWorkspaceShell.vue` 的目录归属。
- 不改六维画像数据加载、推荐算法或教师视角学员选择逻辑。

## 输入依据

- `code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts`
- `code/frontend/src/views/profile/SkillProfile.vue`
- `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
- `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components/navigation/AppRouteLink.vue`

## 当前结论

- `useSkillProfilePage.ts` 的 router 依赖只剩“去做题 / 去题目详情”两类声明式导航。
- 学员选择、画像与推荐数据加载、刷新和错误态 owner 已经都在合适位置，不需要再移动。
- 这一条适合直接收口成 route target contract。

## 设计边界

### `skillProfileRoutes.ts` 本轮负责

- 生成题库页 route target
- 生成题目详情页 route target

### `useSkillProfilePage()` 本轮负责

- 教师/学生六维画像数据加载、推荐数据加载、刷新和错误态 owner
- 暴露 `challengesRoute` 与 `buildChallengeRoute()`，不再直接导航

### `SkillProfileWorkspaceShell.vue` 本轮负责

- 通过共享 `AppRouteLink` 消费 route target
- 保持其它按钮、tab、教师学员选择和刷新交互不变

## 任务切片

### Slice 1：page model 去 router 化

- 目标：
  - 新增 skill profile route helper
  - `useSkillProfilePage.ts` 去掉 `vue-router`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - allowlist 是否可净减少 1 条

### Slice 2：workspace shell 切到 `AppRouteLink`

- 目标：
  - `SkillProfileWorkspaceShell.vue` 的“去做题”和推荐题目跳转改成共享 `AppRouteLink`
  - `SkillProfile.vue` 继续只做组合
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/profile/__tests__/SkillProfile.test.ts`
- Review focus：
  - route target contract 是否清楚，没有把刷新 / 学员选择 owner 一起迁走

### Slice 3：allowlist / review / backlog 收尾

- 目标：
  - 更新 allowlist、review 和 backlog
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `featureRouterImportAllowlist` 是否净减少 1 条

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/profile/__tests__/SkillProfile.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/skill-profile-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-skill-profile-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-skill-profile-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/skill-profile/model/index.ts code/frontend/src/features/skill-profile/model/skillProfileRoutes.ts code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts code/frontend/src/views/profile/SkillProfile.vue code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `SkillProfile.vue` 仍保留 `useUrlSyncedTabs()` 的 window query owner，这轮不一并处理。
