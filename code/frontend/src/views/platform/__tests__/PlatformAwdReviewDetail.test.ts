import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'
import platformAwdReviewDetailSource from '../PlatformAwdReviewDetail.vue?raw'
import TeacherAWDReviewWorkspace from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeMock = {
  params: {
    contestId: 'contest-1',
  },
  query: {} as Record<string, string>,
}

const teachingApiMocks = vi.hoisted(() => ({
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

vi.mock('@/api/teaching', () => teachingApiMocks)

describe('PlatformAwdReviewDetail route owner', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    pushMock.mockReset()
    replaceMock.mockReset()
    routeMock.params.contestId = 'contest-1'
    routeMock.query = {}
    Object.values(teachingApiMocks).forEach((mock) => mock.mockReset())

    teachingApiMocks.getTeacherAWDReview.mockResolvedValue({
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
    teachingApiMocks.exportTeacherAWDReviewArchive.mockResolvedValue({
      report_id: '31',
      status: 'processing',
    })
    teachingApiMocks.exportTeacherAWDReviewReport.mockResolvedValue({
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
    expect(platformAwdReviewDetailSource).not.toContain("from '@/views/teacher/TeacherAWDReviewDetail.vue'")
    expect(platformAwdReviewDetailSource).not.toContain("from '@/api/teacher'")

    const PlatformAwdReviewDetail = (await import('../PlatformAwdReviewDetail.vue')).default
    const wrapper = mount(PlatformAwdReviewDetail)

    await flushPromises()

    expect(teachingApiMocks.getTeacherAWDReview).toHaveBeenCalledWith('contest-1', {
      round: undefined,
      team_id: undefined,
    })
    expect(wrapper.findComponent(TeacherAWDReviewWorkspace).exists()).toBe(true)
    expect(wrapper.text()).toContain('春季 AWD 联训')

    const backButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('返回列表'))

    expect(backButton).toBeTruthy()

    await backButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformAwdReviewIndex' })
  })
})
