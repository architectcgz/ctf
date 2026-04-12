# Student Dashboard Action Panels Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the student dashboard `recommendation` / `category` / `difficulty` panels into an action-first workspace and wire category/difficulty actions into query-synced challenge filters.

**Architecture:** Keep `DashboardView` as the route-level tab shell and rework the three embedded panel components around one shared structure: action heading, three-card `metric-panel` summary strip, flat action list, and compact rationale block. Reuse existing `journal-soft-surface` / `metric-panel` tokens, add only minimal ranking helpers in the student dashboard utility layer, and make `ChallengeList` read/write `category` and `difficulty` route queries so dashboard action buttons land on concrete filtered challenge results.

**Tech Stack:** Vue 3 SFCs, Vue Router 4, TypeScript, Vitest + Vue Test Utils, shared CTF workspace CSS tokens.

---

## File Map

- Modify: `code/frontend/src/views/challenges/ChallengeList.vue`
  Route-query hydration and sync for `category` / `difficulty` filters.
- Modify: `code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts`
  Query-sync regression coverage for challenge filters.
- Modify: `code/frontend/src/views/dashboard/DashboardView.vue`
  Add filtered challenge navigation handlers and pass new emits into category/difficulty panels.
- Modify: `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
  User-facing assertions for the three action-first panels and route-action behavior.
- Modify: `code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue`
  Refactor recommendation panel into action-first workspace shell.
- Modify: `code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue`
  Replace highlight-card layout with action-ranked category list and category-specific CTA.
- Modify: `code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue`
  Replace explanation-heavy layout with action-ranked difficulty list and difficulty-specific CTA.
- Modify: `code/frontend/src/components/dashboard/student/utils.ts`
  Shared ranking helpers for category and difficulty action ordering.
- Modify: `code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts`
  Raw-source assertions for the shared summary-card stack and selector boundaries after the refactor.

## Task 1: Sync Challenge Filters With Route Query

**Files:**
- Modify: `code/frontend/src/views/challenges/ChallengeList.vue`
- Test: `code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts`

- [ ] **Step 1: Add failing tests for query hydration and query updates**

```ts
it('should hydrate category and difficulty filters from route query', async () => {
  const router = createTestRouter()
  await router.push('/challenges?category=crypto&difficulty=medium')
  await router.isReady()

  const wrapper = mount(ChallengeList, { global: { plugins: [router] } })
  await flushPromises()

  expect((wrapper.get('#challenge-category-filter').element as HTMLSelectElement).value).toBe('crypto')
  expect((wrapper.get('#challenge-difficulty-filter').element as HTMLSelectElement).value).toBe('medium')
})

it('should write filter changes back to the route query', async () => {
  const router = createTestRouter()
  await router.push('/challenges')
  await router.isReady()

  const wrapper = mount(ChallengeList, { global: { plugins: [router] } })
  await flushPromises()

  await wrapper.get('#challenge-category-filter').setValue('crypto')
  await flushPromises()

  expect(router.currentRoute.value.query.category).toBe('crypto')
})
```

- [ ] **Step 2: Run the focused test file and confirm the new assertions fail**

Run:

```bash
npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts
```

Expected:

- FAIL because `ChallengeList.vue` currently does not use `useRoute` and does not sync filter state into the URL query.

- [ ] **Step 3: Implement minimal route-query sync in `ChallengeList.vue`**

Use `useRoute` + `useRouter`, hydrate `categoryFilter` and `difficultyFilter` from `route.query`, and keep URL updates scoped to these two filters only.

```ts
const route = useRoute()

function applyRouteFilters(): void {
  categoryFilter.value = isChallengeCategory(route.query.category) ? route.query.category : ''
  difficultyFilter.value = isChallengeDifficulty(route.query.difficulty) ? route.query.difficulty : ''
}

async function syncFilterQuery(): Promise<void> {
  const nextQuery = {
    ...route.query,
    category: categoryFilter.value || undefined,
    difficulty: difficultyFilter.value || undefined,
  }
  await router.replace({ query: nextQuery })
}
```

Implementation notes:

- Keep `keyword` behavior unchanged in this task.
- Guard against invalid query values instead of blindly trusting `route.query`.
- Avoid creating refresh loops: apply route state first, then trigger `refresh()` only when the effective filter set changes.

- [ ] **Step 4: Re-run the challenge-list test file**

Run:

```bash
npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts
```

Expected:

- PASS, including the new query hydration/sync assertions.

- [ ] **Step 5: Commit the isolated query-sync change**

```bash
git add code/frontend/src/views/challenges/ChallengeList.vue code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts
git commit -m "feat(challenges): 同步分类和难度筛选到路由"
```

## Task 2: Refactor Recommendation Into An Action-First Panel

**Files:**
- Modify: `code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue`
- Test: `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- Test: `code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts`

