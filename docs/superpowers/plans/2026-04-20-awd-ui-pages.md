# AWD Frontend Pages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 用新的学生端、管理员端、教师端 AWD workspace 替换旧 AWD 页面，落地文档中的 19 个前端页面，并移除旧 AWD 主入口。

**Architecture:** 在 `src/modules/awd` 下新增独立模块，按 `layouts + adapters + views + shared components` 分层。路由层把 AWD 赛事导向新 workspace，现有 `useContestAWDWorkspace`、`usePlatformContestAwd`、`useTeacherAwdReviewDetail` 继续作为真实数据来源，缺口字段由 adapter 层集中补齐。

**Tech Stack:** Vue 3、TypeScript、Vue Router 4、Vitest、现有 workspace shell 样式、现有 AWD composables / API contracts。

---

## File Structure

### New module files

- Create: `code/frontend/src/modules/awd/navigation.ts`
  AWD 三族页面定义、左侧导航配置、路由辅助常量。
- Create: `code/frontend/src/modules/awd/types.ts`
  共享 view model 与页面 key 类型。
- Create: `code/frontend/src/modules/awd/adapters/studentAwdPageAdapter.ts`
  学生端 5 页数据适配。
- Create: `code/frontend/src/modules/awd/adapters/adminAwdPageAdapter.ts`
  管理员端 9 页数据适配。
- Create: `code/frontend/src/modules/awd/adapters/teacherAwdPageAdapter.ts`
  教师端 5 页数据适配。
- Create: `code/frontend/src/modules/awd/components/AwdContextHero.vue`
  顶部赛事上下文组件。
- Create: `code/frontend/src/modules/awd/components/AwdPageNav.vue`
  页内导航组件。
- Create: `code/frontend/src/modules/awd/components/AwdMetricStrip.vue`
  顶部指标条。
- Create: `code/frontend/src/modules/awd/components/AwdEventTimeline.vue`
  事件流组件。
- Create: `code/frontend/src/modules/awd/components/AwdStatusShell.vue`
  AWD 页面统一 loading / error / empty 壳。
- Create: `code/frontend/src/modules/awd/layouts/StudentAwdWorkspaceLayout.vue`
- Create: `code/frontend/src/modules/awd/layouts/AdminAwdWorkspaceLayout.vue`
- Create: `code/frontend/src/modules/awd/layouts/TeacherAwdWorkspaceLayout.vue`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdOverviewView.vue`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdServicesView.vue`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdTargetsView.vue`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdAttacksView.vue`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdCollabView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdOverviewView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdReadinessView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdRoundsView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdServicesView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdAttacksView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdTrafficView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdAlertsView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdInstancesView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdReplayView.vue`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdOverviewView.vue`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdTeamsView.vue`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdServicesView.vue`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdReplayView.vue`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdExportView.vue`

### New tests

- Create: `code/frontend/src/modules/awd/__tests__/navigation.test.ts`
- Create: `code/frontend/src/modules/awd/__tests__/studentAwdPageAdapter.test.ts`
- Create: `code/frontend/src/modules/awd/__tests__/adminAwdPageAdapter.test.ts`
- Create: `code/frontend/src/modules/awd/__tests__/teacherAwdPageAdapter.test.ts`
- Create: `code/frontend/src/modules/awd/__tests__/StudentAwdWorkspaceLayout.test.ts`
- Create: `code/frontend/src/modules/awd/__tests__/AdminAwdWorkspaceLayout.test.ts`
- Create: `code/frontend/src/modules/awd/__tests__/TeacherAwdWorkspaceLayout.test.ts`

### Existing files to modify

- Modify: `code/frontend/src/router/index.ts`
  新增 19 个 AWD 页面路由，接管旧 AWD 入口与重定向。
- Modify: `code/frontend/src/config/backofficeNavigation.ts`
  教师 / 管理员侧导航入口切到新的 AWD workspace。
- Modify: `code/frontend/src/views/contests/ContestDetail.vue`
  移除旧 AWD 内嵌战场面板，AWD 赛事跳新入口。
- Modify: `code/frontend/src/views/platform/ContestEdit.vue`
  剥离 AWD 运行工作台职责。
- Modify: `code/frontend/src/views/platform/ContestOperationsHub.vue`
  如仍保留，改为跳转壳或从导航中退出主链路。
