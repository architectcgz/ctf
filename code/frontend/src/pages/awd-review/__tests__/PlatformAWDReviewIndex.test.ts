import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'

import PlatformAWDReviewIndex from '@/pages/awd-review/PlatformAwdReviewIndexRoutePage.vue'
import appRouteLinkSource from '@/components/navigation/AppRouteLink.vue?raw'
import platformAwdReviewIndexSource from '@/pages/awd-review/PlatformAwdReviewIndexRoutePage.vue?raw'
import awdReviewIndexPageSource from '@/features/awd-review-workspace/model/useAwdReviewIndex.ts?raw'
import awdReviewIndexRoutePageSource from '@/features/awd-review-workspace/model/useAwdReviewIndexPage.ts?raw'
import awdReviewIndexRoutesSource from '@/features/awd-review-workspace/model/awdReviewIndexRoutes.ts?raw'
import awdReviewHeroPanelSource from '@/widgets/awd-review-workspace/AwdReviewHeroPanel.vue?raw'
import awdReviewDirectoryPanelSource from '@/widgets/awd-review-workspace/AwdReviewDirectoryPanel.vue?raw'
import { useAuthStore } from '@/stores/auth'

const adminApiMocks = vi.hoisted(() => ({
  listPlatformAWDReviews: vi.fn(),
}))

const teacherApiMocks = vi.hoisted(() => ({
  listTeacherAWDReviews: vi.fn(),
}))

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/api/teacher', () => teacherApiMocks)

const combinedSource = [
  platformAwdReviewIndexSource,
  awdReviewHeroPanelSource,
  awdReviewDirectoryPanelSource,
].join('\n')

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/platform/reviews',
        name: 'PlatformAwdReviewIndex',
        component: PlatformAWDReviewIndex,
      },
      {
        path: '/platform/overview',
        name: 'PlatformOverview',
        component: { template: '<div>platform overview</div>' },
      },
      {
        path: '/platform/awd/reviews/:contestId',
        name: 'PlatformAwdReviewDetail',
        component: { template: '<div>platform awd review detail</div>' },
      },
    ],
  })
}