- [ ] **Step 1: Add failing dashboard and raw-source assertions for the new recommendation structure**

Update `DashboardView.test.ts` to assert the new action-first copy and main action list:

```ts
expect(wrapper.text()).toContain('现在先练这几道')
expect(wrapper.text()).toContain('当前目标难度')
expect(wrapper.text()).toContain('浏览全部题目')
```

Update `studentUserSurfaceAlignment.test.ts` to assert the recommendation panel adopts the shared summary-card stack:

```ts
expect(studentRecommendationSource).toContain('metric-panel-grid')
expect(studentRecommendationSource).toContain('metric-panel-card')
expect(studentRecommendationSource).toContain('metric-panel-helper')
```

- [ ] **Step 2: Run the two focused test files and confirm failure**

Run:

```bash
npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts
```

Expected:

- FAIL because the recommendation panel still renders `补短板计划 / 推荐摘要` and does not yet use the new summary-card stack.

- [ ] **Step 3: Rebuild `StudentRecommendationPage.vue` around the action-first shell**

Target structure:

```vue
<section class="journal-soft-surface ...">
  <header>
    <div class="journal-eyebrow">Training Queue</div>
    <h1 class="workspace-tab-heading__title">现在先练这几道</h1>
  </header>

  <div class="metric-panel-grid progress-strip metric-panel-default-surface">
    <article class="progress-card metric-panel-card">...</article>
  </div>

  <section class="recommend-directory">...</section>
  <aside class="recommend-rationale">...</aside>
</section>
```

Implementation notes:

- Keep the existing `openChallenge`, `openChallenges`, and `openSkillProfile` emits.
- Flatten the page into one main recommendation directory and one compact rationale block.
- Replace the current local note-summary ownership with the shared `metric-panel` stack.
- Preserve the existing empty state, but keep only one primary CTA: `浏览全部题目`.

- [ ] **Step 4: Re-run the focused dashboard and alignment tests**

Run:

```bash
npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts
```

Expected:

- PASS for the recommendation-specific assertions.

- [ ] **Step 5: Commit the recommendation-panel refactor**

```bash
git add code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
git commit -m "feat(student-dashboard): 重做训练建议行动页"
```

## Task 3: Turn Category Progress Into A Ranked Action List

**Files:**
- Modify: `code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue`
- Modify: `code/frontend/src/components/dashboard/student/utils.ts`
- Modify: `code/frontend/src/views/dashboard/DashboardView.vue`
- Test: `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- Test: `code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts`

- [ ] **Step 1: Add failing tests for category CTA behavior and new panel structure**

In `DashboardView.test.ts`, add a category-panel case that checks both the new copy and the filtered route push:

```ts
routeState.query = { panel: 'category' }
const wrapper = mountDashboard()
await flushPromises()

expect(wrapper.text()).toContain('优先补这个分类')
await wrapper.get('[data-test=\"category-action-crypto\"]').trigger('click')
expect(pushMock).toHaveBeenCalledWith({ name: 'Challenges', query: { category: 'crypto' } })
```

In `studentUserSurfaceAlignment.test.ts`, assert the category panel now uses the shared metric-card stack instead of the highlight-card pair.

- [ ] **Step 2: Run the dashboard and alignment tests to verify failure**

Run:

```bash
npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts
```

Expected:

- FAIL because the category panel still renders `分类覆盖概况` and has no category-specific CTA emit or `data-test` selector.

- [ ] **Step 3: Implement category ranking helpers and action routing**

Add ranking helpers to `utils.ts`:

```ts
export function rankCategoryStatsForAction(stats: CategoryStat[]) {
  return [...stats]
    .map((item) => ({ ...item, rate: progressRate(item.total, item.solved) }))
    .sort((left, right) => left.rate - right.rate || right.total - left.total)
}
```

Update `StudentCategoryProgressPage.vue`:

- Replace the highlight-card section with one shared three-card summary strip.
- Render a flat action list where each row exposes a deterministic test hook:

```vue
<button
  :data-test="`category-action-${item.category}`"
  class="journal-btn-primary"
  @click="emit('openCategoryChallenges', item.category)"
>
  去练这个分类
</button>
```

Update `DashboardView.vue` to pass the new emit and push filtered challenge routes:

```ts
function openCategoryChallenges(category: string): void {
  void router.push({ name: 'Challenges', query: { category } })
}
```

- [ ] **Step 4: Re-run the dashboard and alignment tests**

Run:

```bash
npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts
```

Expected:

- PASS, including the new category-panel copy and route-action assertions.

- [ ] **Step 5: Commit the category action-list refactor**

```bash
git add code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue code/frontend/src/components/dashboard/student/utils.ts code/frontend/src/views/dashboard/DashboardView.vue code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
git commit -m "feat(student-dashboard): 将分类页改为行动列表"
```

## Task 4: Turn Difficulty Progress Into A Difficulty-Push Workspace

**Files:**
- Modify: `code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue`
- Modify: `code/frontend/src/components/dashboard/student/utils.ts`
- Modify: `code/frontend/src/views/dashboard/DashboardView.vue`
- Test: `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- Test: `code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts`