- Modify: `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
  拆薄为教师 AWD workspace 路由壳或重定向。
- Modify: `code/frontend/src/views/platform/AWDReviewIndex.vue`
  管理员 AWD 复盘入口改接新管理员 AWD workspace。
- Modify: `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  教师 AWD 复盘入口改接新教师 AWD workspace。
- Modify: `code/frontend/src/components/layout/Sidebar.vue`
  若需要补图标 / active 状态逻辑，随导航一起更新。

### Existing files likely removable after migration

- Delete or stop importing: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- Delete or stop importing AWD-only legacy UI fragments from `ContestEdit.vue` path if no longer used as主入口

---

### Task 1: Scaffold AWD module and route metadata

**Files:**
- Create: `code/frontend/src/modules/awd/navigation.ts`
- Create: `code/frontend/src/modules/awd/types.ts`
- Test: `code/frontend/src/modules/awd/__tests__/navigation.test.ts`

- [ ] **Step 1: Write the failing navigation test**

```ts
import { describe, expect, it } from 'vitest'
import {
  ADMIN_AWD_PAGES,
  STUDENT_AWD_PAGES,
  TEACHER_AWD_PAGES,
  buildStudentAwdPath,
} from '@/modules/awd/navigation'

describe('awd navigation', () => {
  it('defines all 19 documented pages', () => {
    expect(STUDENT_AWD_PAGES).toHaveLength(5)
    expect(ADMIN_AWD_PAGES).toHaveLength(9)
    expect(TEACHER_AWD_PAGES).toHaveLength(5)
  })

  it('builds student awd route paths', () => {
    expect(buildStudentAwdPath('42', 'overview')).toBe('/contests/42/awd/overview')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/navigation.test.ts`
Expected: FAIL with module not found for `@/modules/awd/navigation`

- [ ] **Step 3: Write minimal navigation/types implementation**

```ts
export type StudentAwdPageKey = 'overview' | 'services' | 'targets' | 'attacks' | 'collab'
export interface AwdPageDefinition<T extends string> {
  key: T
  label: string
  description: string
}

export const STUDENT_AWD_PAGES: AwdPageDefinition<StudentAwdPageKey>[] = [
  { key: 'overview', label: '战场总览', description: '...' },
]

export function buildStudentAwdPath(contestId: string, page: StudentAwdPageKey): string {
  return `/contests/${encodeURIComponent(contestId)}/awd/${page}`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/navigation.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/frontend/src/modules/awd/navigation.ts \
  code/frontend/src/modules/awd/types.ts \
  code/frontend/src/modules/awd/__tests__/navigation.test.ts
git commit -m "feat(awd): 初始化页面导航定义"
```

### Task 2: Build shared AWD shell primitives

**Files:**
- Create: `code/frontend/src/modules/awd/components/AwdContextHero.vue`
- Create: `code/frontend/src/modules/awd/components/AwdPageNav.vue`
- Create: `code/frontend/src/modules/awd/components/AwdMetricStrip.vue`
- Create: `code/frontend/src/modules/awd/components/AwdEventTimeline.vue`
- Create: `code/frontend/src/modules/awd/components/AwdStatusShell.vue`
- Create: `code/frontend/src/modules/awd/layouts/StudentAwdWorkspaceLayout.vue`
- Create: `code/frontend/src/modules/awd/layouts/AdminAwdWorkspaceLayout.vue`
- Create: `code/frontend/src/modules/awd/layouts/TeacherAwdWorkspaceLayout.vue`
- Test: `code/frontend/src/modules/awd/__tests__/StudentAwdWorkspaceLayout.test.ts`
- Test: `code/frontend/src/modules/awd/__tests__/AdminAwdWorkspaceLayout.test.ts`
- Test: `code/frontend/src/modules/awd/__tests__/TeacherAwdWorkspaceLayout.test.ts`

- [ ] **Step 1: Write failing layout tests for shared shell anatomy**