async function mountPage() {
  const router = createTestRouter()
  await router.push('/platform/reviews')
  await router.isReady()

  const wrapper = mount(PlatformAWDReviewIndex, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
  return { wrapper, router }
}

describe('PlatformAWDReviewIndex', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    useAuthStore().user = { id: 'admin-1', role: 'admin' } as never
    adminApiMocks.listPlatformAWDReviews.mockReset()
    teacherApiMocks.listTeacherAWDReviews.mockReset()
    adminApiMocks.listPlatformAWDReviews.mockImplementation(async (params) => {
      const contests = [
        {
          id: 'contest-1',
          title: '春季 AWD 联训',
          mode: 'awd',
          status: 'running',
          current_round: 2,
          round_count: 6,
          team_count: 8,
          export_ready: false,
        },
        {
          id: 'contest-2',
          title: '期末 AWD 复盘',
          mode: 'awd',
          status: 'ended',
          current_round: 8,
          round_count: 8,
          team_count: 10,
          export_ready: true,
        },
      ]

      const filtered = contests.filter((item) => {
        const matchesStatus = !params?.status || item.status === params.status
        const matchesKeyword = !params?.keyword || item.title.includes(params.keyword)
        return matchesStatus && matchesKeyword
      })

      return {
        list: filtered,
        total: filtered.length,
        page: params?.page ?? 1,
        page_size: params?.page_size ?? 20,
        summary: {
          running_count: filtered.filter((item) => item.status === 'running').length,
          export_ready_count: filtered.filter((item) => item.export_ready).length,
        },
      }
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应使用平台工作台目录壳层而不是教师目录模板', async () => {
    expect(combinedSource).toContain("from '@/components/common/WorkspaceDirectoryToolbar.vue'")
    expect(combinedSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(platformAwdReviewIndexSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero admin-awd-review-shell flex min-h-full flex-1 flex-col"'
    )
    expect(awdReviewHeroPanelSource).toContain(
      'class="admin-summary-grid admin-awd-review-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(combinedSource).toContain(
      'class="workspace-directory-section admin-awd-review-directory"'
    )
    expect(combinedSource).toContain('class="workspace-directory-list admin-awd-review-table"')
    expect(combinedSource).not.toContain('teacher-management-shell')
    expect(combinedSource).not.toContain('teacher-directory-row')

    const { wrapper } = await mountPage()

    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('春季 AWD 联训')
    expect(wrapper.text()).toContain('赛事目录')
    expect(wrapper.text()).toContain('进入复盘')
  })

  it('应支持自动筛选并跳转到平台复盘详情', async () => {
    const { wrapper, router } = await mountPage()

    const searchInput = wrapper.get('input[placeholder="搜索赛事标题"]')
    await searchInput.setValue('期末')

    expect(adminApiMocks.listPlatformAWDReviews).toHaveBeenCalledTimes(1)
    expect(teacherApiMocks.listTeacherAWDReviews).not.toHaveBeenCalled()

    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(adminApiMocks.listPlatformAWDReviews).toHaveBeenCalledTimes(2)
    expect(adminApiMocks.listPlatformAWDReviews).toHaveBeenLastCalledWith(
      {
        status: undefined,
        keyword: '期末',
        page: 1,
        page_size: 20,
      },
      {
        signal: expect.any(AbortSignal),
      }
    )

    await wrapper
      .findAll('a')
      .find((node) => node.text().includes('进入复盘'))
      ?.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('PlatformAwdReviewDetail')
    expect(router.currentRoute.value.params).toMatchObject({ contestId: 'contest-2' })
  })

  it('顶部返回按钮应回到平台概览', async () => {
    const { wrapper, router } = await mountPage()

    const overviewButton = wrapper.get('a.header-btn--ghost')
    expect(overviewButton.text()).toContain('返回平台概览')

    await overviewButton.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('PlatformOverview')
  })

  it('路由壳页应通过 route target 处理返回和详情跳转', () => {
    expect(platformAwdReviewIndexSource).toContain(
      "useAwdReviewIndexPage } from '@/features/awd-review-workspace'"
    )
    expect(awdReviewIndexPageSource).not.toContain("from 'vue-router'")
    expect(awdReviewIndexRoutePageSource).not.toContain("from 'vue-router'")
    expect(platformAwdReviewIndexSource).not.toContain("useRouter } from 'vue-router'")
    expect(platformAwdReviewIndexSource).not.toContain("router.push({ name: 'PlatformOverview' })")
    expect(platformAwdReviewIndexSource).toContain(':overview-route="homeRoute"')
    expect(platformAwdReviewIndexSource).toContain(':build-contest-route="buildContestRoute"')
    expect(awdReviewHeroPanelSource).toContain("from '@/components/navigation/AppRouteLink.vue'")
    expect(awdReviewHeroPanelSource).toContain('<AppRouteLink')
    expect(awdReviewDirectoryPanelSource).toContain("from '@/components/navigation/AppRouteLink.vue'")
    expect(awdReviewDirectoryPanelSource).toContain('<AppRouteLink')
    expect(awdReviewIndexRoutePageSource).toContain('homeRoute: resolveAwdReviewIndexHomeRoute(scope)')
    expect(awdReviewIndexRoutePageSource).toContain(
      'buildContestRoute: (contestId: string) => buildAwdReviewDetailRoute(scope, contestId)'
    )
    expect(awdReviewIndexRoutesSource).toContain(
      "name: scope === 'teacher' ? 'TeacherDashboard' : 'PlatformOverview'"
    )
    expect(awdReviewIndexRoutesSource).toContain(
      "name: scope === 'teacher' ? 'TeacherAWDReviewDetail' : 'PlatformAwdReviewDetail'"
    )
    expect(appRouteLinkSource).toContain('<RouterLink')
  })

  it('路由入口应已迁到 widget route page，而不是旧 views 层', () => {
    expect(platformAwdReviewIndexSource).not.toContain('@/views/platform/AWDReviewIndex.vue')
  })
})