- [ ] **Step 1: Add failing tests for difficulty CTA behavior and new panel copy**

In `DashboardView.test.ts`, add an explicit difficulty-panel case:

```ts
routeState.query = { panel: 'difficulty' }
const wrapper = mountDashboard()
await flushPromises()

expect(wrapper.text()).toContain('下一步把训练强度推到这里')
await wrapper.get('[data-test=\"difficulty-action-medium\"]').trigger('click')
expect(pushMock).toHaveBeenCalledWith({ name: 'Challenges', query: { difficulty: 'medium' } })
```

Extend `studentUserSurfaceAlignment.test.ts` to assert the difficulty panel uses the shared summary strip and no longer depends on the old `difficulty-insight-list` main structure.

- [ ] **Step 2: Run the dashboard and alignment tests and confirm failure**

Run:

```bash
npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts
```

Expected:

- FAIL because the difficulty panel still renders `难度层级总览 / 训练解读` and exposes no difficulty-specific CTA.

- [ ] **Step 3: Implement difficulty ranking and action routing**

Extend `utils.ts` with a difficulty-order helper that still respects the existing canonical order:

```ts
export function rankDifficultyStatsForAction(stats: DifficultyStat[]) {
  return difficultyOrder
    .map((difficulty) => stats.find((item) => item.difficulty === difficulty))
    .filter(Boolean)
    .map((item) => ({ ...item, rate: progressRate(item.total, item.solved) }))
}
```

Update `StudentDifficultyPage.vue`:

- Replace the old top summary note with a shared three-card metric strip.
- Collapse the current explanation-heavy split layout into one action list plus one compact guidance block.
- Highlight the current focus row and add deterministic CTA selectors:

```vue
<button
  :data-test="`difficulty-action-${item.difficulty}`"
  class="journal-btn-primary"
  @click="emit('openDifficultyChallenges', item.difficulty)"
>
  去练这档题
</button>
```

Update `DashboardView.vue`:

```ts
function openDifficultyChallenges(difficulty: string): void {
  void router.push({ name: 'Challenges', query: { difficulty } })
}
```

- [ ] **Step 4: Re-run the dashboard and alignment tests**

Run:

```bash
npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts
```

Expected:

- PASS for the difficulty-panel assertions and shared-surface regression checks.

- [ ] **Step 5: Commit the difficulty action-panel refactor**

```bash
git add code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue code/frontend/src/components/dashboard/student/utils.ts code/frontend/src/views/dashboard/DashboardView.vue code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
git commit -m "feat(student-dashboard): 将难度页改为强度推进工作区"
```

## Task 5: Run Final Focused Verification

**Files:**
- Verify: `code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts`
- Verify: `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- Verify: `code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts`
- Verify: `code/frontend/src/views/challenges/ChallengeList.vue`
- Verify: `code/frontend/src/views/dashboard/DashboardView.vue`

- [ ] **Step 1: Run the three focused regression suites together**

```bash
npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts
```

Expected:

- PASS for all targeted dashboard/challenges regressions.

- [ ] **Step 2: Run frontend typecheck**

```bash
npm run typecheck
```

Expected:

- PASS with no Vue/TypeScript errors introduced by the new emits or route-query handling.

- [ ] **Step 3: Inspect the final diff for scope discipline**

```bash
git diff --stat HEAD~4..HEAD
git diff -- code/frontend/src/components/dashboard/student code/frontend/src/views/dashboard code/frontend/src/views/challenges
```

Expected:

- Only the planned student-dashboard and challenge-filter files are touched.
- No unrelated visual cleanup leaks into the branch.

- [ ] **Step 4: Create the final implementation commit if verification required small cleanup**

```bash
git add code/frontend/src/components/dashboard/student code/frontend/src/views/dashboard code/frontend/src/views/challenges
git commit -m "test(student-dashboard): 收口行动型子页回归"
```

Expected:

- Skip this commit if verification produced no code changes.

- [ ] **Step 5: Prepare handoff notes**

Capture:

- Final commands executed
- Exact tests passed
- Whether the optional cleanup commit was needed
- Any residual UX risk, especially if `keyword` query sync was intentionally left out

## Review Notes

- `docs/superpowers/*` is ignored in this repository. When updating this plan or the spec later, remember to use `git add -f`.
- Subagent review loops from the original skill are intentionally not part of this plan because delegation requires explicit user authorization in this environment.