```ts
it('renders context hero, left nav, and content slot', () => {
  const wrapper = mount(StudentAwdWorkspaceLayout, {
    props: {
      contestTitle: '2026 春季 AWD',
      pageTitle: '战场总览',
      pages: STUDENT_AWD_PAGES,
      currentPage: 'overview',
      heroMetrics: [{ label: '当前轮', value: 'R12' }],
    },
    slots: { default: '<div data-testid="awd-page-slot">body</div>' },
  })

  expect(wrapper.text()).toContain('2026 春季 AWD')
  expect(wrapper.find('[data-testid="awd-page-slot"]').exists()).toBe(true)
  expect(wrapper.findAll('[data-testid="awd-page-nav-item"]')).toHaveLength(5)
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/StudentAwdWorkspaceLayout.test.ts src/modules/awd/__tests__/AdminAwdWorkspaceLayout.test.ts src/modules/awd/__tests__/TeacherAwdWorkspaceLayout.test.ts`
Expected: FAIL with missing layout/component modules

- [ ] **Step 3: Implement shared shell primitives**

```vue
<template>
  <section class="awd-workspace-shell">
    <AwdContextHero :title="contestTitle" :page-title="pageTitle" :metrics="heroMetrics" />
    <div class="awd-workspace-shell__body">
      <AwdPageNav :items="pages" :current-page="currentPage" />
      <AwdStatusShell :loading="loading" :error="error" :empty="empty">
        <slot />
      </AwdStatusShell>
    </div>
  </section>
</template>
```

- [ ] **Step 4: Run layout tests to verify they pass**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/StudentAwdWorkspaceLayout.test.ts src/modules/awd/__tests__/AdminAwdWorkspaceLayout.test.ts src/modules/awd/__tests__/TeacherAwdWorkspaceLayout.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/frontend/src/modules/awd/components \
  code/frontend/src/modules/awd/layouts \
  code/frontend/src/modules/awd/__tests__/StudentAwdWorkspaceLayout.test.ts \
  code/frontend/src/modules/awd/__tests__/AdminAwdWorkspaceLayout.test.ts \
  code/frontend/src/modules/awd/__tests__/TeacherAwdWorkspaceLayout.test.ts
