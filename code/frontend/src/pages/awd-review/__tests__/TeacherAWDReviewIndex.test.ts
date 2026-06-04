import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'

import TeacherAWDReviewIndex from '@/pages/awd-review/TeacherAwdReviewIndexRoutePage.vue'
import appRouteLinkSource from '@/shared/ui/navigation/AppRouteLink.vue?raw'
import teacherAwdReviewIndexSource from '@/pages/awd-review/TeacherAwdReviewIndexRoutePage.vue?raw'
import teacherAwdReviewIndexWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue?raw'
import teacherAwdReviewContestDirectorySource from '@/widgets/awd-review-workspace/AwdReviewContestDirectory.vue?raw'
import teacherAwdReviewDirectorySectionSource from '@/widgets/awd-review-workspace/AwdReviewDirectorySection.vue?raw'
import awdReviewIndexPageSource from '@/features/awd-review-workspace/model/useAwdReviewIndex.ts?raw'
import awdReviewIndexRoutePageSource from '@/features/awd-review-workspace/model/useAwdReviewIndexPage.ts?raw'
import awdReviewIndexRoutesSource from '@/features/awd-review-workspace/model/awdReviewIndexRoutes.ts?raw'
import awdReviewDirectorySource from '@/features/awd-review-workspace/model/useAwdReviewDirectory.ts?raw'

const adminApiMocks = vi.hoisted(() => ({
  listPlatformAWDReviews: vi.fn(),
}))

const teacherApiMocks = vi.hoisted(() => ({
  listTeacherAWDReviews: vi.fn(),
}))

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/api/teacher', () => teacherApiMocks)

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/academy/reviews',
        name: 'TeacherAWDReviewIndex',
        component: TeacherAWDReviewIndex,
      },
      {
        path: '/academy/dashboard',
        name: 'TeacherDashboard',
        component: { template: '<div>teacher dashboard</div>' },
      },
      {
        path: '/academy/awd/reviews/:contestId',
        name: 'TeacherAWDReviewDetail',
        component: { template: '<div>teacher awd review detail</div>' },
      },
    ],
  })
}

