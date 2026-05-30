import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'
import platformAwdReviewDetailSource from '@/pages/awd-review/PlatformAwdReviewDetailRoutePage.vue?raw'
import AwdReviewWorkspace from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue'
import awdReviewDetailRoutesSource from '@/features/awd-review-detail-workspace/model/awdReviewDetailRoutes.ts?raw'
import awdReviewDetailPageSource from '@/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts?raw'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeMock = {
  params: {
    contestId: 'contest-1',
  },
  query: {} as Record<string, string>,
}

const adminApiMocks = vi.hoisted(() => ({
  getPlatformAWDReview: vi.fn(),
  exportPlatformAWDReviewArchive: vi.fn(),
  exportPlatformAWDReviewReport: vi.fn(),
}))

const teacherApiMocks = vi.hoisted(() => ({
  getTeacherAWDReview: vi.fn(),
  exportTeacherAWDReviewArchive: vi.fn(),
  exportTeacherAWDReviewReport: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeMock,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
  }
})

vi.mock('@/api/admin', () => adminApiMocks)
vi.mock('@/api/teacher', () => teacherApiMocks)

describe('PlatformAwdReviewDetail route owner', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    pushMock.mockReset()
    replaceMock.mockReset()
    routeMock.params.contestId = 'contest-1'
    routeMock.query = {}
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    adminApiMocks.getPlatformAWDReview.mockResolvedValue({
      generated_at: '2026-04-12T10:00:00Z',
      scope: {
        snapshot_type: 'live',
        requested_by: 5,
        requested_id: 'contest-1',
      },
      contest: {
        id: 'contest-1',
        title: '春季 AWD 联训',
        mode: 'awd',
        status: 'running',
        current_round: 2,
        round_count: 4,
        team_count: 6,
        export_ready: false,
      },
      overview: {
        round_count: 4,
        team_count: 6,
        service_count: 12,
        attack_count: 8,
        traffic_count: 20,
      },
      rounds: [],
      selected_round: null,
    })
    adminApiMocks.exportPlatformAWDReviewArchive.mockResolvedValue({
      report_id: '31',
      status: 'processing',
    })
    adminApiMocks.exportPlatformAWDReviewReport.mockResolvedValue({
      report_id: '32',
      status: 'processing',
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      class_name: '',
    })
  })

  it('应使用平台 route view，并通过中性 feature 承接 AWD 复盘详情 workflow', async () => {
    expect(platformAwdReviewDetailSource).toContain(
      "import { useAwdReviewDetailPage } from '@/features/awd-review-detail-workspace'"
    )
    expect(awdReviewDetailPageSource).toContain("from '@/api/awd-reviews'")
    expect(awdReviewDetailPageSource).toContain(
      "import { useRouteQueryTransport } from '@/composables/routeQueryTransport'"
    )
    expect(awdReviewDetailPageSource).toContain(
      "import { useRouteNavigationTransport } from '@/composables/routeNavigationTransport'"
    )
    expect(awdReviewDetailPageSource).toContain("from './awdReviewDetailRoutes'")
    expect(awdReviewDetailPageSource).not.toContain("from '@/api/admin'")
    expect(awdReviewDetailPageSource).not.toContain("from '@/api/teacher'")
    expect(awdReviewDetailPageSource).not.toContain("from 'vue-router'")
    expect(awdReviewDetailPageSource).not.toContain('route,')
    expect(awdReviewDetailRoutesSource).toContain('resolveAwdReviewIndexRouteName')
    expect(platformAwdReviewDetailSource).not.toContain("from '@/views/teacher/TeacherAWDReviewDetail.vue'")
    expect(platformAwdReviewDetailSource).not.toContain("from '@/api/teacher'")

    const PlatformAwdReviewDetail = (await import('../PlatformAwdReviewDetail.vue')).default
    const wrapper = mount(PlatformAwdReviewDetail)

    await flushPromises()

    expect(adminApiMocks.getPlatformAWDReview).toHaveBeenCalledWith('contest-1', {
      round: undefined,
      team_id: undefined,
    })
    expect(teacherApiMocks.getTeacherAWDReview).not.toHaveBeenCalled()
    expect(wrapper.findComponent(AwdReviewWorkspace).exists()).toBe(true)
    expect(wrapper.text()).toContain('春季 AWD 联训')

    const backButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('返回列表'))

    expect(backButton).toBeTruthy()

    await backButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformAwdReviewIndex' })
  })
})