git commit -m "feat(awd): 搭建共享工作台骨架"
```

### Task 3: Add student adapter and 5 student AWD pages

**Files:**
- Create: `code/frontend/src/modules/awd/adapters/studentAwdPageAdapter.ts`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdOverviewView.vue`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdServicesView.vue`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdTargetsView.vue`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdAttacksView.vue`
- Create: `code/frontend/src/modules/awd/views/student/StudentAwdCollabView.vue`
- Test: `code/frontend/src/modules/awd/__tests__/studentAwdPageAdapter.test.ts`
- Test: `code/frontend/src/composables/__tests__/useContestAWDWorkspace.test.ts`

- [ ] **Step 1: Write the failing student adapter test**

```ts
it('maps workspace and scoreboard data into 5 page models', () => {
  const result = buildStudentAwdPageModel({
    contest,
    workspace,
    scoreboardRows,
    selectedPage: 'targets',
  })

  expect(result.hero.pageTitle).toBe('目标目录')
  expect(result.targets.rows[0].teamName).toBe('TeamA')
  expect(result.attacks.recent.length).toBeGreaterThan(0)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/studentAwdPageAdapter.test.ts`
Expected: FAIL with missing adapter

- [ ] **Step 3: Implement adapter and page views using existing workspace composable output**

```ts
export function buildStudentAwdPageModel(input: StudentAwdAdapterInput): StudentAwdPageModel {
  return {
    hero: { contestTitle: input.contest.title, pageTitle: pageLabelMap[input.selectedPage] },
    overview: { scoreboard: input.scoreboardRows, recentEvents: input.workspace?.recent_events ?? [] },
    services: mapServices(input.workspace?.services ?? []),
    targets: mapTargets(input.workspace?.targets ?? []),
    attacks: mapRecentAttacks(input.workspace?.recent_events ?? []),
    collab: buildCollabFallback(input.workspace, input.scoreboardRows),
  }
}
```

- [ ] **Step 4: Run focused student tests to verify they pass**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/studentAwdPageAdapter.test.ts src/composables/__tests__/useContestAWDWorkspace.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/frontend/src/modules/awd/adapters/studentAwdPageAdapter.ts \
  code/frontend/src/modules/awd/views/student \
  code/frontend/src/modules/awd/__tests__/studentAwdPageAdapter.test.ts \
  code/frontend/src/composables/__tests__/useContestAWDWorkspace.test.ts
git commit -m "feat(awd): 接入学生端五个页面"
```

### Task 4: Route AWD contests to new student workspace and retire embedded battlefield UI

**Files:**
- Modify: `code/frontend/src/router/index.ts`
- Modify: `code/frontend/src/views/contests/ContestDetail.vue`
- Test: `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
- Test: `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`

- [ ] **Step 1: Write failing route/detail tests**

```ts
it('redirects awd contests to the new overview route', async () => {
  expect(router.resolve({ name: 'ContestAwdOverview', params: { id: '7' } }).fullPath)
    .toBe('/contests/7/awd/overview')
})

it('does not render the legacy ContestAWDWorkspacePanel import anymore', async () => {
  const source = await fs.readFile(sourcePath, 'utf8')
  expect(source).not.toContain('ContestAWDWorkspacePanel')
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
Expected: FAIL because old embedded panel import and new routes do not exist

- [ ] **Step 3: Implement student routes and contest detail redirect**

```ts
{
  path: 'contests/:id/awd/overview',
  name: 'ContestAwdOverview',
  component: () => import('@/modules/awd/views/student/StudentAwdOverviewView.vue'),
}

watch(
  () => contest.value?.mode,
  (mode) => {
    if (mode === 'awd') {
      router.replace({ name: 'ContestAwdOverview', params: { id: contestId.value } })
    }
  },
)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd code/frontend && npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/frontend/src/router/index.ts \
  code/frontend/src/views/contests/ContestDetail.vue \
  code/frontend/src/views/contests/__tests__/ContestDetail.test.ts \
  code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts
git commit -m "feat(awd): 学生端路由接管新战场工作台"
```

### Task 5: Add admin adapter and 9 admin AWD pages

**Files:**
- Create: `code/frontend/src/modules/awd/adapters/adminAwdPageAdapter.ts`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdOverviewView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdReadinessView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdRoundsView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdServicesView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdAttacksView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdTrafficView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdAlertsView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdInstancesView.vue`
- Create: `code/frontend/src/modules/awd/views/admin/AdminAwdReplayView.vue`
- Test: `code/frontend/src/modules/awd/__tests__/adminAwdPageAdapter.test.ts`
- Test: `code/frontend/src/composables/__tests__/useAdminContestAWD.test.ts`

- [ ] **Step 1: Write the failing admin adapter test**

```ts
it('builds page models from readiness, round summary, traffic, attacks, and services', () => {
  const result = buildAdminAwdPageModel({
    contest,
    rounds,
    summary,
    readiness,
    services,
    attacks,
    trafficSummary,
    selectedPage: 'alerts',
  })

  expect(result.hero.pageTitle).toBe('告警中心')
  expect(result.alerts.items.length).toBeGreaterThan(0)
  expect(result.instances.rows.length).toBeGreaterThan(0)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/adminAwdPageAdapter.test.ts`
Expected: FAIL with missing adapter

- [ ] **Step 3: Implement admin adapter and page views, reusing current operation panels where possible**

```ts
export function buildAdminAwdPageModel(input: AdminAwdAdapterInput): AdminAwdPageModel {
  return {
    overview: buildOverview(input),
    readiness: buildReadiness(input.readiness),
    rounds: buildRounds(input.rounds, input.summary),
    services: buildServiceMatrix(input.services),
    attacks: buildAttackLog(input.attacks),
    traffic: buildTraffic(input.trafficSummary),
    alerts: buildAlertFallback(input.services, input.attacks, input.trafficSummary),
    instances: buildInstanceFallback(input.services),
    replay: buildReplayFallback(input.rounds, input.summary, input.attacks),
  }
}
```

- [ ] **Step 4: Run focused admin tests**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/adminAwdPageAdapter.test.ts src/composables/__tests__/useAdminContestAWD.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/frontend/src/modules/awd/adapters/adminAwdPageAdapter.ts \
  code/frontend/src/modules/awd/views/admin \
  code/frontend/src/modules/awd/__tests__/adminAwdPageAdapter.test.ts \
  code/frontend/src/composables/__tests__/useAdminContestAWD.test.ts
git commit -m "feat(awd): 接入管理员端九个页面"
```

### Task 6: Route admin AWD operations to new workspace and strip legacy edit-page ownership

**Files:**
- Modify: `code/frontend/src/router/index.ts`
- Modify: `code/frontend/src/views/platform/ContestEdit.vue`
- Modify: `code/frontend/src/views/platform/ContestOperationsHub.vue`
- Modify: `code/frontend/src/config/backofficeNavigation.ts`
- Test: `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
- Test: `code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts`
- Test: `code/frontend/src/config/__tests__/backofficeNavigation.test.ts`

- [ ] **Step 1: Write failing admin routing/navigation tests**

```ts
it('exposes admin awd workspace routes', () => {
  expect(router.resolve({ name: 'AdminAwdOverview', params: { id: '9' } }).fullPath)
    .toBe('/platform/contests/9/awd/overview')
})

it('uses new awd workspace path in backoffice navigation', () => {
  const modules = getVisibleBackofficeModules('admin')
  expect(JSON.stringify(modules)).toContain('/platform/contests')
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts src/views/platform/__tests__/ContestOperationsHub.test.ts src/config/__tests__/backofficeNavigation.test.ts`
Expected: FAIL because old ContestEdit AWD ownership and old nav paths remain

- [ ] **Step 3: Implement admin route handoff and remove legacy ownership**

```ts
if (contest.value?.mode === 'awd') {
  router.replace({ name: 'AdminAwdOverview', params: { id: contestId.value } })
  return
}
```

- [ ] **Step 4: Run admin route/navigation tests**

Run: `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts src/views/platform/__tests__/ContestOperationsHub.test.ts src/config/__tests__/backofficeNavigation.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/frontend/src/router/index.ts \
  code/frontend/src/views/platform/ContestEdit.vue \
  code/frontend/src/views/platform/ContestOperationsHub.vue \
  code/frontend/src/config/backofficeNavigation.ts \
  code/frontend/src/views/platform/__tests__/ContestEdit.test.ts \
  code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts \
  code/frontend/src/config/__tests__/backofficeNavigation.test.ts
git commit -m "feat(awd): 管理员端切换到新工作台入口"
```

### Task 7: Add teacher adapter and 5 teacher AWD pages

**Files:**
- Create: `code/frontend/src/modules/awd/adapters/teacherAwdPageAdapter.ts`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdOverviewView.vue`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdTeamsView.vue`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdServicesView.vue`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdReplayView.vue`
- Create: `code/frontend/src/modules/awd/views/teacher/TeacherAwdExportView.vue`
- Test: `code/frontend/src/modules/awd/__tests__/teacherAwdPageAdapter.test.ts`
- Test: `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`

- [ ] **Step 1: Write the failing teacher adapter test**

```ts
it('maps review archive data to overview, teams, services, replay, and export pages', () => {
  const result = buildTeacherAwdPageModel({
    review,
    selectedPage: 'services',
  })

  expect(result.hero.pageTitle).toBe('Service 复盘')
  expect(result.services.cards.length).toBeGreaterThan(0)
  expect(result.export.canExportReport).toBe(true)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/teacherAwdPageAdapter.test.ts`
Expected: FAIL with missing adapter

- [ ] **Step 3: Implement teacher adapter and page views**

```ts
export function buildTeacherAwdPageModel(input: TeacherAwdAdapterInput): TeacherAwdPageModel {
  return {
    overview: buildOverview(input.review),
    teams: buildTeams(input.review.selected_round),
    services: buildServicesFallback(input.review),
    replay: buildReplay(input.review),
    export: buildExport(input.review),
  }
}
```

- [ ] **Step 4: Run focused teacher tests**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/teacherAwdPageAdapter.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/frontend/src/modules/awd/adapters/teacherAwdPageAdapter.ts \
  code/frontend/src/modules/awd/views/teacher \
  code/frontend/src/modules/awd/__tests__/teacherAwdPageAdapter.test.ts \
  code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts
git commit -m "feat(awd): 接入教师端五个页面"
```

### Task 8: Route teacher/admin review entrypoints to new review workspace and retire monolithic detail page

**Files:**
- Modify: `code/frontend/src/router/index.ts`
- Modify: `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
- Modify: `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- Modify: `code/frontend/src/views/platform/AWDReviewIndex.vue`
- Modify: `code/frontend/src/config/backofficeNavigation.ts`
- Test: `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- Test: `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
- Test: `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`

- [ ] **Step 1: Write failing review routing tests**

```ts
it('routes teacher review detail to overview subpage', () => {
  expect(router.resolve({ name: 'TeacherAwdOverview', params: { contestId: '5' } }).fullPath)
    .toBe('/academy/awd-reviews/5/overview')
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
Expected: FAIL because old monolithic detail route still owns the UI

- [ ] **Step 3: Implement new teacher/admin review route handoff**

```ts
router.push({
  name: 'TeacherAwdOverview',
  params: { contestId: contest.id },
})
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/frontend/src/router/index.ts \
  code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue \
  code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue \
  code/frontend/src/views/platform/AWDReviewIndex.vue \
  code/frontend/src/config/backofficeNavigation.ts \
  code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts \
  code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts \
  code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts
git commit -m "feat(awd): 教师复盘入口切换到新多页工作台"
```

### Task 9: Remove obsolete AWD UI artifacts and verify all 19 routes

**Files:**
- Delete: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- Modify: `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- Create: `code/frontend/src/modules/awd/__tests__/routeCoverage.test.ts`
- Modify: `code/frontend/src/router/index.ts`

- [ ] **Step 1: Write a failing 19-route coverage test**

```ts
it('registers all 19 awd routes', () => {
  const awdRoutes = router.getRoutes().filter((route) => route.path.includes('/awd'))
  expect(awdRoutes).toHaveLength(19)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/routeCoverage.test.ts`
Expected: FAIL if any new page route is missing

- [ ] **Step 3: Delete obsolete AWD UI artifact and finish route cleanup**

```ts
// Remove ContestAWDWorkspacePanel import sites and any dead redirects
```

- [ ] **Step 4: Run route coverage and source cleanup tests**

Run: `cd code/frontend && npm run test:run -- src/modules/awd/__tests__/routeCoverage.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "refactor(awd): 清理旧战场主入口"
```

### Task 10: Final verification, typecheck, and docs sync

**Files:**
- Modify if needed: `code/frontend/src/modules/awd/**`
- Modify if needed: `docs/superpowers/specs/2026-04-20-awd-ui-pages-design.md`
- Test: existing focused test files touched above

- [ ] **Step 1: Run focused AWD test suite**

Run:

```bash
cd code/frontend
npm run test:run -- \
  src/modules/awd/__tests__/navigation.test.ts \
  src/modules/awd/__tests__/studentAwdPageAdapter.test.ts \
  src/modules/awd/__tests__/adminAwdPageAdapter.test.ts \
  src/modules/awd/__tests__/teacherAwdPageAdapter.test.ts \
  src/modules/awd/__tests__/StudentAwdWorkspaceLayout.test.ts \
  src/modules/awd/__tests__/AdminAwdWorkspaceLayout.test.ts \
  src/modules/awd/__tests__/TeacherAwdWorkspaceLayout.test.ts \
  src/modules/awd/__tests__/routeCoverage.test.ts \
  src/views/contests/__tests__/ContestDetail.test.ts \
  src/views/platform/__tests__/ContestEdit.test.ts \
  src/views/platform/__tests__/ContestOperationsHub.test.ts \
  src/views/platform/__tests__/AWDReviewIndex.test.ts \
  src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts \
  src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts
```

Expected: PASS

- [ ] **Step 2: Run typecheck**

Run: `cd code/frontend && npm run typecheck`
Expected: PASS with no type errors

- [ ] **Step 3: Run production build smoke check**

Run: `cd code/frontend && npm run build`
Expected: PASS with Vite build output and no fatal errors

- [ ] **Step 4: Reconcile docs if implementation diverged**

```md
- Update the spec only if route names, file placement, or cleanup scope changed during implementation.
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "test(awd): 完成工作台切换验证"
```

## Notes for Execution

- 先建骨架，再接学生，再接管理员，再接教师，最后清理旧 AWD 页面，不要反过来。
- adapter 层是唯一允许补 mock 的地方，页面组件只消费 view model。
- 真实接口页必须显式覆盖 loading / error / empty / success / disabled-by-status。
- 如果 `ContestEdit.vue` 的 AWD 配置段仍然承担必要配置职责，可以先通过新路由托管运行态页面，再在后续小提交里移除剩余旧 UI。