async function mountPage() {
  const router = createTestRouter()
  await router.push('/academy/reviews')
  await router.isReady()

  const wrapper = mount(TeacherAWDReviewIndex, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
  return { wrapper, router }
}

describe('TeacherAWDReviewIndex', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.listTeacherAWDReviews.mockResolvedValue({
      list: [
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
      ],
      total: 2,
      page: 1,
      page_size: 20,
      summary: {
        running_count: 1,
        export_ready_count: 1,
      },
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应加载 AWD 赛事目录并渲染进入复盘入口', async () => {
    const { wrapper } = await mountPage()

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledWith(
      {
        status: undefined,
        keyword: undefined,
        page: 1,
        page_size: 20,
      },
      {
        signal: expect.any(AbortSignal),
      }
    )
    expect(adminApiMocks.listPlatformAWDReviews).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('春季 AWD 联训')
    expect(wrapper.text()).toContain('进入复盘')
  })

  it('页面应通过 feature model 获取筛选与摘要状态，不再直接耦合 teacher api', () => {
    expect(teacherAwdReviewIndexSource).toContain(
      "useAwdReviewIndexPage } from '@/features/awd-review-workspace'"
    )
    expect(awdReviewIndexPageSource).toContain("from './useAwdReviewDirectory'")
    expect(awdReviewIndexPageSource).not.toContain("from '@/api/awd-reviews'")
    expect(awdReviewDirectorySource).toContain("from '@/api/awd-reviews'")
    expect(awdReviewDirectorySource).not.toContain("from '@/api/admin'")
    expect(awdReviewDirectorySource).not.toContain("from '@/api/teacher'")
    expect(teacherAwdReviewIndexSource).toContain(
      "import { AwdReviewIndexWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(teacherAwdReviewIndexSource).not.toContain('TeacherAWDReviewIndexWorkspace')
    expect(teacherAwdReviewIndexSource).not.toContain("from '@/api/teacher'")
    expect(teacherAwdReviewIndexSource).not.toContain('const statusOptions = [')
    expect(teacherAwdReviewIndexSource).not.toContain('function contestStatusLabel')
    expect(awdReviewIndexPageSource).not.toContain("from 'vue-router'")
    expect(awdReviewIndexRoutePageSource).not.toContain("from 'vue-router'")
    expect(teacherAwdReviewIndexSource).not.toContain("useRouter } from 'vue-router'")
    expect(teacherAwdReviewIndexSource).not.toContain("router.push({ name: 'TeacherDashboard' })")
    expect(teacherAwdReviewIndexSource).toContain(':dashboard-route="homeRoute"')
    expect(teacherAwdReviewIndexSource).toContain(':build-contest-route="buildContestRoute"')
    expect(teacherAwdReviewIndexWorkspaceSource).toContain("from '@/shared/ui/navigation/AppRouteLink.vue'")
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('<AppRouteLink')
    expect(teacherAwdReviewContestDirectorySource).toContain(':build-contest-route="buildContestRoute"')
    expect(teacherAwdReviewContestDirectorySource).not.toContain("@open-contest=\"emit('openContest', $event)\"")
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
    expect(teacherAwdReviewIndexSource).not.toContain('contests.filter((item) => item.status ===')
    expect(teacherAwdReviewIndexSource).not.toContain('contests.filter((item) => item.export_ready)')
  })

  it('应在停止输入后自动筛选，不再依赖显式筛选按钮', async () => {
    const { wrapper } = await mountPage()

    const statusSelect = wrapper.find('select')
    const keywordInput = wrapper.find('input[placeholder="搜索赛事标题"]')

    expect(statusSelect.exists()).toBe(true)
    expect(keywordInput.exists()).toBe(true)
    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(1)

    await statusSelect.setValue('ended')
    await keywordInput.setValue('期末')

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalledTimes(2)
    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenLastCalledWith(
      {
        status: 'ended',
        keyword: '期末',
        page: 1,
        page_size: 20,
      },
      {
        signal: expect.any(AbortSignal),
      }
    )
    expect(wrapper.text()).not.toContain('应用筛选')
  })

  it('头部概览按钮应返回教学概览', async () => {
    const { wrapper, router } = await mountPage()

    const overviewButton = wrapper.get('a.header-btn--ghost')

    expect(overviewButton.text()).toContain('教学概览')

    await overviewButton.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('TeacherDashboard')
  })

  it('目录行应通过 route target 进入教师复盘详情', async () => {
    const { wrapper, router } = await mountPage()

    await wrapper.get('a.teacher-directory-row').trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('TeacherAWDReviewDetail')
    expect(router.currentRoute.value.params).toMatchObject({ contestId: 'contest-1' })
  })

  it('赛事概览条不应继续保留多余的底部分隔线', () => {
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('<AwdReviewSummaryPanel')
  })

  it('平台 AWD 复盘页头部应切到 workspace 语义，不再保留 teacher journal eyebrow', () => {
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('<AwdReviewWorkspaceHeader')
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('AWD_REVIEW_INDEX_WORKSPACE_COPY')
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain('TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY')
    expect(teacherAwdReviewIndexWorkspaceSource).toContain(
      ':overline="AWD_REVIEW_INDEX_WORKSPACE_COPY.overline"'
    )
    expect(teacherAwdReviewIndexWorkspaceSource).toContain(
      ':title="AWD_REVIEW_INDEX_WORKSPACE_COPY.title"'
    )
    expect(teacherAwdReviewIndexWorkspaceSource).toContain('buildAwdReviewIndexSummaryItems')
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain('buildTeacherAwdReviewIndexSummaryItems')
  })

  it('筛选区源码不应继续保留表单提交和应用筛选按钮', () => {
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain('@submit.prevent="loadContests"')
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain('应用筛选')
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain('赛事筛选')
    expect(teacherAwdReviewIndexWorkspaceSource).not.toContain(
      '支持按状态或关键字快速定位要进入的 AWD 赛事。'
    )
  })

  it('教师端路由入口应已迁到 widget route page，而不是旧 views 层', () => {
    expect(teacherAwdReviewIndexSource).not.toContain('@/views/teacher/TeacherAWDReviewIndex.vue')
  })
})
